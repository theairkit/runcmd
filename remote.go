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
	cmdline    string
	session    *ssh.Session
	connection *timeBoundedConnection
	timeouts   *Timeouts
}

// Remote is implementation of Runner interface for remote commands
type Remote struct {
	serverConn *ssh.Client
	connection *timeBoundedConnection
	timeouts   *Timeouts
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
	if connection.readTimeout != 0 {
		err := connection.Conn.SetReadDeadline(time.Now().Add(
			connection.readTimeout,
		))
		if err != nil {
			return 0, err
		}
	}

	return connection.Conn.Read(p)
}

func (connection *timeBoundedConnection) Write(p []byte) (int, error) {
	if connection.writeTimeout != 0 {
		err := connection.Conn.SetWriteDeadline(time.Now().Add(
			connection.writeTimeout,
		))
		if err != nil {
			return 0, err
		}
	}

	return connection.Conn.Write(p)
}

// NewRemoteRawKeyAuthRunnerWithTimeouts is same, as NewRemoteKeyAuthRunnerWithTimeouts,
// but key should be raw byte sequence instead of path.
func NewRemoteRawKeyAuthRunnerWithTimeouts(
	user, host, key string, timeouts Timeouts,
) (*Remote, error) {
	signer, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		return nil, errors.New("can't parse PEM data: " + err.Error())
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}

	dialer := net.Dialer{
		Timeout:   timeouts.ConnectionTimeout,
		KeepAlive: timeouts.KeepAlive,
	}

	if timeouts.ConnectionTimeout != 0 {
		dialer.Deadline = time.Now().Add(timeouts.ConnectionTimeout)
	}

	conn, err := dialer.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	connection := &timeBoundedConnection{
		Conn: conn,
	}

	sshConnection, channels, requests, err := ssh.NewClientConn(
		connection, host, config,
	)
	if err != nil {
		return nil, err
	}

	return &Remote{
		serverConn: ssh.NewClient(sshConnection, channels, requests),
		connection: connection,
		timeouts:   &timeouts,
	}, nil
}

// NewRemoteKeyAuthRunnerWithTimeouts is one of functions for creating
// remote runner. Use this one instead of NewRemoteKeyAuthRunner if you need to
// setup nondefault timeouts for ssh connection
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

	return NewRemoteRawKeyAuthRunnerWithTimeouts(
		user, host, string(pemBytes), timeouts,
	)
}

// NewRemotePassAuthRunnerWithTimeouts is one of functions for creating remote
// runner. Use this one instead of NewRemotePassAuthRunner if you need to setup
// nondefault timeouts for ssh connection
func NewRemotePassAuthRunnerWithTimeouts(
	user, host, password string, timeouts Timeouts,
) (*Remote, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}

	dialer := net.Dialer{
		Timeout:   timeouts.ConnectionTimeout,
		KeepAlive: timeouts.KeepAlive,
	}

	if timeouts.ConnectionTimeout != 0 {
		dialer.Deadline = time.Now().Add(timeouts.ConnectionTimeout)
	}

	conn, err := dialer.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	connection := &timeBoundedConnection{
		Conn: conn,
	}

	sshConnection, channels, requests, err := ssh.NewClientConn(
		connection, host, config,
	)
	if err != nil {
		return nil, err
	}

	return &Remote{
		serverConn: ssh.NewClient(sshConnection, channels, requests),
		connection: connection,
		timeouts:   &timeouts,
	}, nil
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

	return &Remote{
		serverConn: server,
		connection: nil,
		timeouts:   nil,
	}, nil
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

	return &Remote{
		serverConn: server,
		connection: nil,
		timeouts:   nil,
	}, nil
}

// Command creates worker for current command execution
func (runner *Remote) Command(cmdline string) (CmdWorker, error) {
	if cmdline == "" {
		return nil, errors.New("command cannot be empty")
	}

	session, err := runner.serverConn.NewSession()
	if err != nil {
		return nil, err
	}

	return &RemoteCmd{
		cmdline:    cmdline,
		session:    session,
		connection: runner.connection,
		timeouts:   runner.timeouts,
	}, nil
}

// CloseConnection is method for closing ssh connection of current runner
func (runner *Remote) CloseConnection() error {
	return runner.serverConn.Close()
}

// Run executes current command and returns output splitted by newline
func (cmd *RemoteCmd) Run() (result []string, err error) {
	defer func() {
		closeErr := cmd.session.Close()
		if err == nil && closeErr != nil {
			if closeErr.Error() != "EOF" {
				err = errors.New("can't close ssh session: " + closeErr.Error())
			}
		}
	}()

	cmd.initTimeouts()

	return run(cmd)
}

// Start begins current command execution
func (cmd *RemoteCmd) Start() error {
	cmd.initTimeouts()

	return cmd.session.Start(cmd.cmdline)
}

// Wait returns error after command execution if current command return nonzero
// exit code
func (cmd *RemoteCmd) Wait() (err error) {
	defer func() {
		closeErr := cmd.session.Close()
		if err == nil && closeErr != nil {
			if closeErr.Error() != "EOF" {
				err = errors.New("can't close ssh session: " + closeErr.Error())
			}
		}
	}()

	return cmd.session.Wait()
}

// StdinPipe returns stdin of current worker
func (cmd *RemoteCmd) StdinPipe() (io.WriteCloser, error) {
	return cmd.session.StdinPipe()
}

// StdoutPipe returns stdout of current worker
func (cmd *RemoteCmd) StdoutPipe() (io.Reader, error) {
	return cmd.session.StdoutPipe()
}

// StderrPipe returns stderr of current worker
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

// GetCommandLine returns cmdline for current worker
func (cmd *RemoteCmd) GetCommandLine() string {
	return cmd.cmdline
}

func (cmd *RemoteCmd) initTimeouts() {
	if cmd.connection == nil {
		return
	}
	cmd.connection.readTimeout = cmd.timeouts.SendTimeout
	cmd.connection.writeTimeout = cmd.timeouts.ReceiveTimeout
}
