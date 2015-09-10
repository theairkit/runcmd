package runcmd

import (
	"errors"
	"io"
	"os/exec"
	"strings"
)

type LocalCmd struct {
	stdinPipe  io.WriteCloser
	stdoutPipe io.Reader
	stderrPipe io.Reader
	cmd        *exec.Cmd
}

type Local struct{}

func NewLocalRunner() (*Local, error) {
	return &Local{}, nil
}

func (runner *Local) Command(cmd string) (CmdWorker, error) {
	if cmd == "" {
		return nil, errors.New("command cannot be empty")
	}
	c := exec.Command(strings.Fields(cmd)[0], strings.Fields(cmd)[1:]...)
	stdinPipe, err := c.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := c.StderrPipe()
	if err != nil {
		return nil, err
	}
	return &LocalCmd{
		stdinPipe,
		stdoutPipe,
		stderrPipe,
		c,
	}, nil
}

func (cmd *LocalCmd) Run() ([]string, error) {
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return run(cmd)
}

func (cmd *LocalCmd) Start() error {
	return cmd.cmd.Start()
}

func (cmd *LocalCmd) Wait() error {
	if err := cmd.stdinPipe.Close(); err != nil && err != io.EOF {
		return err
	}

	return cmd.cmd.Wait()
}

func (cmd *LocalCmd) StdinPipe() io.WriteCloser {
	return cmd.stdinPipe
}

func (cmd *LocalCmd) StdoutPipe() io.Reader {
	return cmd.stdoutPipe
}

func (cmd *LocalCmd) SetStdout(buffer io.Writer) {
	cmd.cmd.Stdout = buffer
}

func (cmd *LocalCmd) SetStderr(buffer io.Writer) {
	cmd.cmd.Stderr = buffer
}

func (cmd *LocalCmd) StderrPipe() io.Reader {
	return cmd.stderrPipe
}
