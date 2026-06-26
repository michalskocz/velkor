package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"velkor/configuration"
	"velkor/internal"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func RunPipeline() error {
	log.Println(colorCyan + "Starting CI/CD pipeline..." + colorReset)

	for cfg.Stages.Next() {
		stage, err := fetchStage()
		if err != nil {
			return err
		}

		log.Printf(colorCyan+"--- Running Stage: %s ---"+colorReset+"\n", stage.Name)

		if err := runStage(stage); err != nil {
			return fmt.Errorf(colorRed+"pipeline aborted at stage '%s': %w"+colorReset, stage.Name, err)
		}
	}

	log.Println(colorGreen + "Success! All stages completed successfully." + colorReset)
	return nil
}

func fetchStage() (configuration.Stage, error) {
	rawStage, err := cfg.Stages.Get()
	if err != nil {
		return configuration.Stage{}, fmt.Errorf("error getting stage: %w", err)
	}

	stage, ok := rawStage.(configuration.Stage)
	if !ok {
		return configuration.Stage{}, fmt.Errorf("invalid Stage object type")
	}

	return stage, nil
}

func runStage(stage configuration.Stage) error {
	if stage.Tasks == nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workers := resolveWorkers()

	var wg sync.WaitGroup
	errChan := make(chan error, workers)
	sem := make(chan struct{}, workers)

	if err := os.MkdirAll(internal.LOG_DIR, 0755); err != nil {
		return fmt.Errorf("cannot create %s dir", internal.LOG_DIR)
	}

	for stage.Tasks.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}

		task, err := fetchTask(stage)
		if err != nil {
			cancel()
			return err
		}

		sem <- struct{}{}
		wg.Add(1)

		go func(t configuration.Task) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := runTask(ctx, t); err != nil {
				errChan <- err
				cancel()
			}
		}(task)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

func fetchTask(stage configuration.Stage) (configuration.Task, error) {
	rawTask, err := stage.Tasks.Get()
	if err != nil {
		return configuration.Task{}, err
	}

	task, ok := rawTask.(configuration.Task)
	if !ok {
		return configuration.Task{}, fmt.Errorf("invalid Task object type")
	}

	return task, nil
}

func resolveWorkers() int {
	if cfg.Threads <= 0 {
		return configuration.DEFAULT_THREADS
	}
	return cfg.Threads
}

func runTask(ctx context.Context, task configuration.Task) error {
	if err := ctx.Err(); err != nil {
		return nil
	}

	log.Printf(colorYellow+"[Task: %s]"+colorReset+" Running in container: %s\n", task.Name, task.Image.Name)

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	logFile, err := createLogFile(task)
	if err != nil {
		return err
	}
	defer logFile.Close()

	env := buildEnv(task)
	engine := resolveEngine(task)
	runArgs := buildRunArgs(task, pwd)

	if debug == internal.DEBUG_ON {
		log.Printf(
			colorCyan+"[DEBUG] task=%s engine=%s args=%s"+colorReset+"\n",
			task.Name,
			engine,
			strings.Join(runArgs, " "),
		)
	}

	defer stopTaskAndRM(ctx, task, engine, env, logFile)

	if err := runCommand(ctx, engine, runArgs, env, logFile); err != nil {

		return fmt.Errorf("task '%s' failed %v", task.Name, err)
	}

	if err := runScript(ctx, task, engine, env, logFile); err != nil {
		return err
	}

	if err := copyArtifacts(ctx, task, engine, env, logFile); err != nil {
		return err
	}

	log.Printf(colorGreen+"[Task: %s] Completed successfully."+colorReset+"\n", task.Name)
	return nil
}

func runCommand(ctx context.Context, engine string, args []string, env []string, logFile *os.File) error {
	cmd := exec.CommandContext(ctx, engine, args...)
	cmd.Env = env
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	return cmd.Run()
}

func runScript(ctx context.Context, task configuration.Task, engine string, env []string, logFile *os.File) error {
	for i, line := range task.Script {
		if err := ctx.Err(); err != nil {
			return nil
		}

		args := buildExecArgs(task, line)

		if debug == internal.DEBUG_ON {
			log.Printf(
				colorCyan+"[DEBUG] task=%s script=%d cmd=%s"+colorReset+"\n",
				task.Name,
				i,
				strings.Join(args, " "),
			)
		}

		cmd := exec.CommandContext(ctx, engine, args...)
		cmd.Env = env
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("task '%s' failed at line '%d:%s': %w", task.Name, i, line, err)
		}
	}
	return nil
}

func copyArtifacts(ctx context.Context, task configuration.Task, engine string, env []string, logFile *os.File) error {
	for _, a := range task.Artifacts {
		dest := filepath.Join("artifacts", task.Name, filepath.Dir(a))

		if err := os.MkdirAll(dest, 0755); err != nil {
			return fmt.Errorf("failed to create artifact directory '%s': %w", dest, err)
		}

		args := buildCopyArgs(task, a)

		if debug == internal.DEBUG_ON {
			log.Printf(
				colorCyan+"[DEBUG] task=%s artifact=%s cmd=%s"+colorReset+"\n",
				task.Name,
				a,
				strings.Join(args, " "),
			)
		}

		cmd := exec.CommandContext(ctx, engine, args...)
		cmd.Env = env
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("task '%s' failed at artifact [%s]: %w", task.Name, a, err)
		}
	}
	return nil
}

func stopTaskAndRM(ctx context.Context, task configuration.Task, engine string, env []string, logFile *os.File) error {
	cmd := exec.CommandContext(ctx, engine, "stop", task.Name)
	cmd.Env = env
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("task '%s' failed: %w", task.Name, err)
	}

	rm := exec.CommandContext(ctx, engine, "rm", "-f", task.Name)
	rm.Env = env
	rm.Stdout = logFile
	rm.Stderr = logFile

	if err := rm.Run(); err != nil {
		return fmt.Errorf("task '%s' failed: %w", task.Name, err)
	}

	return nil
}

func createLogFile(task configuration.Task) (*os.File, error) {
	image := strings.ReplaceAll(task.Image.Name, "/", "_")
	image = strings.ReplaceAll(image, ":", "_")

	name := fmt.Sprintf("%s_%s.log", task.Name, image)
	path := filepath.Join("log", name)

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return file, nil
}

func buildEnv(task configuration.Task) []string {
	env := os.Environ()

	for _, v := range cfg.GlobalVariables {
		env = append(env, fmt.Sprintf("%s=\"%s\"", v.Name, v.Value))
	}

	for _, v := range task.Variables {
		env = append(env, fmt.Sprintf("%s=\"%s\"", v.Name, v.Value))
	}

	return env
}

func resolveEngine(task configuration.Task) string {
	if task.Image.ContainerType == configuration.PODMAN {
		return "podman"
	}
	return "docker"
}

func buildRunArgs(task configuration.Task, pwd string) []string {
	args := []string{
		"run", "-dit",
		"--name", task.Name,
		"-v", fmt.Sprintf("%s:/workspace:O", pwd),
		"-w", "/workspace",
	}

	for _, v := range cfg.GlobalVariables {
		args = append(args, "-e", fmt.Sprintf("%s=%s", v.Name, v.Value))
	}

	for _, v := range task.Variables {
		args = append(args, "-e", fmt.Sprintf("%s=%s", v.Name, v.Value))
	}

	args = append(args, task.Image.Name)

	return args
}

func buildExecArgs(task configuration.Task, line string) []string {
	return []string{"exec", task.Name, "sh", "-c", line}
}

func buildCopyArgs(task configuration.Task, artifact string) []string {
	return []string{
		"cp",
		fmt.Sprintf("%s:/workspace/%s", task.Name, artifact),
		fmt.Sprintf("artifacts/%s/%s", task.Name, artifact),
	}
}
