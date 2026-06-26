package configuration

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"velkor/internal"
)

type Variable struct {
	Name  string
	Value string
}

const DOCKER = 0
const PODMAN = 1
const DEFAULT_CONTAINER_TYPE = DOCKER

type DockerImage struct {
	ContainerTye int // Docker or Podman. Int in case of more option in feature
	Name         string
	Repo         string
}

type Task struct {
	Name  string
	Image DockerImage

	Script    []string
	Variables []Variable
}

// Checking the Queu Implementation
var _ internal.Queue = &TaskQueue{}

type TaskQueue struct {
	tasks  []Task
	curent int
	mut    sync.Mutex
}

type Stage struct {
	Tasks *TaskQueue
	Name  string
}

// Checking the Queu Implementation
var _ internal.Queue = &StageQueue{}

type StageQueue struct {
	stages []Stage
	curent int
}

const DEFAULT_THREADS = 1

type Config struct {
	GlobalVariables []Variable
	Threads         int
	Stages          StageQueue
}

func (t *TaskQueue) Next() bool {
	t.mut.Lock()
	defer t.mut.Unlock()
	return t.curent < len(t.tasks)
}

func (s *StageQueue) Next() bool {
	return s.curent < len(s.stages)
}

func (t *TaskQueue) Get() (interface{}, error) {
	var task Task
	if t.Next() {
		t.mut.Lock()
		task = t.tasks[t.curent]
		t.curent++
		t.mut.Unlock()
		return task, nil
	}
	return task, nil
}

func (s *StageQueue) Get() (interface{}, error) {
	var stage Stage
	if s.Next() {
		stage = s.stages[s.curent]
		s.curent++
		return stage, nil
	}
	return stage, nil
}

func (t *TaskQueue) Add(item interface{}) error {
	task, ok := item.(Task)
	if !ok {
		return errors.New("Invalid item type. Expected Task")
	}
	defer t.mut.Unlock()
	t.mut.Lock()

	for _, tt := range t.tasks {
		if tt.Name == task.Name {
			return errors.New("Dupciated Task in Queue")
		}
	}
	t.tasks = append(t.tasks, task)
	return nil
}

func (s *StageQueue) Add(item interface{}) error {
	stage, ok := item.(Stage)
	if !ok {
		return errors.New("Ivalid item type. Expected Stage")
	}

	for _, ss := range s.stages {
		if ss.Name == stage.Name {
			return errors.New("Duplicated Stage in Queue")
		}
	}
	s.stages = append(s.stages, stage)
	return nil
}

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
)

func (c *Config) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n%s%s=== CONFIGURATION ===%s\n", colorBold, colorCyan, colorReset))
	sb.WriteString(fmt.Sprintf("%sThreads:%s %d\n", colorBold, colorReset, c.Threads))

	sb.WriteString(fmt.Sprintf("\n%s%s[Global Variables]%s\n", colorBold, colorBlue, colorReset))
	if len(c.GlobalVariables) == 0 {
		sb.WriteString("  None\n")
	}
	for _, v := range c.GlobalVariables {
		sb.WriteString(fmt.Sprintf("  %s%s%s = %s%s%s\n", colorGreen, v.Name, colorReset, colorYellow, v.Value, colorReset))
	}

	sb.WriteString(fmt.Sprintf("\n%s%s[Stages and Tasks]%s\n", colorBold, colorPurple, colorReset))
	if len(c.Stages.stages) == 0 {
		sb.WriteString("  No stages defined\n")
	}

	for i, stage := range c.Stages.stages {
		sb.WriteString(fmt.Sprintf("%s► Stage %d: %s%s\n", colorBold, i+1, stage.Name, colorReset))

		if stage.Tasks == nil || len(stage.Tasks.tasks) == 0 {
			sb.WriteString("    No tasks\n")
			continue
		}

		stage.Tasks.mut.Lock()
		for _, task := range stage.Tasks.tasks {
			sb.WriteString(fmt.Sprintf("    %s◇ Task: %s%s\n", colorCyan, task.Name, colorReset))

			cType := "Docker"
			if task.Image.ContainerTye == PODMAN {
				cType = "Podman"
			}

			sb.WriteString(fmt.Sprintf("      Image:  %s%s%s (%s)\n", colorYellow, task.Image.Name, colorReset, cType))
			if task.Image.Repo != "" {
				sb.WriteString(fmt.Sprintf("      Repo:   %s\n", task.Image.Repo))
			}

			if len(task.Variables) > 0 {
				sb.WriteString("      Task variables:\n")
				for _, tv := range task.Variables {
					sb.WriteString(fmt.Sprintf("        %s%s%s = %s\n", colorGreen, tv.Name, colorReset, tv.Value))
				}
			}

			if len(task.Script) > 0 {
				sb.WriteString("      Script:\n")
				for _, line := range task.Script {
					sb.WriteString(fmt.Sprintf("        > %s\n", line))
				}
			}
		}
		stage.Tasks.mut.Unlock()
	}

	sb.WriteString(fmt.Sprintf("%s%s====================%s\n", colorBold, colorCyan, colorReset))

	return sb.String()
}
