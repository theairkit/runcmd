package runcmd

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"code.google.com/p/go.crypto/ssh"
)

type RemoteCmd struct {
	stdinPipe  io.WriteCloser
	stdoutPipe io.Reader
	stderrPipe io.Reader
	cmd        string
	session    *ssh.Session
}

type Remote struct {
	serverConn *ssh.Client
}

func NewRemoteKeyAuthRunner(user, host, key string) (*Remote, error) {
	if _, err := os.Stat(key); os.IsNotExist(err) {
		return nil, err
	}
	bs, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(bs)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}
	server, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, err
	}
	return &Remote{server}, nil
}

func NewRemotePassAuthRunner(user, host, password string) (*Remote, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}
	server, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, err
	}
	return &Remote{server}, nil
}

func (runner *Remote) Command(cmd string) (CmdWorker, error) {
	if cmd == "" {
		return nil, errors.New("command cannot be empty")
	}
	s, err := runner.serverConn.NewSession()
	if err != nil {
		return nil, err
	}
	stdinPipe, err := s.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdoutPipe, err := s.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := s.StderrPipe()
	if err != nil {
		return nil, err
	}
	return &RemoteCmd{
		stdinPipe,
		stdoutPipe,
		stderrPipe,
		cmd,
		s,
	}, nil
}

func (runner *Remote) CloseConnection() error {
	return runner.serverConn.Close()
}

func (cmd *RemoteCmd) Run() ([]string, error) {
	defer cmd.session.Close()
	var out []string
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	stdout := cmd.StdoutPipe()
	bOut, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}
	stderr := cmd.StderrPipe()
	bErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
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

func (cmd *RemoteCmd) Start() error {
	return cmd.session.Start(cmd.cmd)
}

func (cmd *RemoteCmd) Wait() error {
	defer cmd.session.Close()
	cerr := cmd.StderrPipe()
	bErr, readErr := ioutil.ReadAll(cerr)

	// In cmd case EOF is not error: http://golang.org/pkg/io/
	// EOF is the error returned by Read when no more input is available.
	// Functions should return EOF only to signal a graceful end of input.
	if err := cmd.stdinPipe.Close(); err != nil && err != io.EOF {
		return newExecError(err, readErr, bErr)
	}
	if err := cmd.session.Wait(); err != nil {
		return newExecError(err, readErr, bErr)
	}
	return nil
}

func (cmd *RemoteCmd) StdinPipe() io.WriteCloser {
	return cmd.stdinPipe
}

func (cmd *RemoteCmd) StdoutPipe() io.Reader {
	return cmd.stdoutPipe
}

func (cmd *RemoteCmd) StderrPipe() io.Reader {
	return cmd.stderrPipe
}
