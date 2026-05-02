package main

import (
	"context"
	"fmt"
)

type EngineRunner struct {
	container FunctionRunner
	binary    FunctionRunner
}

func NewEngineRunner() EngineRunner {
	return EngineRunner{
		container: DockerCLIRunner{},
		binary:    BinaryRunner{},
	}
}

func (r EngineRunner) Run(ctx context.Context, spec FunctionSpec, stdin []byte) ([]byte, []byte, error) {
	switch spec.Engine {
	case EngineContainer:
		return r.container.Run(ctx, spec, stdin)
	case EngineBinary:
		return r.binary.Run(ctx, spec, stdin)
	default:
		return nil, nil, fmt.Errorf("unsupported engine %q", spec.Engine)
	}
}
