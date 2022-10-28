package internal

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"github.com/nick-jones/gost/pkg/scan"
)

type Context struct {
	tempDir string
	results []scan.Result
}

func NewContext() *Context {
	return &Context{}
}

func (c *Context) aBinaryBuiltFromSourceFile(fileName string, src *godog.DocString) error {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return err
	}

	c.tempDir, err = ioutil.TempDir("", "gost")
	if err != nil {
		return err
	}

	srcFile := filepath.Join(c.tempDir, fileName)
	if err := ioutil.WriteFile(srcFile, []byte(src.Content), 0600); err != nil {
		return err
	}

	// -gcflags '-l' disables inlining, which gives more reliable file/line information
	cmd := exec.Command(goBin, "build", "-gcflags", "-l", "-o", filepath.Join(c.tempDir, "bin"), srcFile)
	cmd.Env = os.Environ()

	if goOS := os.Getenv("GODOG_GOOS"); goOS != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", goOS))
	}
	if goArch := os.Getenv("GODOG_GOARCH"); goArch != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOARCH=%s", goArch))
	}

	return cmd.Run()
}

func (c *Context) thatBinaryIsAnalysed() error {
	f, err := os.Open(filepath.Join(c.tempDir, "bin"))
	if err != nil {
		return err
	}
	defer f.Close()

	c.results, err = scan.Run(f)
	if err != nil {
		return err
	}

	return nil
}

func (c *Context) theFollowingResultsAreReturned(table *godog.Table) error {
	type summary struct {
		val      string
		fileRefs []string
		symRefs  []string
	}

	expected := make(map[string]summary)
	header := table.Rows[0].Cells
	for _, row := range table.Rows[1:] {
		var s summary
		for i, cell := range row.Cells {
			switch header[i].Value {
			case "String":
				s.val = cell.Value
			case "File References":
				s.fileRefs = strings.Fields(cell.Value)
			case "Symbol References":
				s.symRefs = strings.Fields(cell.Value)
			}
		}
		expected[s.val] = s
	}

	actual := make(map[string]summary)
	for _, res := range c.results {
		s := summary{val: res.Value}
		for _, ref := range res.Refs {
			s.fileRefs = append(s.fileRefs, fmt.Sprintf("%s:%d", filepath.Base(ref.File), ref.Line))
			s.symRefs = append(s.symRefs, ref.SymbolName)
		}
		actual[res.Value] = s
	}

	for _, exp := range expected {
		act, found := actual[exp.val]
		if !found {
			return fmt.Errorf("failed to find string with value %s", exp.val)
		}
		if !equalStringSlice(exp.fileRefs, act.fileRefs) {
			return fmt.Errorf("differing file references for %q, expected %v, actual %v", exp.val, exp.fileRefs, act.fileRefs)
		}
		if !equalStringSlice(exp.symRefs, act.symRefs) {
			return fmt.Errorf("differing symbol references for %q, expected %v, actual %v", exp.val, exp.symRefs, act.symRefs)
		}
	}
	return nil
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (c *Context) RegisterHooks(sc *godog.ScenarioContext) {
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if err == nil && c.tempDir != "" {
			_ = os.RemoveAll(c.tempDir)
		}
		c.tempDir = ""
		c.results = nil
		return ctx, err
	})

	sc.Step(`^a binary built from source file (.+):$`, c.aBinaryBuiltFromSourceFile)
	sc.Step(`^that binary is analysed$`, c.thatBinaryIsAnalysed)
	sc.Step(`^the following results are returned:$`, c.theFollowingResultsAreReturned)
}
