package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/michalskocz/velkor/src/configuration"
	"github.com/michalskocz/velkor/src/internal"
	"github.com/michalskocz/velkor/src/logger"
)

var cfg *configuration.Config
var debug int
var goodIsolation bool
var cpu int

func RunPipeline(local configuration.Local) error {
	cfg = local.Cfg
	goodIsolation = local.GoodIsolation
	cpu = local.Cpu
	debug = local.Debug

	logger.Init(debug == internal.DEBUG_ON)

	logger.Log.Info().
		Msg("Starting CI/CD pipeline")

	for cfg.Stages.Next() {
		stage, err := fetchStage()
		if err != nil {
			return err
		}

		logger.Log.Info().
			Str("stage", stage.Name).
			Msg("Running stage")

		if err := runStage(stage); err != nil {
			logger.Log.Error().
				Err(err).
				Str("stage", stage.Name).
				Msg("Pipeline aborted")

			return fmt.Errorf("pipeline aborted at stage '%s': %w", stage.Name, err)
		}
	}

	logger.Log.Info().
		Msg("Pipeline completed successfully")

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
	taskLog := logger.Log.With().
		Str("task", task.Name).
		Logger()

	for i, line := range task.Script {
		if err := ctx.Err(); err != nil {
			return nil
		}

		args := buildExecArgs(task, line)

		taskLog.Debug().
			Int("script", i+1).
			Str("cmd", strings.Join(args, " ")).
			Msg("Executing script")

		cmd := exec.CommandContext(ctx, engine, args...)
		cmd.Env = env
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		if err := cmd.Run(); err != nil {
			taskLog.Error().
				Err(err).
				Int("script", i+1).
				Msg("Script execution failed")

			return fmt.Errorf("task '%s' failed at line '%d:%s': %w", task.Name, i+1, line, err)
		}
	}

	return nil
}

func stopTaskAndRM(task configuration.Task, engine string, volume string) {
	_ = exec.Command(engine, "rm", "-f", "-t", "0", task.Name).Run()

	if task.Image.ContainerType == configuration.DOCKER {
		_ = exec.Command("docker", "volume", "rm", "-f", volume).Run()
	}
}

func resolveEngine(task configuration.Task) string {
	if task.Image.ContainerType == configuration.PODMAN {
		return "podman"
	}
	return "docker"
}

func createOverlayVolume(pwd string) (string, error) {
	upper := filepath.Join(os.TempDir(), uuid.NewString(), "upper")
	work := filepath.Join(os.TempDir(), uuid.NewString(), "work")

	if err := os.MkdirAll(upper, 0755); err != nil {
		return "", err
	}

	if err := os.MkdirAll(work, 0755); err != nil {
		return "", err
	}

	volume := "overlay-" + uuid.NewString()

	args := []string{
		"volume", "create",
		"--driver", "local",
		"--opt", "type=overlay",
		"--opt", "device=overlay",
		"--opt", fmt.Sprintf(
			"o=lowerdir=%s,upperdir=%s,workdir=%s",
			pwd,
			upper,
			work,
		),
		volume,
	}

	logger.Log.Debug().
		Str("engine", "docker").
		Str("cmd", strings.Join(args, " ")).
		Msg("Creating overlay volume")

	cmd := exec.Command("docker", args...)

	if out, err := cmd.CombinedOutput(); err != nil {
		logger.Log.Error().
			Err(err).
			Bytes("output", out).
			Msg("Failed to create overlay volume")

		return "", fmt.Errorf("%v: %s", err, out)
	}

	return volume, nil
}
