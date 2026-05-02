package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestConvertAndUpperHandlers_NoFilename(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "/convert/pdf", nil)
	rr1 := httptest.NewRecorder()
	convertPdfHandler(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("convertPdfHandler status = %d, want %d", rr1.Code, http.StatusOK)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/toupper/txt", nil)
	rr2 := httptest.NewRecorder()
	toUpperHandler(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("toUpperHandler status = %d, want %d", rr2.Code, http.StatusOK)
	}
}

func TestUpload_GetAndPost(t *testing.T) {
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}

	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, "test"), 0o755); err != nil {
		t.Fatalf("mkdir test dir error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "upload.gtpl"), []byte("token={{.}}"), 0o644); err != nil {
		t.Fatalf("write template error = %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir error = %v", err)
	}
	defer func() { _ = os.Chdir(origWD) }()

	// GET branch
	reqGet := httptest.NewRequest(http.MethodGet, "/upload", nil)
	rrGet := httptest.NewRecorder()
	upload(rrGet, reqGet)
	if rrGet.Code != http.StatusOK {
		t.Fatalf("upload GET status = %d, want %d", rrGet.Code, http.StatusOK)
	}
	if !strings.Contains(rrGet.Body.String(), "token=") {
		t.Fatalf("upload GET body = %q, expected token output", rrGet.Body.String())
	}

	// POST success branch
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("uploadfile", "a.txt")
	if err != nil {
		t.Fatalf("CreateFormFile error = %v", err)
	}
	if _, err := part.Write([]byte("hello file")); err != nil {
		t.Fatalf("part.Write error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close error = %v", err)
	}

	reqPost := httptest.NewRequest(http.MethodPost, "/upload", &body)
	reqPost.Header.Set("Content-Type", writer.FormDataContentType())
	rrPost := httptest.NewRecorder()
	upload(rrPost, reqPost)

	savedPath := filepath.Join(tmp, "test", "a.txt")
	saved, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("expected saved file error = %v", err)
	}
	if string(saved) != "hello file" {
		t.Fatalf("saved file content = %q, want %q", string(saved), "hello file")
	}
}

func TestTodoCreate_InvalidJSON(t *testing.T) {
	resetRepoState()
	req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()

	TodoCreate(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("TodoCreate status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestTodoShow_InvalidIDReturnsBadRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/todos/not-int", nil)
	req = mux.SetURLVars(req, map[string]string{"todoId": "not-int"})
	rr := httptest.NewRecorder()
	TodoShow(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("TodoShow status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
