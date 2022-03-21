package bench

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	internalwasi "github.com/tetratelabs/wazero/internal/wasi"
	wasm "github.com/tetratelabs/wazero/internal/wasm"
	"github.com/tetratelabs/wazero/wasi"
)

var testMem = &wasm.MemoryInstance{
	Min: 1,
	Buffer: []byte{
		0,                // environBuf is after this
		'a', '=', 'b', 0, // null terminated "a=b",
		'b', '=', 'c', 'd', 0, // null terminated "b=cd"
		0,          // environ is after this
		1, 0, 0, 0, // little endian-encoded offset of "a=b"
		5, 0, 0, 0, // little endian-encoded offset of "b=cd"
		0,
	},
}

func Test_EnvironGet(t *testing.T) {
	sys, err := wasm.NewSystemContext()
	require.NoError(t, err)

	err = sys.WithEnviron("a=b", "b=cd")
	require.NoError(t, err)

	testCtx := newCtx(make([]byte, 20), sys)
	environGet := internalwasi.NewAPI().EnvironGet

	require.Equal(t, wasi.ErrnoSuccess, environGet(testCtx, 11, 1))
	require.Equal(t, testCtx.Memory(), testMem)
}

func Benchmark_EnvironGet(b *testing.B) {
	sys, err := wasm.NewSystemContext()
	if err != nil {
		b.Fatal(err)
	}

	err = sys.WithEnviron("a=b", "b=cd")
	if err != nil {
		b.Fatal(err)
	}

	testCtx := newCtx([]byte{
		0,                // environBuf is after this
		'a', '=', 'b', 0, // null terminated "a=b",
		'b', '=', 'c', 'd', 0, // null terminated "b=cd"
		0,          // environ is after this
		1, 0, 0, 0, // little endian-encoded offset of "a=b"
		5, 0, 0, 0, // little endian-encoded offset of "b=cd"
		0,
	}, sys)

	environGet := internalwasi.NewAPI().EnvironGet
	b.Run("EnvironGet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if environGet(testCtx, 0, 4) != wasi.ErrnoSuccess {
				b.Fatal()
			}
		}
	})
}

func newCtx(buf []byte, sys *wasm.SystemContext) *wasm.ModuleContext {
	ret := wasm.NewModuleContext(context.Background(), nil, &wasm.ModuleInstance{
		Memory: &wasm.MemoryInstance{Min: 1, Buffer: buf},
	})
	ret.System = sys
	return ret
}
