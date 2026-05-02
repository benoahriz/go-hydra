package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

var execLookPath = exec.LookPath
var execCommandContext = exec.CommandContext

type DockerCLIRunner struct{}

func (r DockerCLIRunner) engineBinary() string {
	if _, err := execLookPath("docker"); err == nil {
		return "docker"
	}
	if _, err := execLookPath("podman"); err == nil {
		return "podman"
	}
	return "docker"
}

func (r DockerCLIRunner) Run(ctx context.Context, spec FunctionSpec, stdin []byte) ([]byte, []byte, error) {
	timeout := time.Duration(spec.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := []string{"run", "-i", spec.Image}
	args = append(args, spec.Command...)

	cmd := execCommandContext(ctx, r.engineBinary(), args...)
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
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("docker run failed: %w", err)
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}
