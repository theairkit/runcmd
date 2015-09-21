package runcmd

import (
	"errors"
	"io"
	"os/exec"
	"strings"
)

type LocalCmd struct {
	cmdline string
	cmd     *exec.Cmd
}

type Local struct{}

func NewLocalRunner() (*Local, error) {
	return &Local{}, nil
}

func (runner *Local) Command(cmdline string) (CmdWorker, error) {
	if cmdline == "" {
		return nil, errors.New("command cannot be empty")
	}

	command := exec.Command(strings.Fields(cmdline)[0], strings.Fields(cmdline)[1:]...)
	return &LocalCmd{
		cmdline: cmdline,
		cmd:     command,
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
	return cmd.cmd.Wait()
}

func (cmd *LocalCmd) StdinPipe() (io.WriteCloser, error) {
	return cmd.cmd.StdinPipe()
}

func (cmd *LocalCmd) StdoutPipe() (io.Reader, error) {
	return cmd.cmd.StdoutPipe()
}

func (cmd *LocalCmd) StderrPipe() (io.Reader, error) {
	return cmd.cmd.StderrPipe()
}

func (cmd *LocalCmd) SetStdout(buffer io.Writer) {
	cmd.cmd.Stdout = buffer
}

func (cmd *LocalCmd) SetStderr(buffer io.Writer) {
	cmd.cmd.Stderr = buffer
}

func (cmd *LocalCmd) GetCommandLine() string {
	return cmd.cmdline
}
