package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeRunner struct{}

type errRunner struct{}

func (f fakeRunner) Run(ctx context.Context, spec FunctionSpec, stdin []byte) ([]byte, []byte, error) {
	if spec.Name == FunctionTextUppercase {
		return bytes.ToUpper(stdin), nil, nil
	}
	if spec.Name == FunctionConvertMarkdown {
		return []byte("# converted"), nil, nil
	}
	if spec.Name == FunctionRenderURLToPDF {
		return []byte("pdf-bytes"), nil, nil
	}
	return nil, nil, nil
}

func (f errRunner) Run(ctx context.Context, spec FunctionSpec, stdin []byte) ([]byte, []byte, error) {
	return nil, []byte("boom"), errors.New("runner failed")
}

func TestValidateMarkdownInput(t *testing.T) {
	tests := []struct {
		name    string
		in      MarkdownInput
		wantErr bool
	}{
		{name: "url only valid", in: MarkdownInput{URL: "https://example.com"}, wantErr: false},
		{name: "html only valid", in: MarkdownInput{HTML: "<html></html>"}, wantErr: false},
		{name: "both invalid", in: MarkdownInput{URL: "https://example.com", HTML: "<html></html>"}, wantErr: true},
		{name: "neither invalid", in: MarkdownInput{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMarkdownInput(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateMarkdownInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInvokeFunctionHandler(t *testing.T) {
	originalRunner := invokeRunner
	invokeRunner = fakeRunner{}
	defer func() { invokeRunner = originalRunner }()

	tests := []struct {
		name       string
		method     string
		body       interface{}
		wantStatus int
		wantOK     bool
	}{
		{
			name:       "uppercase success",
			method:     http.MethodPost,
			body:       map[string]interface{}{"function": FunctionTextUppercase, "input": map[string]interface{}{"text": "hello"}},
			wantStatus: http.StatusOK,
			wantOK:     true,
		},
		{
			name:       "unknown function",
			method:     http.MethodPost,
			body:       map[string]interface{}{"function": "unknown.function", "input": map[string]interface{}{}},
			wantStatus: http.StatusNotFound,
			wantOK:     false,
		},
		{
			name:       "render url to pdf success",
			method:     http.MethodPost,
			body:       map[string]interface{}{"function": FunctionRenderURLToPDF, "input": map[string]interface{}{"url": "https://example.com"}},
			wantStatus: http.StatusOK,
			wantOK:     true,
		},
		{
			name:       "markdown from html success",
			method:     http.MethodPost,
			body:       map[string]interface{}{"function": FunctionConvertMarkdown, "input": map[string]interface{}{"html": "<html><body>ok</body></html>"}},
			wantStatus: http.StatusOK,
			wantOK:     true,
		},
		{
			name:       "missing function",
			method:     http.MethodPost,
			body:       map[string]interface{}{"input": map[string]interface{}{"text": "hello"}},
			wantStatus: http.StatusUnprocessableEntity,
			wantOK:     false,
		},
		{
			name:       "invalid uppercase input",
			method:     http.MethodPost,
			body:       map[string]interface{}{"function": FunctionTextUppercase, "input": map[string]interface{}{"text": "   "}},
			wantStatus: http.StatusUnprocessableEntity,
			wantOK:     false,
		},
		{
			name:       "invalid render url input",
			method:     http.MethodPost,
			body:       map[string]interface{}{"function": FunctionRenderURLToPDF, "input": map[string]interface{}{"url": "example.com"}},
			wantStatus: http.StatusUnprocessableEntity,
			wantOK:     false,
		},
		{
			name:       "markdown both fields invalid",
			method:     http.MethodPost,
			body:       map[string]interface{}{"function": FunctionConvertMarkdown, "input": map[string]interface{}{"url": "https://example.com", "html": "<html></html>"}},
			wantStatus: http.StatusUnprocessableEntity,
			wantOK:     false,
		},
		{
			name:       "method not allowed",
			method:     http.MethodGet,
			body:       map[string]interface{}{},
			wantStatus: http.StatusMethodNotAllowed,
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			req := httptest.NewRequest(tt.method, "/functions/invoke", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			InvokeFunctionHandler(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}

			var got InvokeResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("json.Unmarshal(response) error = %v; body=%s", err, rr.Body.String())
			}

			if got.OK != tt.wantOK {
				t.Fatalf("response ok = %v, want %v", got.OK, tt.wantOK)
			}

			if got.OK && rr.Header().Get("X-Function-Output-Content-Type") == "" {
				t.Fatalf("missing X-Function-Output-Content-Type header for success response")
			}
		})
	}
}

func TestInvokeFunctionHandler_InvalidJSON(t *testing.T) {
	originalRunner := invokeRunner
	invokeRunner = fakeRunner{}
	defer func() { invokeRunner = originalRunner }()

	req := httptest.NewRequest(http.MethodPost, "/functions/invoke", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	InvokeFunctionHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	var got InvokeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal(response) error = %v; body=%s", err, rr.Body.String())
	}
	if got.OK {
		t.Fatalf("response ok = true, want false")
	}
}

func TestExecuteInvoke_UnknownFunction(t *testing.T) {
	req := InvokeRequest{Function: "unknown.fn", Input: json.RawMessage(`{}`)}

	_, invokeErr, status := executeInvoke(req)
	if invokeErr == nil {
		t.Fatal("invokeErr = nil, want non-nil")
	}
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", status, http.StatusNotFound)
	}
}

func TestExecuteInvoke_RunnerFailure(t *testing.T) {
	originalRunner := invokeRunner
	invokeRunner = errRunner{}
	defer func() { invokeRunner = originalRunner }()

	req := InvokeRequest{
		Function: FunctionTextUppercase,
		Input:    json.RawMessage(`{"text":"hello"}`),
	}

	_, invokeErr, status := executeInvoke(req)
	if invokeErr == nil {
		t.Fatal("invokeErr = nil, want non-nil")
	}
	if invokeErr.Code != "container_execution_failed" {
		t.Fatalf("invokeErr.Code = %q, want %q", invokeErr.Code, "container_execution_failed")
	}
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", status, http.StatusInternalServerError)
	}
}
