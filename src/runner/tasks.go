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
	"strings"
	"velkor/configuration"
	"velkor/internal"
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
	runArgs, volume, err := buildRunArgs(task, pwd)

	if err != nil {
		return err
	}

	if debug == internal.DEBUG_ON {
		log.Printf(
			colorCyan+"[DEBUG]"+colorReset+" task=%s engine=%s args=%s"+colorReset+"\n",
			task.Name,
			engine,
			strings.Join(runArgs, " "),
		)
	}

	defer func() {
		stopTaskAndRM(task, engine, volume)
	}()

	if err := runCommand(ctx, engine, runArgs, env, logFile); err != nil {

		return fmt.Errorf("task '%s' failed %v", task.Name, err)
	}

	if err := runScript(ctx, task, engine, env, logFile); err != nil {
		return err
	}

	if err := copyArtifacts(ctx, task, engine, env, logFile); err != nil {
		return err
	}

	log.Printf(colorGreen+"[Task: %s]"+colorReset+" Completed successfully."+colorReset+"\n", task.Name)
	return nil
}
