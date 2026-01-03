package sandbox

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestWasmSandbox_Execute(t *testing.T) {
	ctx := context.Background()
	s, err := NewWasmSandbox(ctx)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}
	defer s.Close(ctx)

	// Load the test.wasm we just built
	wasmBinary, err := os.ReadFile("test/test.wasm")
	if err != nil {
		t.Fatalf("failed to read test.wasm: %v", err)
	}

	args := []string{"test-exe", "arg1", "arg2"}
	result, err := s.Execute(ctx, wasmBinary, args)
	if err != nil {
		t.Fatalf("failed to execute WASM: %v", err)
	}

	t.Logf("WASM output: %q", result)

	if !strings.Contains(result, "WASM_SUCCESS") {
		t.Errorf("expected output to contain WASM_SUCCESS, got %q", result)
	}
	if !strings.Contains(result, "arg1") || !strings.Contains(result, "arg2") {
		t.Errorf("expected output to contain args, got %q", result)
	}
}
