package runcmd

import (
	"errors"
	"github.com/mattn/go-shellwords"
	"io"
	"os/exec"
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

	parser := shellwords.NewParser()
	parser.ParseBacktick = false
	parser.ParseEnv = false
	args, err := parser.Parse(cmdline)
	if err != nil {
		return nil, errors.New("cannot parse cmdline")
	}
	command := exec.Command(args[0], args[1:]...)
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
