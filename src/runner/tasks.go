package runner

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/michalskocz/velkor/src/configuration"
	"github.com/michalskocz/velkor/src/logger"
)

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

func runTask(ctx context.Context, task configuration.Task) error {
	if err := ctx.Err(); err != nil {
		return nil
	}

	taskLog := logger.Log.With().
		Str("task", task.Name).
		Logger()

	taskLog.Info().
		Str("container", task.Image.Name).
		Msg("Running task")

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
	runArgs, volume, err := buildRunArgs(task, pwd)
	if err != nil {
		return err
	}

	taskLog.Debug().
		Str("engine", engine).
		Str("cmd", strings.Join(runArgs, " ")).
		Msg("Starting container")

	defer stopTaskAndRM(task, engine, volume)

	if err := runCommand(ctx, engine, runArgs, env, logFile); err != nil {
		taskLog.Error().
			Err(err).
			Msg("Failed to start container")

		return fmt.Errorf("task '%s' failed: %w", task.Name, err)
	}

	if err := runScript(ctx, task, engine, env, logFile); err != nil {
		return err
	}

	if err := copyArtifacts(ctx, task, engine, env, logFile); err != nil {
		return err
	}

	taskLog.Info().
		Msg("Task completed")

	return nil
}
