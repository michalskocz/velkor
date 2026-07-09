/*
 Copyright (c) 2026 Michał Skoczylas

 Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer
    in the documentation and/or other materials provided with the distribution.

 THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
 INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
 IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
 OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA,
 OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/

package configuration

import (
	"fmt"
	"strings"

	"github.com/michalskocz/velkor/src/internal"
)

func (c *Config) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "\n%s%s=== CONFIGURATION ===%s\n", internal.ColorBold, internal.ColorCyan, internal.ColorReset)
	fmt.Fprintf(&sb, "%sThreads:%s %d\n", internal.ColorBold, internal.ColorReset, c.Threads)
	fmt.Fprintf(&sb, "\n%s%s[Global Variables]%s\n", internal.ColorBold, internal.ColorBlue, internal.ColorReset)

	if len(c.GlobalVariables) == 0 {
		sb.WriteString("  None\n")
	} else {
		for _, v := range c.GlobalVariables {
			fmt.Fprintf(&sb, "  %s%s%s = %s%s%s\n", internal.ColorGreen, v.Name, internal.ColorReset, internal.ColorYellow, v.Value, internal.ColorReset)
		}
	}

	fmt.Fprintf(&sb, "\n%s%s[Stages and Tasks]%s\n", internal.ColorBold, internal.ColorPurple, internal.ColorReset)
	if len(c.Stages.stages) == 0 {
		sb.WriteString("  No stages defined\n")
	}

	for i, stage := range c.Stages.stages {
		fmt.Fprintf(&sb, "%s► Stage %d: %s%s\n", internal.ColorBold, i+1, stage.Name, internal.ColorReset)

		if stage.Tasks == nil {
			sb.WriteString("    No tasks\n")
			continue
		}

		stage.Tasks.mut.Lock()
		tasks := stage.Tasks.tasks
		if len(tasks) == 0 {
			sb.WriteString("    No tasks\n")
		}

		for _, task := range tasks {
			fmt.Fprintf(&sb, "    %s◇ Task: %s%s\n", internal.ColorCyan, task.Name, internal.ColorReset)

			cType := "Docker"
			if task.Image.ContainerType == PODMAN {
				cType = "Podman"
			}

			fmt.Fprintf(&sb, "      Image:  %s%s%s (%s)\n", internal.ColorYellow, task.Image.Name, internal.ColorReset, cType)

			if len(task.Variables) > 0 {
				sb.WriteString("      Task variables:\n")
				for _, tv := range task.Variables {
					fmt.Fprintf(&sb, "        %s%s%s = %s\n", internal.ColorGreen, tv.Name, internal.ColorReset, tv.Value)
				}
			}

			if len(task.Script) > 0 {
				sb.WriteString("      Script:\n")
				for _, line := range task.Script {
					fmt.Fprintf(&sb, "        > %s\n", line)
				}
			}
		}
		stage.Tasks.mut.Unlock()
	}

	fmt.Fprintf(&sb, "%s%s====================%s\n", internal.ColorBold, internal.ColorCyan, internal.ColorReset)
	return sb.String()
}
