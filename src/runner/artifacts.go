package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/michalskocz/velkor/src/configuration"
	"github.com/michalskocz/velkor/src/logger"
)

func copyArtifacts(ctx context.Context, task configuration.Task, engine string, env []string, logFile *os.File) error {
	taskLog := logger.Log.With().
		Str("task", task.Name).
		Logger()

	for _, artifact := range task.Artifacts {
		dest := filepath.Join("artifacts", task.Name, filepath.Dir(artifact))

		if err := os.MkdirAll(dest, 0755); err != nil {
			return fmt.Errorf("failed to create artifact directory '%s': %w", dest, err)
		}

		args := buildCopyArgs(task, artifact)

		taskLog.Debug().
			Str("artifact", artifact).
			Str("cmd", strings.Join(args, " ")).
			Msg("Copying artifact")

		cmd := exec.CommandContext(ctx, engine, args...)
		cmd.Env = env
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		if err := cmd.Run(); err != nil {
			taskLog.Error().
				Err(err).
				Str("artifact", artifact).
				Msg("Failed to copy artifact")

			return fmt.Errorf("task '%s' failed at artifact '%s': %w", task.Name, artifact, err)
		}
	}

	return nil
}
