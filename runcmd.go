package runcmd

import (
	"bytes"
	"io"
	"strings"
)

type ExecError struct {
	ExecutionError error
	Output         []string
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
	SetStdout(buffer io.Writer)
	SetStderr(buffer io.Writer)
}

func newExecError(
	execErr error, output []string,
) ExecError {
	return ExecError{execErr, output}
}

func (err ExecError) Error() string {
	errString := err.ExecutionError.Error()
	errString = errString + "\n" + strings.Join(err.Output, "\n")

	return errString
}

func run(worker CmdWorker) ([]string, error) {
	var buffer bytes.Buffer

	worker.SetStdout(&buffer)
	worker.SetStderr(&buffer)

	err := worker.Wait()
	output := strings.Split(string(buffer.Bytes()), "\n")

	if err != nil {
		return nil, newExecError(err, output)
	}

	return output, nil
}
