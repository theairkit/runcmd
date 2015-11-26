package runcmd

import (
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/mattn/go-shellwords"
)

// LocalCmd is imlementation of CmdWorker interface for local commands
type LocalCmd struct {
	cmdline string
	cmd     *exec.Cmd
}

// Local is implementation of Runner interface for local commands
type Local struct{}

// NewLocalRunner is function for creating local runner
func NewLocalRunner() (*Local, error) {
	return &Local{}, nil
}

// Command method create worker for execution current command
func (runner *Local) Command(cmdline string) (CmdWorker, error) {
	if cmdline == "" {
		return nil, errors.New("command cannot be empty")
	}

	parser := shellwords.NewParser()
	parser.ParseBacktick = false
	parser.ParseEnv = false
	args, err := parser.Parse(cmdline)
	if err != nil {
		return nil, fmt.Errorf("cannot parse cmdline: %s", err.Error())
	}

	command := exec.Command(args[0], args[1:]...)
	return &LocalCmd{
		cmdline: cmdline,
		cmd:     command,
	}, nil
}

// Run method execute current command and retun output splitted by newline
func (cmd *LocalCmd) Run() ([]string, error) {

	return run(cmd)
}

// Start method begin current command execution
func (cmd *LocalCmd) Start() error {
	return cmd.cmd.Start()
}

// Wait method return error after end of command execution if current command
// return nonzero exit code
func (cmd *LocalCmd) Wait() error {
	return cmd.cmd.Wait()
}

// StdinPipe metod return stdin of current worker
func (cmd *LocalCmd) StdinPipe() (io.WriteCloser, error) {
	return cmd.cmd.StdinPipe()
}

// StdoutPipe metod return stdout of current worker
func (cmd *LocalCmd) StdoutPipe() (io.Reader, error) {
	return cmd.cmd.StdoutPipe()
}

// StderrPipe metod return stderr of current worker
func (cmd *LocalCmd) StderrPipe() (io.Reader, error) {
	return cmd.cmd.StderrPipe()
}

// SetStdout is method for binding your own writer to worker stdout
func (cmd *LocalCmd) SetStdout(buffer io.Writer) {
	cmd.cmd.Stdout = buffer
}

// SetStderr is method for binding your own writer to worker stderr
func (cmd *LocalCmd) SetStderr(buffer io.Writer) {
	cmd.cmd.Stderr = buffer
}

// GetCommandLine method return cmdline for current worker
func (cmd *LocalCmd) GetCommandLine() string {
	return cmd.cmdline
}
