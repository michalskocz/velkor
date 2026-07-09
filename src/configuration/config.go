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
	"errors"
	"sync"

	"github.com/michalskocz/velkor/src/internal"
)

type Variable struct {
	Name  string
	Value string
}

const DOCKER = 0
const PODMAN = 1

type DockerImage struct {
	ContainerType int // Docker or Podman. Int in case of more option in feature
	Name          string
}

type Task struct {
	Name  string
	Image DockerImage

	Script    []string
	Variables []Variable
	Artifacts []string
}

// Checking the Queu Implementation
var _ internal.Queue = &TaskQueue{}

type TaskQueue struct {
	tasks  []Task
	curent int
	mut    sync.Mutex
}

// Checking the Queu Implementation
var _ internal.Queue = &StageQueue{}

type StageQueue struct {
	stages []Stage
	curent int
	mut    sync.Mutex
}

type Stage struct {
	Tasks *TaskQueue
	Name  string
}

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
	s.mut.Lock()
	defer s.mut.Unlock()
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
		s.mut.Lock()
		stage = s.stages[s.curent]
		s.curent++
		s.mut.Unlock()
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

	defer s.mut.Unlock()
	s.mut.Lock()

	for _, ss := range s.stages {
		if ss.Name == stage.Name {
			return errors.New("Duplicated Stage in Queue")
		}
	}
	s.stages = append(s.stages, stage)
	return nil
}
