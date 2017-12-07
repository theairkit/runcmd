package runcmd

import (
	"io"
	"os/exec"
)

// LocalCmd is implementation of CmdWorker interface for local commands
type LocalCmd struct {
	cmd *exec.Cmd
}

// Local is implementation of Runner interface for local commands
type Local struct{}

// NewLocalRunner creates instance of local runner
func NewLocalRunner() (*Local, error) {
	return &Local{}, nil
}

// Command creates worker for current command execution
func (local *Local) Command(name string, arg ...string) CmdWorker {
	return &LocalCmd{
		cmd: exec.Command(name, arg...),
	}
}

// stub to suite interface
func (local *LocalCmd) CmdError() error {
	return nil
}

// Run executes current command and retuns output splitted by newline
func (cmd *LocalCmd) Run() error {
	return run(cmd)
}

func (cmd *LocalCmd) Output() ([]byte, []byte, error) {
	return output(cmd)
}

// Start begins current command execution
func (cmd *LocalCmd) Start() error {
	return cmd.cmd.Start()
}

// Wait returns error after command execution if current command return nonzero
// exit code
func (cmd *LocalCmd) Wait() error {
	return cmd.cmd.Wait()
}

// StdinPipe returns stdin of current worker
func (cmd *LocalCmd) StdinPipe() (io.WriteCloser, error) {
	return cmd.cmd.StdinPipe()
}

// StdoutPipe returns stdout of current worker
func (cmd *LocalCmd) StdoutPipe() (io.Reader, error) {
	return cmd.cmd.StdoutPipe()
}

// StderrPipe returns stderr of current worker
func (cmd *LocalCmd) StderrPipe() (io.Reader, error) {
	return cmd.cmd.StderrPipe()
}

// SetStdout is for binding your own writer to worker stdout
func (cmd *LocalCmd) SetStdout(buffer io.Writer) {
	cmd.cmd.Stdout = buffer
}

// SetStderr is for binding your own writer to worker stderr
func (cmd *LocalCmd) SetStderr(buffer io.Writer) {
	cmd.cmd.Stderr = buffer
}

func (cmd *LocalCmd) SetStdin(reader io.Reader) {
	cmd.cmd.Stdin = reader
}

// GetArgs returns cmdline for current worker
func (cmd *LocalCmd) GetArgs() []string {
	return cmd.cmd.Args
}
