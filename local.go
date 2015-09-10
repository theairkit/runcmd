package runcmd

import (
	"errors"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

type LocalCmd struct {
	stdinPipe  io.WriteCloser
	stdoutPipe io.Reader
	stderrPipe io.Reader
	cmd        *exec.Cmd
}

type Local struct {
}

func NewLocalRunner() (*Local, error) {
	return &Local{}, nil
}

func (this *Local) Command(cmd string) (CmdWorker, error) {
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

func (this *LocalCmd) Run() ([]string, error) {
	out := make([]string, 0)
	if err := this.Start(); err != nil {
		return nil, err
	}
	stdout := this.StdoutPipe()
	bOut, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}
	stderr := this.StderrPipe()
	bErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		return nil, err
	}
	if err := this.Wait(); err != nil {
		if len(bErr) > 0 {
			return nil, errors.New(err.Error() + "\n" + string(bErr))
		}
		return nil, err
	}
	if len(bOut) > 0 {
		out = append(out, strings.Split(strings.Trim(string(bOut), "\n"), "\n")...)
	}
	if len(bErr) > 0 {
		out = append(out, strings.Split(strings.Trim(string(bErr), "\n"), "\n")...)
	}
	return out, nil
}

func (this *LocalCmd) Start() error {
	return this.cmd.Start()
}

func (this *LocalCmd) Wait() error {
	cerr := this.StderrPipe()
	bErr, readErr := ioutil.ReadAll(cerr)

	// In this case EOF is not error: http://golang.org/pkg/io/
	// EOF is the error returned by Read when no more input is available.
	// Functions should return EOF only to signal a graceful end of input.
	if err := this.stdinPipe.Close(); err != nil && err != io.EOF {
		return newExecError(err, readErr, bErr)
	}
	if err := this.cmd.Wait(); err != nil {
		return newExecError(err, readErr, bErr)
	}
	return nil
}

func (this *LocalCmd) StdinPipe() io.WriteCloser {
	return this.stdinPipe
}

func (this *LocalCmd) StdoutPipe() io.Reader {
	return this.stdoutPipe
}

func (this *LocalCmd) StderrPipe() io.Reader {
	return this.stderrPipe
}
