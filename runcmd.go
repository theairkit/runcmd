package runcmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// ExecError is type for generating better error messages
type ExecError struct {
	ExecutionError error
	CommandLine    string
	Output         []string
}

// Runner is interface for creating workers
type Runner interface {
	Command(cmd string) (CmdWorker, error)
}

// CmdWorker is interface for executing commands
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

	if err := worker.Start(); err != nil {
		return nil, err
	}

	err := worker.Wait()
	output := strings.Split(strings.Trim(buffer.String(), "\n"), "\n")

	if err != nil {
		return nil, newExecError(err, worker.GetCommandLine(), output)
	}

	return output, nil
}
