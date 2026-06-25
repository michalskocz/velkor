package configuration

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type yamlTask struct {
	Stage  string   `yaml:"stage"`
	Image  string   `yaml:"image"`
	Repo   string   `yaml:"repo"`
	Type   string   `yaml:"type"`
	Script []string `yaml:"script"`
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

	if varNode, exists := rawMap["variables"]; exists {
		var rawVars []map[string]string
		if err := varNode.Decode(&rawVars); err != nil {
			return nil, fmt.Errorf("failed to decode variables section: %w", err)
		}
		for _, m := range rawVars {
			for k, v := range m {
				config.GlobalVariables = append(config.GlobalVariables, Variable{Name: k, Value: v})
			}
		}
	}

	if threadsNode, exists := rawMap["threads"]; exists {
		var threads int
		if err := threadsNode.Decode(&threads); err != nil {
			return nil, fmt.Errorf("failed to decode threads section: %w", err)
		}

		if threads <= 0 {
			return nil, fmt.Errorf("threads must be greater than 0")
		}

		config.Threads = threads
	}

	taskQueueMap := make(map[string]*TaskQueue)

	if stagesNode, exists := rawMap["stages"]; exists {
		var stageNames []string
		if err := stagesNode.Decode(&stageNames); err != nil {
			return nil, fmt.Errorf("failed to decode stages section: %w", err)
		}
		for _, name := range stageNames {
			tq := &TaskQueue{tasks: []Task{}}
			stage := Stage{
				Name:  name,
				Tasks: tq,
			}
			taskQueueMap[name] = tq

			if err := config.Stages.Add(stage); err != nil {
				return nil, fmt.Errorf("failed to add stage %s: %w", name, err)
			}
		}
	}

	for key, node := range rawMap {
		if key == "stages" || key == "variables" || key == "threads" {
			continue
		}

		var yt yamlTask
		if err := node.Decode(&yt); err != nil {
			return nil, fmt.Errorf("failed to decode task section '%s': %w", key, err)
		}

		containerType := DEFAULT_CONTAINER_TYPE
		switch strings.ToLower(yt.Type) {
		case "podman":
			containerType = PODMAN
		case "docker":
			containerType = DOCKER
		}

		task := Task{
			Name: key,
			Image: DockerImage{
				ContainerTye: containerType,
				Name:         yt.Image,
				Repo:         yt.Repo,
			},
			Script:    yt.Script,
			Variables: []Variable{},
		}

		tq, exists := taskQueueMap[yt.Stage]
		if !exists {
			return nil, fmt.Errorf("task '%s' is assigned to a non-existent stage '%s'", key, yt.Stage)
		}

		if err := tq.Add(task); err != nil {
			return nil, fmt.Errorf("failed to add task '%s' to stage '%s': %w", key, yt.Stage, err)
		}
	}

	return config, nil
}
