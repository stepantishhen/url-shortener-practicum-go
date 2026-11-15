// Command staticlint runs the project's combined Go static analysis suite.
package main

import "golang.org/x/tools/go/analysis/multichecker"

// main starts the multichecker with the analyzers registered in analyzers.go.
func main() {
	multichecker.Main(analyzers...)
}
