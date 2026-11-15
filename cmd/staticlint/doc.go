// Package main implements staticlint, the repository's combined Go multichecker.
//
// Run the checker from the repository root with:
//
//	go run ./cmd/staticlint ./...
//
// Build a reusable executable with:
//
//	go build -o bin/staticlint ./cmd/staticlint
//	./bin/staticlint ./...
//
// The checker combines the complete go vet analyzer suite from
// golang.org/x/tools/go/analysis/passes (plus shadow), every SA analyzer
// published by Staticcheck, every ST style analyzer, the project's own
// osexit analyzer, and the public nilerr and forcetypeassert analyzers.
//
// The go vet analyzers are:
//
//   - appends: detects suspicious append calls that may overwrite data.
//   - asmdecl: checks consistency between Go assembly declarations and bodies.
//   - assign: detects useless self assignments.
//   - atomic: detects mistakes in use of sync/atomic.
//   - bools: detects suspicious Boolean expressions.
//   - buildtag: validates build constraints and their placement.
//   - cgocall: checks invalid Go pointer passing to C.
//   - composite: detects unkeyed literals of external struct types.
//   - copylock: detects copying values containing locks.
//   - defers: detects common mistakes in defer statements.
//   - directive: validates compiler and tool directives.
//   - errorsas: validates the target passed to errors.As.
//   - framepointer: checks assembly frame-pointer declarations.
//   - hostport: detects malformed host and port construction.
//   - httpresponse: checks HTTP response handling and body closing.
//   - ifaceassert: detects impossible interface assertions.
//   - loopclosure: detects loop-variable capture by escaping closures.
//   - lostcancel: detects context cancel functions that are not called.
//   - nilfunc: detects comparisons of functions with nil that are always false.
//   - printf: checks printf-style formatting calls and arguments.
//   - shadow: detects variables that unintentionally shadow outer variables.
//   - shift: detects invalid or suspicious shift operations.
//   - sigchanyzer: checks signal.Notify channel usage.
//   - slog: checks structured logging calls made with log/slog.
//   - stdmethods: checks method signatures that should implement standard interfaces.
//   - stdversion: detects use of standard-library APIs newer than the module Go version.
//   - stringintconv: detects suspicious conversions between strings and integers.
//   - structtag: validates struct field tags.
//   - testinggoroutine: detects unsafe testing.T calls from goroutines.
//   - tests: checks test and example declarations for common mistakes.
//   - timeformat: checks time format strings.
//   - unmarshal: checks standard unmarshaler method signatures.
//   - unreachable: detects unreachable code.
//   - unsafeptr: checks suspicious unsafe.Pointer conversions.
//   - unusedresult: detects ignored results of selected functions.
//   - waitgroup: detects incorrect sync.WaitGroup usage.
//
// Staticcheck's SA family finds correctness bugs and suspicious standard-library
// use. All SA analyzers exposed by honnef.co/go/tools/staticcheck are registered;
// their individual IDs and descriptions are documented in ANALYZERS.md. The ST
// family reports style problems such as missing package documentation, naming
// mistakes, and inconsistent formatting conventions. Each ST analyzer's own
// documentation describes its specific check.
//
// The project's osexit analyzer rejects direct os.Exit calls inside func main
// declarations in package main. Such calls bypass deferred cleanup; put the
// application logic in a function that returns an error and handle that error
// from main instead.
//
// nilerr detects branches that return nil despite a non-nil error, or return an
// error despite a nil check. forcetypeassert finds unchecked type assertions
// that can panic; use a comma-ok assertion or a type switch when failure is
// possible.
package main
