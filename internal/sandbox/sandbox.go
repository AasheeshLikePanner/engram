package sandbox

import (
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
	// For now, this is a skeleton for Pillar 3.
	// In a real scenario, we would load the wasmBinary and run the exported function.

	// Mocking the execution for the demonstration of the Durable Agent Engine
	return fmt.Sprintf("WASM Execution Result for args %v (Simulated)", args), nil
}

func (s *WasmSandbox) Close(ctx context.Context) error {
	return s.runtime.Close(ctx)
}
