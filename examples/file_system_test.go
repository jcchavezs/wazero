package examples

import (
	"bytes"
	"embed"
	_ "embed"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/wazero"
)

// catFS is an embedded filesystem limited to cat.go
//go:embed testdata/cat.go
var catFS embed.FS

// catGo is the TinyGo source
//go:embed testdata/cat.go
var catGo string

// catWasm was compiled from catGo
//go:embed testdata/cat.wasm
var catWasm []byte

// Test_Cat writes the input file to stdout, just like `cat`.
func Test_Cat(t *testing.T) {
	r := wazero.NewRuntime()
	stdoutBuf := bytes.NewBuffer(nil)

	// Since wazero uses fs.FS we can use standard libraries to do things like trim the leading path.
	rooted, err := fs.Sub(catFS, "testdata")
	require.NoError(t, err)

	// Next, setup stdout so we can verify it. Then configure the filesystem to tell cat to print itself!
	file := "cat.go"
	wasiConfig := wazero.NewWASIConfig().WithStdout(stdoutBuf).WithFS(rooted).WithArgs(file)
	wasi, err := r.InstantiateModule(wazero.WASISnapshotPreview1WithConfig(wasiConfig))
	require.NoError(t, err)
	defer wasi.Close()

	// Finally, start the program which executes the main function (compiled to Wasm as _start).
	mod, err := wazero.StartWASICommandFromSource(r, catWasm)
	require.NoError(t, err)
	defer mod.Close()

	// To ensure it worked, this verifies stdout from WebAssembly had what we expected.
	require.Equal(t, catGo, stdoutBuf.String())
}
