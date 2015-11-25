package runcmd

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type RemoteCmd struct {
	cmdline string
	session *ssh.Session
}

type Remote struct {
	serverConn *ssh.Client
}

type Timeouts struct {
	ConnectionTimeout time.Duration
	SendTimeout       time.Duration
	RecieveTimeout    time.Duration
	KeepAlive         time.Duration
}

type timeBoundedConnection struct {
	net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func (connection *timeBoundedConnection) Read(p []byte) (int, error) {
	err := connection.Conn.SetReadDeadline(time.Now().Add(
		connection.readTimeout,
	))
	if err != nil {
		return 0, err
	}

	return connection.Conn.Read(p)
}

func (connection *timeBoundedConnection) Write(p []byte) (int, error) {
	err := connection.Conn.SetWriteDeadline(time.Now().Add(
		connection.writeTimeout,
	))
	if err != nil {
		return 0, err
	}

	return connection.Conn.Write(p)
}

func NewRemoteKeyAuthRunnerWithTimeouts(
	user, host, key string, timeouts Timeouts,
) (*Remote, error) {
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

	dialer := net.Dialer{
		Timeout:   timeouts.ConnectionTimeout,
		Deadline:  time.Now().Add(timeouts.ConnectionTimeout),
		KeepAlive: timeouts.KeepAlive,
	}

	conn, err := dialer.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	connection := &timeBoundedConnection{
		conn, timeouts.SendTimeout, timeouts.RecieveTimeout,
	}

	sshConnection, channels, requests, err := ssh.NewClientConn(
		connection, host, config,
	)
	if err != nil {
		return nil, err
	}

	return &Remote{ssh.NewClient(sshConnection, channels, requests)}, nil
}

func NewRemotePassAuthRunnerWithTimeouts(
	user, host, password string, timeouts Timeouts,
) (*Remote, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}

	dialer := net.Dialer{
		Timeout:   timeouts.ConnectionTimeout,
		Deadline:  time.Now().Add(timeouts.ConnectionTimeout),
		KeepAlive: timeouts.KeepAlive,
	}

	conn, err := dialer.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	connection := &timeBoundedConnection{
		conn, timeouts.SendTimeout, timeouts.RecieveTimeout,
	}

	sshConnection, channels, requests, err := ssh.NewClientConn(
		connection, host, config,
	)
	if err != nil {
		return nil, err
	}

	return &Remote{ssh.NewClient(sshConnection, channels, requests)}, nil
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
