package main

import "testing"

func resetRepoState() {
	currentId = 0
	todos = Todos{}
}

func TestRepoCreateFindDestroyTodo(t *testing.T) {
	resetRepoState()

	created := RepoCreateTodo(Todo{Name: "test"})
	if created.Id != 1 {
		t.Fatalf("created id = %d, want 1", created.Id)
	}

	found := RepoFindTodo(created.Id)
	if found.Id != created.Id {
		t.Fatalf("found id = %d, want %d", found.Id, created.Id)
	}

	if err := RepoDestroyTodo(created.Id); err != nil {
		t.Fatalf("RepoDestroyTodo() error = %v", err)
	}

	missing := RepoFindTodo(created.Id)
	if missing.Id != 0 {
		t.Fatalf("missing todo id = %d, want 0", missing.Id)
	}
}

func TestRepoDestroyTodo_NotFound(t *testing.T) {
	resetRepoState()

	err := RepoDestroyTodo(99)
	if err == nil {
		t.Fatal("RepoDestroyTodo() error = nil, want non-nil")
	}
}
