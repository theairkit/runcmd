package runcmd

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

	"code.google.com/p/go.crypto/ssh"
)

type RemoteCmd struct {
	cmdline string
	session *ssh.Session
}

type Remote struct {
	serverConn *ssh.Client
}

func NewRemoteKeyAuthRunner(user, host, key string) (*Remote, error) {
	if _, err := os.Stat(key); os.IsNotExist(err) {
		return nil, err
	}
	pemBytes, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
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

func (runner *Remote) Command(cmdline string) (CmdWorker, error) {
	if cmdline == "" {
		return nil, errors.New("command cannot be empty")
	}

	session, err := runner.serverConn.NewSession()
	if err != nil {
		return nil, err
	}

	return &RemoteCmd{
		cmdline: cmdline,
		session: session,
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
	return cmd.session.Start(cmd.cmdline)
}

func (cmd *RemoteCmd) Wait() error {
	defer cmd.session.Close()

	return cmd.session.Wait()
}

func (cmd *RemoteCmd) StdinPipe() (io.WriteCloser, error) {
	return cmd.session.StdinPipe()
}

func (cmd *RemoteCmd) StdoutPipe() (io.Reader, error) {
	return cmd.session.StdoutPipe()
}

func (cmd *RemoteCmd) StderrPipe() (io.Reader, error) {
	return cmd.session.StderrPipe()
}

func (cmd *RemoteCmd) SetStdout(buffer io.Writer) {
	cmd.session.Stdout = buffer
}

func (cmd *RemoteCmd) SetStderr(buffer io.Writer) {
	cmd.session.Stderr = buffer
}

func (cmd *RemoteCmd) GetCommandLine() string {
	return cmd.cmdline
}
