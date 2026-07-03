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

	"github.com/michalskocz/velkor/src/internal"

	"github.com/michalskocz/velkor/src/configuration"
)

func copyArtifacts(ctx context.Context, task configuration.Task, engine string, env []string, logFile *os.File) error {
	for _, a := range task.Artifacts {
		dest := filepath.Join("artifacts", task.Name, filepath.Dir(a))

		if err := os.MkdirAll(dest, 0755); err != nil {
			return fmt.Errorf("failed to create artifact directory '%s': %w", dest, err)
		}

		args := buildCopyArgs(task, a)

		if debug == internal.DEBUG_ON {
			log.Printf(
				internal.ColorCyan+"[DEBUG]"+internal.ColorReset+" task=%s artifact=%s cmd=%s"+internal.ColorReset+"\n",
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
