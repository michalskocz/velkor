/*
 Copyright (c) 2026 Michał Skoczylas

 Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

 THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/

package configuration

import (
	"fmt"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	threads_str   = "threads"
	variables_str = "variables"
	stages_str    = "stages"
)

type yamlTask struct {
	Stage     string              `yaml:"stage"`
	Image     string              `yaml:"image"`
	Repo      string              `yaml:"repo"`
	Type      string              `yaml:"type"`
	Script    []string            `yaml:"script"`
	Artifacts []string            `yaml:"artifacts"`
	Variables []map[string]string `yaml:"variables"`
}

func testForContainer() int {
	if _, err := exec.LookPath("docker"); err == nil {
		return DOCKER
	}

	if _, err := exec.LookPath("podman"); err == nil {
		return PODMAN
	}

	return -1
}

func ParseYAMLConfig(yamlData []byte) (*Config, error) {

	var rawMap map[string]yaml.Node
	if err := yaml.Unmarshal(yamlData, &rawMap); err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	config := &Config{
		GlobalVariables: []Variable{},
		Stages:          StageQueue{stages: []Stage{}},
		Threads:         DEFAULT_THREADS,
	}

	if err := lookUpVars(&rawMap, config); err != nil {
		return nil, err
	}

	if err := lookUpThreads(&rawMap, config); err != nil {
		return nil, err
	}

	taskQueueMap := make(map[string]*TaskQueue)

	if err := lookUpTasks(&rawMap, config, &taskQueueMap); err != nil {
		return nil, err
	}

	if err := parseTasks(&rawMap, config, &taskQueueMap); err != nil {
		return nil, err
	}

	return config, nil
}

func lookUpThreads(data *map[string]yaml.Node, config *Config) error {
	if err := notNullNodes(data, config); err != nil {
		return err
	}
	rawMap := *data
	if threadsNode, exists := rawMap[threads_str]; exists {
		var threads int
		if err := threadsNode.Decode(&threads); err != nil {
			return fmt.Errorf("failed to decode %s section: %w", threads_str, err)
		}

		if threads <= 0 {
			return fmt.Errorf("%s must be greater than 0", threads_str)
		}

		config.Threads = threads
	}

	return nil
}

func lookUpVars(data *map[string]yaml.Node, config *Config) error {
	if err := notNullNodes(data, config); err != nil {
		return err
	}
	rawMap := *data

	if varNode, exists := rawMap[variables_str]; exists {
		var rawVars []map[string]string
		if err := varNode.Decode(&rawVars); err != nil {
			return fmt.Errorf("failed to decode variables section: %w", err)
		}
		for _, m := range rawVars {
			for k, v := range m {
				config.GlobalVariables = append(config.GlobalVariables, Variable{Name: k, Value: v})
			}
		}
	}
	return nil
}

func lookUpTasks(data *map[string]yaml.Node, config *Config, taskMap *map[string]*TaskQueue) error {
	if err := notNullNodes(data, config); err != nil {
		return err
	} else if taskMap == nil {
		return errNullPointer
	}

	rawMap := *data
	taskQueueMap := *taskMap

	if stagesNode, exists := rawMap[stages_str]; exists {
		var stageNames []string
		if err := stagesNode.Decode(&stageNames); err != nil {
			return fmt.Errorf("failed to decode %s section: %w", stages_str, err)
		}
		for _, name := range stageNames {
			tq := &TaskQueue{tasks: []Task{}}
			stage := Stage{
				Name:  name,
				Tasks: tq,
			}
			taskQueueMap[name] = tq

			if err := config.Stages.Add(stage); err != nil {
				return fmt.Errorf("failed to add stage %s: %w", name, err)
			}
		}
	}

	return nil
}

func parseTasks(data *map[string]yaml.Node, config *Config, taskMap *map[string]*TaskQueue) error {
	if err := notNullNodes(data, config); err != nil {
		return err
	} else if taskMap == nil {
		return errNullPointer
	}
	rawMap := *data
	taskQueueMap := *taskMap

	for key, node := range rawMap {
		if key == stages_str || key == variables_str || key == threads_str {
			continue
		}

		var yt yamlTask
		if err := node.Decode(&yt); err != nil {
			return fmt.Errorf("failed to decode task section '%s': %w", key, err)
		}

		containerType := testForContainer()
		switch strings.ToLower(yt.Type) {
		case "podman":
			containerType = PODMAN
		case "docker":
			containerType = DOCKER
		}

		taskVars := make([]Variable, 0)

		for _, m := range yt.Variables {
			for k, v := range m {
				taskVars = append(taskVars, Variable{
					Name:  k,
					Value: v,
				})
			}
		}

		task := Task{
			Name: key,
			Image: DockerImage{
				ContainerType: containerType,
				Name:          yt.Image,
				Repo:          yt.Repo,
			},
			Script:    yt.Script,
			Variables: taskVars,
			Artifacts: yt.Artifacts,
		}

		tq, exists := taskQueueMap[yt.Stage]
		if !exists {
			return fmt.Errorf("task '%s' is assigned to a non-existent stage '%s'", key, yt.Stage)
		}

		if err := tq.Add(task); err != nil {
			return fmt.Errorf("failed to add task '%s' to stage '%s': %w", key, yt.Stage, err)
		}
	}
	return nil
}

var errNullPointer = fmt.Errorf("Null pointer error in look up")

func notNullNodes(data *map[string]yaml.Node, config *Config) error {
	if data == nil || config == nil {
		return errNullPointer
	}

	return nil
}
