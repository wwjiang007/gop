/*
 Copyright 2021 The GoPlus Authors (goplus.org)

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

// Package gengo implements the ``gop go'' command.
package gengo

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/goplus/gop/cmd/gengo"
	"github.com/goplus/gop/cmd/internal/base"
)

// -----------------------------------------------------------------------------

var (
	errTestFailed = errors.New("test failed")
)

func testPkg(p *gengo.Runner, dir string, flags int) error {
	if flags == gengo.PkgFlagGo { // don't test Go packages
		return nil
	}
	cmd1 := exec.Command("go", "run", path.Join(dir, "gop_autogen.go"))
	gorun, err := cmd1.CombinedOutput()
	if err != nil {
		os.Stderr.Write(gorun)
		fmt.Fprintf(os.Stderr, "[ERROR] `%v` failed: %v\n", cmd1, err)
		return err
	}
	cmd2 := exec.Command("gop", "run", "-quiet", dir) // -quiet: don't generate any log
	qrun, err := cmd2.CombinedOutput()
	if err != nil {
		os.Stderr.Write(qrun)
		fmt.Fprintf(os.Stderr, "[ERROR] `%v` failed: %v\n", cmd2, err)
		return err
	}
	if !bytes.Equal(gorun, qrun) {
		fmt.Fprintf(os.Stderr, "[ERROR] Output has differences!\n")
		fmt.Fprintf(os.Stderr, ">>> Output of `%v`:\n", cmd1)
		os.Stderr.Write(gorun)
		fmt.Fprintf(os.Stderr, "\n>>> Output of `%v`:\n", cmd2)
		os.Stderr.Write(qrun)
		return errTestFailed
	}
	return nil
}

// -----------------------------------------------------------------------------

// Cmd - gop go
var Cmd = &base.Command{
	UsageLine: "gop go [-test] <gopSrcDir>",
	Short:     "Convert Go+ packages into Go packages",
}

var (
	flag     = &Cmd.Flag
	flagTest = flag.Bool("test", false, "test Go+ package")
)

func init() {
	Cmd.Run = runCmd
}

func runCmd(cmd *base.Command, args []string) {
	flag.Parse(args)
	if flag.NArg() < 1 {
		cmd.Usage(os.Stderr)
		return
	}
	dir := flag.Arg(0)
	runner := gengo.NewRunner(nil, nil)
	if *flagTest {
		runner.SetAfter(testPkg)
	}
	runner.GenGo(dir, true)
	errs := runner.Errors()
	if errs != nil {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(-1)
	}
}

// -----------------------------------------------------------------------------