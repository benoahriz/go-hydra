package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTodoRoutes_CreateIndexShow(t *testing.T) {
	originalRunner := invokeRunner
	invokeRunner = fakeRunner{}
	defer func() { invokeRunner = originalRunner }()

	resetRepoState()
	router := NewRouter()

	body := bytes.NewBufferString(`{"name":"New Todo"}`)
	reqCreate := httptest.NewRequest(http.MethodPost, "/todos", body)
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	router.ServeHTTP(rrCreate, reqCreate)

	if rrCreate.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", rrCreate.Code, http.StatusCreated)
	}

	var created Todo
	if err := json.Unmarshal(rrCreate.Body.Bytes(), &created); err != nil {
		t.Fatalf("create unmarshal error = %v", err)
	}
	if created.Id == 0 {
		t.Fatal("created todo id = 0, want non-zero")
	}

	reqIndex := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rrIndex := httptest.NewRecorder()
	router.ServeHTTP(rrIndex, reqIndex)

	if rrIndex.Code != http.StatusOK {
		t.Fatalf("index status = %d, want %d", rrIndex.Code, http.StatusOK)
	}

	reqShow := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	rrShow := httptest.NewRecorder()
	router.ServeHTTP(rrShow, reqShow)

	if rrShow.Code != http.StatusOK {
		t.Fatalf("show status = %d, want %d", rrShow.Code, http.StatusOK)
	}
}

func TestTodoShow_NotFound(t *testing.T) {
	resetRepoState()
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/todos/999", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}
