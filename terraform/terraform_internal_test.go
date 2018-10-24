package terraform

import "os/exec"

func FakeExec(execCmd func(string, ...string) *exec.Cmd) Option {
	return func(c *TerraformCLI) error {
		c.execCmd = execCmd
		return nil
	}
}
