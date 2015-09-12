package runcmd

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

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

	session, err := runner.serverConn.NewSession()
	if err != nil {
		return nil, err
	}

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}

	return &RemoteCmd{
		stdinPipe,
		stdoutPipe,
		stderrPipe,
		cmd,
		session,
	}, nil
}

func (runner *Remote) CloseConnection() error {
	return runner.serverConn.Close()
}

func (cmd *RemoteCmd) Run() ([]string, error) {
	defer cmd.session.Close()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return run(cmd)
}

func (cmd *RemoteCmd) Start() error {
	return cmd.session.Start(cmd.cmd)
}

func (cmd *RemoteCmd) Wait() error {
	defer cmd.session.Close()

	// In cmd case EOF is not error: http://golang.org/pkg/io/
	// EOF is the error returned by Read when no more input is available.
	// Functions should return EOF only to signal a graceful end of input.
	if err := cmd.stdinPipe.Close(); err != nil && err != io.EOF {
		return err
	}

	return cmd.session.Wait()
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

func (cmd *RemoteCmd) SetStdout(buffer io.Writer) {
	cmd.session.Stdout = buffer
}

func (cmd *RemoteCmd) SetStderr(buffer io.Writer) {
	cmd.session.Stderr = buffer
}
