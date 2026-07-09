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
	"testing"
)

func TestParseYAMLConfig_Valid(t *testing.T) {
	yamlData := []byte(`
threads: 4
variables:
  - env: production
stages:
  - build
test_task:
  stage: build
  image: alpine
  repo: local
  type: docker
  script:
    - echo "hello"
`)

	config, err := ParseYAMLConfig(yamlData)
	if err != nil {
		t.Fatalf("ParseYAMLConfig() unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("ParseYAMLConfig() returned nil config")
	}

	if config.Threads != 4 {
		t.Errorf("Threads = %d; want 4", config.Threads)
	}
}

func TestParseYAMLConfig_InvalidYAML(t *testing.T) {
	yamlData := []byte(`: invalid`)
	_, err := ParseYAMLConfig(yamlData)
	if err == nil {
		t.Fatal("ParseYAMLConfig() error = nil; want error for invalid YAML")
	}
}

func TestParseYAMLConfig_InvalidThreads(t *testing.T) {
	yamlData := []byte(`
threads: -1
stages:
  - build
`)
	_, err := ParseYAMLConfig(yamlData)
	if err == nil {
		t.Fatal("ParseYAMLConfig() error = nil; want error for negative threads")
	}
}

func TestParseYAMLConfig_NonExistentStage(t *testing.T) {
	yamlData := []byte(`
stages:
  - build
test_task:
  stage: deploy
  image: alpine
`)
	_, err := ParseYAMLConfig(yamlData)
	if err == nil {
		t.Fatal("ParseYAMLConfig() error = nil; want error for non-existent stage assignment")
	}
}

func TestNotNullNodes_NilData(t *testing.T) {
	err := notNullNodes(nil, &Config{})
	if !errors.Is(err, errNullPointer) {
		t.Errorf("notNullNodes() error = %v; want %v", err, errNullPointer)
	}
}

func BenchmarkParseYAMLConfig(b *testing.B) {
	yamlData := []byte(`
threads: 8
variables:
  - key1: value1
  - key2: value2
stages:
  - test
  - deploy
task1:
  stage: test
  image: ubuntu
  repo: remote
  type: podman
  script:
    - make test
task2:
  stage: deploy
  image: alpine
  repo: remote
  type: docker
  script:
    - make deploy
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseYAMLConfig(yamlData)
	}
}
