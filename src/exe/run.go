package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"velkor/configuration"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func RunPipeline() error {
	fmt.Println(colorCyan + "Starting CI/CD pipeline..." + colorReset)

	for cfg.Stages.Next() {
		rawStage, err := cfg.Stages.Get()
		if err != nil {
			return fmt.Errorf("error getting stage: %w", err)
		}

		stage, ok := rawStage.(configuration.Stage)
		if !ok {
			return fmt.Errorf("invalid Stage object type")
		}

		fmt.Printf("\n"+colorCyan+"--- Running Stage: %s ---"+colorReset+"\n", stage.Name)

		err = runStageTasks(stage)
		if err != nil {
			return fmt.Errorf(colorRed+"pipeline aborted at stage '%s': %w"+colorReset, stage.Name, err)
		}
	}

	fmt.Println("\n" + colorGreen + "Success! All stages completed successfully." + colorReset)
	return nil
}

func runStageTasks(stage configuration.Stage) error {
	if stage.Tasks == nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	numWorkers := cfg.Threads
	if numWorkers <= 0 {
		numWorkers = configuration.DEFAULT_THREADS
	}

	var wg sync.WaitGroup
	errChan := make(chan error, numWorkers)

	sem := make(chan struct{}, numWorkers)

	for stage.Tasks.Next() {

		select {
		case <-ctx.Done():
			break
		default:
		}

		rawTask, err := stage.Tasks.Get()
		if err != nil {
			cancel()
			return err
		}

		task, ok := rawTask.(configuration.Task)
		if !ok {
			cancel()
			return fmt.Errorf("invalid Task object type")
		}

		sem <- struct{}{}
		wg.Add(1)

		go func(t configuration.Task) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := executeTask(ctx, t); err != nil {
				errChan <- err
				cancel()
			}
		}(task)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}

	return nil
}

func executeTask(ctx context.Context, task configuration.Task) error {

	if err := ctx.Err(); err != nil {
		return nil
	}

	fmt.Printf(colorYellow+"[Task: %s]"+colorReset+" Running in container: %s\n", task.Name, task.Image.Name)

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	env := os.Environ()
	for _, v := range cfg.GlobalVariables {
		env = append(env, fmt.Sprintf("%s=%s", v.Name, v.Value))
	}
	for _, v := range task.Variables {
		env = append(env, fmt.Sprintf("%s=%s", v.Name, v.Value))
	}

	for _, scriptLine := range task.Script {
		if err := ctx.Err(); err != nil {
			return nil
		}

		engine := "docker"
		if task.Image.ContainerTye == configuration.PODMAN {
			engine = "podman"
		}

		args := []string{"run", "--rm"}
		args = append(args, "-v", fmt.Sprintf("%s:/workspace:O", pwd))
		args = append(args, "--workdir", "/workspace")

		for _, v := range cfg.GlobalVariables {
			args = append(args, "-e", fmt.Sprintf("%s=%s", v.Name, v.Value))
		}

		for _, v := range task.Variables {
			args = append(args, "-e", fmt.Sprintf("%s=%s", v.Name, v.Value))
		}
		args = append(args, task.Image.Name, "sh", "-c", scriptLine)

		cmd := exec.CommandContext(ctx, engine, args...)
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("task '%s' failed at line '%s': %w", task.Name, scriptLine, err)
		}
	}

	fmt.Printf(colorGreen+"[Task: %s] Completed successfully."+colorReset+"\n", task.Name)
	return nil
}
