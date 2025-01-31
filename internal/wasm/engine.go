package internalwasm

import "io"

// Engine is a Store-scoped mechanism to compile functions declared or imported by a module.
// This is a top-level type implemented by an interpreter or JIT compiler.
type Engine interface {
	// NewModuleEngine compiles down the function instances in a module, and returns ModuleEngine for the module.
	//
	// * name is the name the module was instantiated with used for error handling.
	// * importedFunctions: functions this module imports, already compiled in this engine.
	// * moduleFunctions: functions declared in this module that must be compiled.
	// * table: a possibly shared table used by this module. When nil tableInit will be nil.
	// * tableInit: a mapping of TableInstance.Table index to the function index it should point to.
	//
	// Note: Input parameters must be pre-validated with internalwasm.Module Validate, to ensure no fields are invalid
	// due to reasons such as out-of-bounds.
	NewModuleEngine(name string, importedFunctions, moduleFunctions []*FunctionInstance, table *TableInstance, tableInit map[Index]Index) (ModuleEngine, error)
}

// ModuleEngine implements function calls for a given module.
type ModuleEngine interface {
	// Closer releases the resources allocated by functions in this ModuleEngine.
	io.Closer
	// ^^ io.Closer not due to I/O, but to allow future static analysis to catch leaks (unclosed Closers).

	// Call invokes a function instance f with given parameters.
	// Returns the results from the function.
	// The ctx's context.Context will be the outer-most ancestor of the argument to wasm.Function.
	Call(ctx *ModuleContext, f *FunctionInstance, params ...uint64) (results []uint64, err error)
}
