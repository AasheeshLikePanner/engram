package sandbox

import (
	"bytes"
	"context"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type Sandbox interface {
	Execute(ctx context.Context, wasmBinary []byte, args []string) (string, error)
}

type WasmSandbox struct {
	runtime wazero.Runtime
}

func NewWasmSandbox(ctx context.Context) (*WasmSandbox, error) {
	r := wazero.NewRuntime(ctx)
	// Instantiate WASI if needed
	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	return &WasmSandbox{runtime: r}, nil
}

func (s *WasmSandbox) Execute(ctx context.Context, wasmBinary []byte, args []string) (string, error) {
	if wasmBinary == nil {
		return "", fmt.Errorf("no WASM binary provided for execution")
	}
	// 1. Compile the module
	compiled, err := s.runtime.CompileModule(ctx, wasmBinary)
	if err != nil {
		return "", fmt.Errorf("failed to compile module: %w", err)
	}
	defer compiled.Close(ctx)

	// 2. Prepare for stdout capture
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// 3. Setup module configuration
	conf := wazero.NewModuleConfig().
		WithArgs(args...).
		WithStdout(&stdout).
		WithStderr(&stderr)

	// 4. Instantiate and run the module
	// InstantiateModule runs the _start function by default for WASI modules
	mod, err := s.runtime.InstantiateModule(ctx, compiled, conf)
	if err != nil {
		return "", fmt.Errorf("failed to instantiate module: %w (stderr: %s)", err, stderr.String())
	}
	defer mod.Close(ctx)

	return stdout.String(), nil
}

func (s *WasmSandbox) Close(ctx context.Context) error {
	return s.runtime.Close(ctx)
}
