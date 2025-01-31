package examples

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/wasm"
)

type testKey struct{}

// hostFuncWasm was compiled from TinyGo testdata/host_func.go
//go:embed testdata/host_func.wasm
var hostFuncWasm []byte

func Test_hostFunc(t *testing.T) {
	// The function for allocating the in-Wasm memory region.
	// We resolve this function after main module instantiation.
	allocateInWasmBuffer := func(wasm.Module, uint32) uint32 {
		panic("unimplemented")
	}

	var expectedBase64String string

	// Host-side implementation of get_random_string on Wasm import.
	getRandomBytes := func(ctx wasm.Module, retBufPtr uint32, retBufSize uint32) {
		// Assert that context values passed in from CallFunctionContext are accessible.
		contextValue := ctx.Context().Value(testKey{}).(int64)
		require.Equal(t, int64(12345), contextValue)

		const bufferSize = 10
		offset := allocateInWasmBuffer(ctx, bufferSize)

		// Store the address info to the memory.
		require.True(t, ctx.Memory().WriteUint32Le(retBufPtr, offset))
		require.True(t, ctx.Memory().WriteUint32Le(retBufSize, uint32(bufferSize)))

		// Now store the random values in the region.
		b, ok := ctx.Memory().Read(offset, bufferSize)
		require.True(t, ok)

		n, err := rand.Read(b)
		require.NoError(t, err)
		require.Equal(t, bufferSize, n)

		expectedBase64String = base64.StdEncoding.EncodeToString(b)
	}

	r := wazero.NewRuntime()

	_, err := r.NewModuleBuilder("env").ExportFunction("get_random_bytes", getRandomBytes).Instantiate()
	require.NoError(t, err)

	// Note: host_func.go doesn't directly use WASI, but TinyGo needs to be initialized as a WASI Command.
	stdout := bytes.NewBuffer(nil)
	wasi, err := r.InstantiateModule(wazero.WASISnapshotPreview1WithConfig(wazero.NewWASIConfig().WithStdout(stdout)))
	require.NoError(t, err)
	defer wasi.Close()

	module, err := wazero.StartWASICommandFromSource(r, hostFuncWasm)
	require.NoError(t, err)
	defer module.Close()

	allocateInWasmBufferFn := module.ExportedFunction("allocate_buffer")
	require.NotNil(t, allocateInWasmBuffer)

	// Implement the function pointer. This mainly shows how you can decouple a module function dependency.
	allocateInWasmBuffer = func(ctx wasm.Module, size uint32) uint32 {
		res, err := allocateInWasmBufferFn.Call(ctx, uint64(size))
		require.NoError(t, err)
		return uint32(res[0])
	}

	// Set a context variable that should be available in wasm.Module.
	ctx := context.WithValue(context.Background(), testKey{}, int64(12345))

	// Invoke a module-defined function that depends on a host function import
	_, err = module.ExportedFunction("base64").Call(module.WithContext(ctx))
	require.NoError(t, err)

	// Verify that in-Wasm calculated base64 string matches the one calculated in native Go.
	require.Equal(t, expectedBase64String, strings.TrimSpace(stdout.String()))
}
