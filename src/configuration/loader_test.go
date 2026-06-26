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
