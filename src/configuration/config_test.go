package configuration

import (
	"testing"
)

func TestTaskQueue_AddAndGet(t *testing.T) {
	tq := &TaskQueue{}
	task1 := Task{Name: "build"}
	task2 := Task{Name: "test"}

	if err := tq.Add(task1); err != nil {
		t.Fatalf("Unexpected error adding task: %v", err)
	}
	if err := tq.Add(task2); err != nil {
		t.Fatalf("Unexpected error adding task: %v", err)
	}

	if err := tq.Add(task1); err == nil {
		t.Error("Expected error when adding a duplicated task, got nil")
	}

	if !tq.Next() {
		t.Fatal("Expected Next() to return true for a non-empty queue")
	}

	rawTask, err := tq.Get()
	if err != nil {
		t.Fatalf("Unexpected error getting task from queue: %v", err)
	}

	fetchedTask, ok := rawTask.(Task)
	if !ok || fetchedTask.Name != "build" {
		t.Errorf("Expected task name 'build', got: '%v'", fetchedTask.Name)
	}
}

func TestStageQueue_AddAndGet(t *testing.T) {
	sq := &StageQueue{}
	stage1 := Stage{Name: "setup"}

	if err := sq.Add(stage1); err != nil {
		t.Fatalf("Unexpected error adding stage: %v", err)
	}

	if err := sq.Add(Stage{Name: "setup"}); err == nil {
		t.Error("Expected error when adding a duplicated stage, got nil")
	}

	if !sq.Next() {
		t.Fatal("Expected Next() to return true")
	}

	rawStage, err := sq.Get()
	if err != nil {
		t.Fatalf("Unexpected error getting stage from queue: %v", err)
	}

	fetchedStage, ok := rawStage.(Stage)
	if !ok || fetchedStage.Name != "setup" {
		t.Errorf("Expected stage name 'setup', got: '%v'", fetchedStage.Name)
	}
}

func BenchmarkTaskQueue_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tq := &TaskQueue{}
		_ = tq.Add(Task{Name: "benchmark_task"})
	}
}
