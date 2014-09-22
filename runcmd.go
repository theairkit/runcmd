package runcmd

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"code.google.com/p/go.crypto/ssh"
)

type Runner interface {
	Command(cmd string) (CmdWorker, error)
}

type CmdWorker interface {
	Run() ([]string, error)
	Start() error
	Wait() error
	StdinPipe() io.WriteCloser
	StdoutPipe() io.Reader
	StderrPipe() io.Reader
}

type LocalCmd struct {
	stdinPipe  io.WriteCloser
	stdoutPipe io.Reader
	stderrPipe io.Reader
	cmd        *exec.Cmd
}

type RemoteCmd struct {
	stdinPipe  io.WriteCloser
	stdoutPipe io.Reader
	stderrPipe io.Reader
	cmd        string
	session    *ssh.Session
}

type Local struct {
}

type Remote struct {
	serverConn *ssh.Client
}

func (this Local) Command(cmd string) (CmdWorker, error) {
	c := strings.Split(cmd, " ")
	command := exec.Command(c[0], c[1:]...)
	stdinPipe, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := command.StderrPipe()
	if err != nil {
		return nil, err
	}
	return &LocalCmd{
		stdinPipe,
		stdoutPipe,
		stderrPipe,
		command,
	}, nil
}

func (this Remote) Command(cmd string) (CmdWorker, error) {
	session, err := this.serverConn.NewSession()
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

func (this *LocalCmd) Run() ([]string, error) {
	if err := this.Start(); err != nil {
		return nil, err
	}
	cout := this.StdoutPipe()
	bOut, err := ioutil.ReadAll(cout)
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
		return strings.Split(strings.Trim(string(bOut), "\n"), "\n"), nil
	}
	return nil, nil
}

func (this *LocalCmd) Start() error {
	return this.cmd.Start()
}

func (this *LocalCmd) Wait() error {
	cerr := this.StderrPipe()
	bErr, err := ioutil.ReadAll(cerr)
	if err != nil {
		return err
	}
	if err := this.cmd.Wait(); err != nil {
		if len(bErr) > 0 {
			return errors.New(err.Error() + "\n" + string(bErr))
		}
		return err
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

func (this *RemoteCmd) Run() ([]string, error) {
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
		return strings.Split(strings.Trim(string(bOut), "\n"), "\n"), nil
	}
	return nil, nil
}

func (this *RemoteCmd) Start() error {
	return this.session.Start(this.cmd)
}

func (this *RemoteCmd) Wait() error {
	defer this.session.Close()
	cerr := this.StderrPipe()
	bErr, err := ioutil.ReadAll(cerr)
	if err != nil {
		return err
	}
	if err := this.session.Wait(); err != nil {
		if len(bErr) > 0 {
			return errors.New(err.Error() + "\n" + string(bErr))
		}
		return err
	}
	return nil
}

func (this *RemoteCmd) StdinPipe() io.WriteCloser {
	return this.stdinPipe
}

func (this *RemoteCmd) StdoutPipe() io.Reader {
	return this.stdoutPipe
}

func (this *RemoteCmd) StderrPipe() io.Reader {
	return this.stderrPipe
}

func NewLocalRunner() (*Local, error) {
	return &Local{}, nil
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
