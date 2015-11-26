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

// RemoteCmd is implementation of CmdWorker interface for remote commands
type RemoteCmd struct {
	cmdline string
	session *ssh.Session
}

// Remote is implementation of Runner interface for remote commands
type Remote struct {
	serverConn *ssh.Client
}

// Timeouts is struct for setting various timeouts for ssh connection
type Timeouts struct {
	ConnectionTimeout time.Duration
	SendTimeout       time.Duration
	ReceiveTimeout    time.Duration
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

// NewRemoteKeyAuthRunnerWithTimeouts is one of functions for creating remote
// runner
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
		return nil, errors.New("can't parse pem data: " + err.Error())
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
		Conn:         conn,
		readTimeout:  timeouts.SendTimeout,
		writeTimeout: timeouts.ReceiveTimeout,
	}

	sshConnection, channels, requests, err := ssh.NewClientConn(
		connection, host, config,
	)
	if err != nil {
		return nil, err
	}

	return &Remote{ssh.NewClient(sshConnection, channels, requests)}, nil
}

// NewRemotePassAuthRunnerWithTimeouts is one of functions for creating remote
// runner
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
		Conn:         conn,
		readTimeout:  timeouts.SendTimeout,
		writeTimeout: timeouts.ReceiveTimeout,
	}

	sshConnection, channels, requests, err := ssh.NewClientConn(
		connection, host, config,
	)
	if err != nil {
		return nil, err
	}

	return &Remote{ssh.NewClient(sshConnection, channels, requests)}, nil
}

// NewRemoteKeyAuthRunner is one of functions for creating remote runner
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

// NewRemotePassAuthRunner is one of functions for creating remote runner
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

// Command method create worker for execution current command
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

// CloseConnection is method for closing ssh connection of current runner
func (runner *Remote) CloseConnection() error {
	return runner.serverConn.Close()
}

// Run method execute current command and retun output splitted by newline
func (cmd *RemoteCmd) Run() (result []string, err error) {
	defer func() {
		closeErr := cmd.session.Close()
		if err == nil {
			err = errors.New("can't close ssh session: " + closeErr.Error())
		}
	}()

	return run(cmd)
}

// Start method begin current command execution
func (cmd *RemoteCmd) Start() error {
	return cmd.session.Start(cmd.cmdline)
}

// Wait method return error after end of command execution if current command
// return nonzero exit code
func (cmd *RemoteCmd) Wait() (err error) {
	defer func() {
		closeErr := cmd.session.Close()
		if err == nil {
			err = errors.New("can't close ssh session: " + closeErr.Error())
		}
	}()

	return cmd.session.Wait()
}

// StdinPipe metod return stdin of current worker
func (cmd *RemoteCmd) StdinPipe() (io.WriteCloser, error) {
	return cmd.session.StdinPipe()
}

// StdoutPipe metod return stdout of current worker
func (cmd *RemoteCmd) StdoutPipe() (io.Reader, error) {
	return cmd.session.StdoutPipe()
}

// StderrPipe metod return stderr of current worker
func (cmd *RemoteCmd) StderrPipe() (io.Reader, error) {
	return cmd.session.StderrPipe()
}

// SetStdout is method for binding your own writer to worker stdout
func (cmd *RemoteCmd) SetStdout(buffer io.Writer) {
	cmd.session.Stdout = buffer
}

// SetStderr is method for binding your own writer to worker stderr
func (cmd *RemoteCmd) SetStderr(buffer io.Writer) {
	cmd.session.Stderr = buffer
}

// GetCommandLine method return cmdline for current worker
func (cmd *RemoteCmd) GetCommandLine() string {
	return cmd.cmdline
}
