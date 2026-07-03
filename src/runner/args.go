/*
 Copyright (c) 2026 Michał Skoczylas

 Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

 THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/

package runner

import (
	"fmt"

	"github.com/michalskocz/velkor/src/configuration"
)

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

func buildRunArgs(task configuration.Task, pwd string) ([]string, string, error) {
	args := []string{"run", "-dit"}
	var volume string
	if task.Image.ContainerType == configuration.DOCKER {
		volume, err := createOverlayVolume(pwd)
		if err != nil {
			return []string{}, volume, err
		}
		args = append(args, "-v", fmt.Sprintf("%s:/workspace", volume))
	} else {
		args = append(args, "-v", fmt.Sprintf("%s:/workspace:O", pwd))
		args = append(args, "--security-opt", "label=disable")
	}

	if !goodIsolation {
		args = append(args, "--ipc=host", "--pid=host", "--userns=host")
	}

	args = append(args, "--name", task.Name, "-w", "/workspace")
	args = append(args, fmt.Sprintf("--cpus=%d", cpu))
	args = append(args, getContainerEnvironment(task)...)
	args = append(args, task.Image.Name)

	return args, volume, nil
}
