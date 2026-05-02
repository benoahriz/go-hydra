package main

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type BinaryRunner struct{}

func (r BinaryRunner) Run(ctx context.Context, spec FunctionSpec, stdin []byte) ([]byte, []byte, error) {
	if spec.BinaryPath == "" {
		return nil, nil, fmt.Errorf("binary path is required")
	}

	timeout := time.Duration(spec.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := execCommandContext(ctx, spec.BinaryPath, spec.Args...)
	cmd.Stdin = bytes.NewReader(stdin)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("function timeout exceeded")
	}
	if err != nil {
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("binary run failed: %w", err)
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}
