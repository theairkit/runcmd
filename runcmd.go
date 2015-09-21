package runcmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type ExecError struct {
	ExecutionError error
	CommandLine    string
	Output         []string
}

type Runner interface {
	Command(cmd string) (CmdWorker, error)
}

type CmdWorker interface {
	Run() ([]string, error)
	Start() error
	Wait() error
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	SetStdout(buffer io.Writer)
	SetStderr(buffer io.Writer)
	GetCommandLine() string
}

func newExecError(
	execErr error, cmdline string, output []string,
) ExecError {
	return ExecError{execErr, cmdline, output}
}

func (err ExecError) Error() string {
	errString := fmt.Sprintf(
		"`%s` failed: %s", err.CommandLine, err.ExecutionError,
	)

	output := strings.Join(err.Output, "\n")
	if strings.TrimSpace(output) != "" {
		errString = errString + ", output: \n" + output
	}

	return errString
}

func run(worker CmdWorker) ([]string, error) {
	var buffer bytes.Buffer

	worker.SetStdout(&buffer)
	worker.SetStderr(&buffer)

	err := worker.Wait()
	output := strings.Split(string(buffer.Bytes()), "\n")

	if err != nil {
		return nil, newExecError(err, worker.GetCommandLine(), output)
	}

	return output, nil
}
