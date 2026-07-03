/*
 Copyright (c) 2026 Michał Skoczylas

 Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

 THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/

package runner

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/michalskocz/velkor/src/configuration"
	"github.com/michalskocz/velkor/src/internal"

	"github.com/google/uuid"
)

var cfg configuration.Config
var debug int
var goodIsolation bool
var cpu int

func RunPipeline(Config configuration.Config, Isolation bool, Cpu int, Debug int) error {
	cfg = Config
	goodIsolation = Isolation
	cpu = Cpu
	debug = Debug

	log.Println(internal.ColorCyan + "Starting CI/CD pipeline..." + internal.ColorReset)

	for cfg.Stages.Next() {
		stage, err := fetchStage()
		if err != nil {
			return err
		}

		log.Printf(internal.ColorCyan+"--- Running Stage: %s ---"+internal.ColorReset+"\n", stage.Name)

		if err := runStage(stage); err != nil {
			return fmt.Errorf(internal.ColorRed+"pipeline aborted at stage '%s': %w"+internal.ColorReset, stage.Name, err)
		}
	}

	log.Println(internal.ColorGreen + "Success! All stages completed successfully." + internal.ColorReset)
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
				internal.ColorCyan+"[DEBUG]"+internal.ColorReset+" task=%s script=%d cmd=%s"+internal.ColorReset+"\n",
				task.Name,
				i+1,
				strings.Join(args, " "),
			)
		}

		cmd := exec.CommandContext(ctx, engine, args...)
		cmd.Env = env
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		if err := cmd.Run(); err != nil {
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
		"--opt", fmt.Sprintf("o=lowerdir=%s,upperdir=%s,workdir=%s",
			pwd, upper, work),
		volume,
	}

	cmd := exec.Command(
		"docker", args...,
	)

	if debug == internal.DEBUG_ON {
		log.Printf(
			internal.ColorCyan+"[DEBUG]"+internal.ColorReset+" cmd=docker args=%s"+internal.ColorReset+"\n",
			strings.Join(args, " "),
		)
	}

	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%v: %s", err, out)
	}

	return volume, nil
}
