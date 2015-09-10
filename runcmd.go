package runcmd

import (
	"fmt"
	"io"
)

type ExecError struct {
	ExecutionError  error
	StderrReadError error

	Stderr []byte
}

type Runner interface {
	Command(cmd string) (CmdWorker, error)
}

type CmdWorker interface {
	Run() ([]string, error)
	Start() error
	Wait() error
	StdinPipe() io.WriteCloser
	StdoutPipe() io.Reader
	StderrPipe() io.Reader
}

func newExecError(execErr, readErr error, stderr []byte) ExecError {
	return ExecError{execErr, readErr, stderr}
}

func (err ExecError) Error() string {
	errString := err.ExecutionError.Error()
	if err.StderrReadError != nil {
		errString += fmt.Sprintf(
			"\nerror while reading stderr: %s", err.StderrReadError.Error(),
		)
	}
	if len(err.Stderr) != 0 {
		errString += fmt.Sprintf(
			"\n%s", string(err.Stderr),
		)
	}
	return errString
}
