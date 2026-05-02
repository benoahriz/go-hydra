package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestDockerCLIRunner_EngineBinary(t *testing.T) {
	orig := execLookPath
	defer func() { execLookPath = orig }()

	t.Run("prefers docker", func(t *testing.T) {
		execLookPath = func(file string) (string, error) {
			if file == "docker" {
				return "/usr/bin/docker", nil
			}
			return "", errors.New("not found")
		}
		if got := (DockerCLIRunner{}).engineBinary(); got != "docker" {
			t.Fatalf("engineBinary() = %q, want docker", got)
		}
	})

	t.Run("falls back to podman", func(t *testing.T) {
		execLookPath = func(file string) (string, error) {
			if file == "podman" {
				return "/usr/bin/podman", nil
			}
			return "", errors.New("not found")
		}
		if got := (DockerCLIRunner{}).engineBinary(); got != "podman" {
			t.Fatalf("engineBinary() = %q, want podman", got)
		}
	})
}

func TestDockerCLIRunner_Run(t *testing.T) {
	orig := execCommandContext
	defer func() { execCommandContext = orig }()

	spec := FunctionSpec{Name: FunctionTextUppercase, Image: "busybox", Command: []string{"tr"}, TimeoutMS: 1000}

	t.Run("success", func(t *testing.T) {
		execCommandContext = func(ctx context.Context, command string, args ...string) *exec.Cmd {
			cs := []string{"-test.run=TestHelperProcess", "--", command}
			cs = append(cs, args...)
			cmd := exec.CommandContext(ctx, os.Args[0], cs...)
			cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "HELPER_EXIT=0", "HELPER_STDOUT=HELLO")
			return cmd
		}

		stdout, _, err := (DockerCLIRunner{}).Run(context.Background(), spec, []byte("hello"))
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if string(stdout) != "HELLO" {
			t.Fatalf("stdout = %q, want %q", string(stdout), "HELLO")
		}
	})

	t.Run("failure", func(t *testing.T) {
		execCommandContext = func(ctx context.Context, command string, args ...string) *exec.Cmd {
			cs := []string{"-test.run=TestHelperProcess", "--", command}
			cs = append(cs, args...)
			cmd := exec.CommandContext(ctx, os.Args[0], cs...)
			cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "HELPER_EXIT=1", "HELPER_STDOUT=")
			return cmd
		}

		_, _, err := (DockerCLIRunner{}).Run(context.Background(), spec, []byte("hello"))
		if err == nil {
			t.Fatal("Run() error = nil, want non-nil")
		}
	})
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	_, _ = os.Stdout.WriteString(os.Getenv("HELPER_STDOUT"))
	if os.Getenv("HELPER_EXIT") == "1" {
		os.Exit(1)
	}
	os.Exit(0)
}
