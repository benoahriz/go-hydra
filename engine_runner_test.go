package main

import (
	"context"
	"testing"
)

type stubRunner struct {
	stdout []byte
	stderr []byte
	err    error
}

func (s stubRunner) Run(ctx context.Context, spec FunctionSpec, stdin []byte) ([]byte, []byte, error) {
	return s.stdout, s.stderr, s.err
}

func TestEngineRunnerDispatch(t *testing.T) {
	r := EngineRunner{
		container: stubRunner{stdout: []byte("container")},
		binary:    stubRunner{stdout: []byte("binary")},
	}

	out, _, err := r.Run(context.Background(), FunctionSpec{Engine: EngineContainer}, []byte("x"))
	if err != nil {
		t.Fatalf("container run error = %v", err)
	}
	if string(out) != "container" {
		t.Fatalf("container stdout = %q, want %q", string(out), "container")
	}

	out, _, err = r.Run(context.Background(), FunctionSpec{Engine: EngineBinary}, []byte("x"))
	if err != nil {
		t.Fatalf("binary run error = %v", err)
	}
	if string(out) != "binary" {
		t.Fatalf("binary stdout = %q, want %q", string(out), "binary")
	}
}

func TestEngineRunnerUnsupported(t *testing.T) {
	r := EngineRunner{
		container: stubRunner{},
		binary:    stubRunner{},
	}

	_, _, err := r.Run(context.Background(), FunctionSpec{Engine: "unknown"}, nil)
	if err == nil {
		t.Fatal("expected error for unsupported engine")
	}
}

func TestBinaryRunnerRequiresPath(t *testing.T) {
	_, _, err := (BinaryRunner{}).Run(context.Background(), FunctionSpec{Engine: EngineBinary}, []byte("x"))
	if err == nil {
		t.Fatal("expected binary path error")
	}
}
