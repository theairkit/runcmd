package runcmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"sync"

	"github.com/reconquest/nopio-go"
)

// MockRunner represents runner which is suitable for writing tests for code
// that uses runcmd library.
type MockRunner struct {
	// Stdout is a bytes slice which will be returned as program's mock stdout.
	Stdout []byte

	// Stdout is a bytes slice which will be returned as program's mock stderr.
	Stderr []byte

	// Error is a error which will be used to fail program's run.
	Error error

	// OnCommand is a callback which will be called on each Command() call with
	// worker which will be similar to exec.Cmd.
	OnCommand func(*MockRunnerWorker)
}

// Command returns standard CmdWorker with streams setup to return mocked
// stdout and stderr.
func (runner MockRunner) Command(name string, args ...string) CmdWorker {
	worker := &MockRunnerWorker{
		MockRunner: runner,
		Args:       append([]string{name}, args...),
	}

	if runner.OnCommand != nil {
		runner.OnCommand(worker)
	}

	return worker
}

type MockRunnerWorker struct {
	MockRunner

	Args []string

	streams struct {
		stdout struct {
			writer io.Writer
			err    error
		}

		stderr struct {
			writer io.Writer
			err    error
		}

		stdin struct {
			reader io.Reader
			err    error
		}

		lock sync.WaitGroup
	}
}

func (worker *MockRunnerWorker) CmdError() error {
	return nil
}

// Run runs Start() and then Wait().
func (worker *MockRunnerWorker) Run() error {
	err := worker.Start()
	if err != nil {
		return err
	}

	return worker.Wait()
}

// Output returns mocked stdout, stderr and execution error.
func (worker *MockRunnerWorker) Output() ([]byte, []byte, error) {
	err := worker.Run()
	if err != nil {
		return nil, nil, err
	}

	return []byte(worker.Stdout), []byte(worker.Stderr), worker.error()
}

// Start will read all incoming stdin (if any), write stdout and stderr to
// specified output streams (if any is set up with SetStdout() or SetStderr()
// methods) and then returns error if any.
func (worker *MockRunnerWorker) Start() error {
	worker.communicate()

	return worker.error()
}

// Wait waits stdin, stdout and stderr streams are processed and returns error
// if any.
func (worker *MockRunnerWorker) Wait() error {
	worker.streams.lock.Wait()

	return worker.error()
}

// StdinPipe returns no-op writer.
func (worker *MockRunnerWorker) StdinPipe() (io.WriteCloser, error) {
	return nopio.NopWriteCloser{}, nil
}

// StdoutPipe returns reader with mocked stdout.
func (worker *MockRunnerWorker) StdoutPipe() (io.Reader, error) {
	return bytes.NewBuffer(worker.Stdout), nil
}

// StderrPipe returns reader with mocked stderr.
func (worker *MockRunnerWorker) StderrPipe() (io.Reader, error) {
	return bytes.NewBuffer(worker.Stderr), nil
}

// SetStdout sets writer which will be used to write mocked stdout.
func (worker *MockRunnerWorker) SetStdout(writer io.Writer) {
	worker.streams.stdout.writer = writer
}

// SetStderr sets writer which will be used to write mocked stderr.
func (worker *MockRunnerWorker) SetStderr(writer io.Writer) {
	worker.streams.stderr.writer = writer
}

// SetStdin sets reader which will be fully read on Start() or Run() till
// Wait().
func (worker *MockRunnerWorker) SetStdin(reader io.Reader) {
	worker.streams.stdin.reader = reader
}

// GetArgs returns original command line arguments.
func (worker *MockRunnerWorker) GetArgs() []string {
	return worker.Args
}

func (worker *MockRunnerWorker) communicate() {
	if worker.streams.stdin.reader != nil {
		worker.stream(func() {
			_, worker.streams.stdin.err = ioutil.ReadAll(
				worker.streams.stdin.reader,
			)
		})
	}

	if worker.streams.stdout.writer != nil {
		worker.stream(func() {
			_, worker.streams.stdout.err = worker.streams.stdout.writer.Write(
				worker.Stdout,
			)
		})
	}

	if worker.streams.stderr.writer != nil {
		worker.stream(func() {
			_, worker.streams.stderr.err = worker.streams.stderr.writer.Write(
				worker.Stderr,
			)
		})
	}
}

func (worker *MockRunnerWorker) stream(body func()) {
	defer worker.streams.lock.Add(1)
	go func() {
		defer worker.streams.lock.Done()

		body()
	}()
}

func (worker *MockRunnerWorker) error() error {
	if worker.streams.stdin.err != nil {
		return worker.streams.stdin.err
	}

	if worker.streams.stdout.err != nil {
		return worker.streams.stdout.err
	}

	if worker.streams.stdin.err != nil {
		return worker.streams.stdin.err
	}

	return worker.Error
}
