// Static analytic service. Include static analytic packages:
// - golang.org/x/tools/go/analysis/passes
// - all SA classes staticcheck.io
// - Go-critic and nilerr linters
// - OsExitAnalyzer to check os.Exit calls in main.go main package
//
// How to run:
//
// ./cmd/staticlint/main ./...

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gocritic "github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/gostaticanalysis/nilerr"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	localanalyzer "github.com/paramonies/pkg/analyzer"
)

// Config  name config
const Config = `config/config.json`

// ConfigData describe config file
type ConfigData struct {
	Staticcheck []string
	Stylecheck  []string
}

var allPassesChecks = []*analysis.Analyzer{
	asmdecl.Analyzer,             // reports mismatches between assembly files and Go declarations
	assign.Analyzer,              // detects useless assignments
	atomic.Analyzer,              // checks for common mistakes using the sync/atomic package
	atomicalign.Analyzer,         // checks for non-64-bit-aligned arguments to sync/atomic functions
	bools.Analyzer,               // detects common mistakes involving boolean operators
	buildssa.Analyzer,            // constructs the SSA representation of an error-free package and returns the set of all functions within it.
	buildtag.Analyzer,            // checks build tags
	cgocall.Analyzer,             // detects some violations of the cgo pointer passing rules
	composite.Analyzer,           // checks for unkeyed composite literals
	copylock.Analyzer,            // checks for locks erroneously passed by value
	ctrlflow.Analyzer,            // provides a syntactic control-flow graph (CFG) for the body of a function
	deepequalerrors.Analyzer,     // checks for the use of reflect.DeepEqual with error values
	errorsas.Analyzer,            // checks that the second argument to errors.As is a pointer to a type implementing error
	fieldalignment.Analyzer,      // detects structs that would use less memory if their fields were sorted
	findcall.Analyzer,            // serves as a trivial example and test of the Analysis API
	framepointer.Analyzer,        // reports assembly code that clobbers the frame pointer before saving it
	httpresponse.Analyzer,        // checks for mistakes using HTTP responses
	ifaceassert.Analyzer,         // flags impossible interface-interface type assertions
	inspect.Analyzer,             // provides an AST inspector (golang.org/x/tools/go/ast/inspector.Inspector) for the syntax trees of a package
	loopclosure.Analyzer,         // checks for references to enclosing loop variables from within nested functions
	lostcancel.Analyzer,          // checks for failure to call a context cancellation function
	nilfunc.Analyzer,             // checks for useless comparisons against nil
	nilness.Analyzer,             // inspects the control-flow graph of an SSA function and reports errors such as nil pointer dereferences and degenerate nil pointer comparisons.
	pkgfact.Analyzer,             // demonstration and test of the package fact mechanism
	printf.Analyzer,              // checks consistency of Printf format strings and arguments
	reflectvaluecompare.Analyzer, // checks for accidentally using == or reflect.DeepEqual to compare reflect.Value values
	shadow.Analyzer,              // checks for shadowed variables
	shift.Analyzer,               // checks for shifts that exceed the width of an integer
	sigchanyzer.Analyzer,         // detects misuse of unbuffered signal as argument to signal.Notify
	sortslice.Analyzer,           // checks for calls to sort.Slice that do not use a slice type as first argument
	stdmethods.Analyzer,          // checks for misspellings in the signatures of methods similar to well-known interfaces
	stringintconv.Analyzer,       // flags type conversions from integers to strings
	structtag.Analyzer,           // checks struct field tags are well formed
	tests.Analyzer,               // checks for common mistaken usages of tests and examples
	unmarshal.Analyzer,           // checks for passing non-pointer or non-interface types to unmarshal and decode functions
	unreachable.Analyzer,         // checks for unreachable code
	unsafeptr.Analyzer,           // checks for invalid conversions of uintptr to unsafe.Pointer
	unusedresult.Analyzer,        // checks for unused results of calls to certain pure functions
	unusedwrite.Analyzer,         // checks for unused writes to the elements of a struct or array object
	usesgenerics.Analyzer,        // checks for usage of generic features
}

var publicChecks = []*analysis.Analyzer{
	gocritic.Analyzer, // Go-critic
	nilerr.Analyzer,   // nilerr
}

func main() {
	var allChecks []*analysis.Analyzer
	allChecks = append(allChecks, allPassesChecks...)
	allChecks = append(allChecks, publicChecks...)
	allChecks = append(allChecks, localanalyzer.OsExitAnalyzer)

	appfile, err := os.Executable()
	if err != nil {
		panic(err)
	}

	fmt.Printf("!!! %s %s \n", appfile, filepath.Dir(appfile))

	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		panic(err)
	}

	var cfg ConfigData
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	for _, v := range staticcheck.Analyzers {
		for _, sc := range cfg.Staticcheck {
			if strings.HasPrefix(v.Name, sc) {
				allChecks = append(allChecks, v)
			}
		}
	}

	for _, v := range stylecheck.Analyzers {
		for _, sc := range cfg.Stylecheck {
			if strings.HasPrefix(v.Name, sc) {
				allChecks = append(allChecks, v)
			}
		}
	}

	fmt.Println("Run all checks")
	for _, c := range allChecks {
		fmt.Printf("%s\n", c.Name)
	}

	multichecker.Main(
		allChecks...,
	)
}
