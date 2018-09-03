/*
Package fakeexec is used to mock calls to exec.Command.
The following function must be included in a *_test.go file.
	func TestExecCommandHelper(t *testing.T) {
		if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
			return
		}
		fmt.Print(os.Getenv("STDOUT"))
		i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
		os.Exit(i)
	}
*/
package fakeexec

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Expect controls the expected execution of a command
type Expect struct {
	f        func(t testing.TB, actualCommand string, actualArgs ...string)
	exitCode int
	stdout   string
}

// E represents a set of expected executions of a command
type E struct {
	t  testing.TB
	es []*Expect
}

// New create a new E
func New(t testing.TB) *E {
	e := new(E)
	e.t = t
	return e
}

// Exits set the exit code of the execution
func (e *Expect) Exits(code int) {
	e.exitCode = code
}

// Outputs sets the executions standard output
func (e *Expect) Outputs(stdout string) {
	e.stdout = stdout
}

// Expect is adds a new expectation
func (e *E) Expect(command string, args ...string) *Expect {
	return e.ExpectFunc(func(t testing.TB, actualCommand string, actualArgs ...string) {
		exp := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
		act := fmt.Sprintf("%s %s", actualCommand, strings.Join(actualArgs, " "))
		require.Equal(t, exp, act)
	})
}

// ExpectFunc adds a new expectation, command and args can be checked for correctness by the supplied function
func (e *E) ExpectFunc(f func(t testing.TB, command string, args ...string)) *Expect {
	ex := &Expect{
		f: f,
	}
	e.es = append(e.es, ex)
	return ex
}

// Cmd returns a function with a signiture that matches (os/exec).Command
func (e *E) Cmd() func(command string, args ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		e.t.Helper()
		if len(e.es) == 0 {
			e.t.Fatalf("unexpected call to %s %v", command, args)
		}
		thisProg, err := os.Executable()
		if err != nil {
			panic(err)
		}
		var expectation *Expect
		expectation, e.es = e.es[0], e.es[1:]
		expectation.f(e.t, command, args...)
		cs := []string{"-test.run=TestExecCommandHelper", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(thisProg, cs...)
		es := strconv.Itoa(expectation.exitCode)
		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			"EXIT_STATUS=" + es,
			"STDOUT=" + expectation.stdout,
		}
		return cmd
	}
}

// Finish checks all expectations have been met
func (e *E) Finish() {
	e.t.Helper()
	if len(e.es) != 0 {
		e.t.Fatalf("%d missing calls to os.Exec", len(e.es))
	}
}
