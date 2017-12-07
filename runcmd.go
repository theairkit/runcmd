package runcmd

import (
	"bytes"
	"fmt"
	"io"

	"github.com/reconquest/ser-go"
)

// ExecError represents error messages occured while executing command.
type ExecError struct {
	ExecutionError error
	Args           []string
	Output         []byte
}

// Runner creates command workers.
type Runner interface {
	Command(name string, arg ...string) CmdWorker
}

// CmdWorker executes commands.
type CmdWorker interface {
	Run() error
	Output() ([]byte, []byte, error)
	Start() error
	Wait() error
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	SetStdout(io.Writer)
	SetStderr(io.Writer)
	SetStdin(io.Reader)
	GetArgs() []string
	CmdError() error
}

func (err ExecError) Error() string {
	errString := fmt.Sprintf(
		"%q failed: %s", err.Args, err.ExecutionError,
	)

	if len(err.Output) > 0 {
		errString = errString + ", output: \n" + string(err.Output)
	}

	return errString
}

func run(worker CmdWorker) error {
	err := worker.Start()
	if err != nil {
		return ser.Errorf(
			err, "can't exec %q", worker.GetArgs(),
		)
	}

	err = worker.Wait()
	if err != nil {
		return ExecError{
			ExecutionError: err,
			Args:           worker.GetArgs(),
		}
	}

	return nil
}

func output(worker CmdWorker) ([]byte, []byte, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	worker.SetStdout(&stdout)
	worker.SetStderr(&stderr)

	err := run(worker)
	if err != nil {
		if execErr, ok := err.(ExecError); ok {
			execErr.Output = append(stdout.Bytes(), stderr.Bytes()...)
		}
		return stdout.Bytes(), stderr.Bytes(), err
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}
