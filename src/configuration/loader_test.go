package configuration

import (
	"testing"
)

const validYAML = `
threads: 4
variables:
  - ENV: production
  - VERSION: 1.0.0
stages:
  - build
  - deploy

build_task:
  stage: build
  image: golang:1.20
  type: podman
  script:
    - go build -o app
    - echo "Build done"

deploy_task:
  stage: deploy
  image: alpine:latest
  type: docker
  script:
    - echo "Deploying..."
`

func TestParseYAMLConfig_Success(t *testing.T) {
	config, err := ParseYAMLConfig([]byte(validYAML))
	if err != nil {
		t.Fatalf("Nie udało się sparsować poprawnego YAML: %v", err)
	}

	if config.Threads != 4 {
		t.Errorf("Oczekiwano 4 wątków, otrzymano %d", config.Threads)
	}

	if len(config.GlobalVariables) != 2 {
		t.Errorf("Oczekiwano 2 zmiennych globalnych, otrzymano %d", len(config.GlobalVariables))
	}

	if len(config.Stages.stages) != 2 {
		t.Fatalf("Oczekiwano 2 etapów, otrzymano %d", len(config.Stages.stages))
	}

	buildStage := config.Stages.stages[0]
	if buildStage.Name != "build" {
		t.Errorf("Oczekiwano etapu 'build', otrzymano '%s'", buildStage.Name)
	}
	if len(buildStage.Tasks.tasks) != 1 {
		t.Fatalf("Oczekiwano 1 zadania w etapie build, otrzymano %d", len(buildStage.Tasks.tasks))
	}

	task := buildStage.Tasks.tasks[0]
	if task.Name != "build_task" {
		t.Errorf("Oczekiwano zadania 'build_task', otrzymano '%s'", task.Name)
	}
	if task.Image.ContainerTye != PODMAN {
		t.Errorf("Oczekiwano typu kontenera PODMAN (%d), otrzymano %d", PODMAN, task.Image.ContainerTye)
	}
}

func TestParseYAMLConfig_InvalidThreads(t *testing.T) {
	invalidYAML := `
threads: 0
stages:
  - build
`
	_, err := ParseYAMLConfig([]byte(invalidYAML))
	if err == nil {
		t.Error("Oczekiwano błędu z powodu 0 wątków, otrzymano nil")
	}
}

func BenchmarkParseYAMLConfig(b *testing.B) {
	data := []byte(validYAML)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseYAMLConfig(data)
		if err != nil {
			b.Fatalf("Benchmark przerwany przez błąd: %v", err)
		}
	}
}
