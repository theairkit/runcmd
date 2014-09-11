package runcmd

import (
	"errors"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"

	"code.google.com/p/go.crypto/ssh"
)

type Runner interface {
	Command(cmd string) CmdWorker
}

type CmdWorker interface {
	Run() ([]string, error)
	Start() error
	Wait() error
	Stdin() io.WriteCloser
	Stdout() io.Reader
	Stderr() io.Reader
}

type LocalCmd struct {
	StdinPipe  io.WriteCloser
	StdoutPipe io.Reader
	StderrPipe io.Reader
	cmd        *exec.Cmd
}

type RemoteCmd struct {
	StdinPipe  io.WriteCloser
	StdoutPipe io.Reader
	StderrPipe io.Reader
	cmd        string
	session    *ssh.Session
}

type Local struct {
}

type Remote struct {
	serverConn *ssh.Client
}

func (this Local) Command(cmd string) CmdWorker {
	c := strings.Split(cmd, " ")
	return &LocalCmd{nil, nil, nil, exec.Command(c[0], c[1:]...)}
}

func (this Remote) Command(cmd string) CmdWorker {
	session, err := this.serverConn.NewSession()
	if err != nil {
		return &RemoteCmd{}
	}
	return &RemoteCmd{nil, nil, nil, cmd, session}
}

func (this *LocalCmd) Run() ([]string, error) {
	err := this.Start()
	if err != nil {
		return nil, err
	}
	bOut, err := ioutil.ReadAll(this.StdoutPipe)
	if err != nil {
		return nil, err
	}
	if err := this.Wait(); err != nil {
		return nil, err
	}
	if len(bOut) > 0 {
		return strings.Split(strings.Trim(string(bOut), "\n"), "\n"), nil
	}
	return nil, nil
}

func (this *LocalCmd) Start() error {
	var err error
	this.StdinPipe, err = this.cmd.StdinPipe()
	if err != nil {
		return err
	}
	this.StdoutPipe, err = this.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	this.StderrPipe, err = this.cmd.StderrPipe()
	if err != nil {
		return err
	}
	return this.cmd.Start()
}

func (this *LocalCmd) Wait() error {
	bErr, err := ioutil.ReadAll(this.StderrPipe)
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

func (this *LocalCmd) Stdin() io.WriteCloser {
	return this.StdinPipe
}

func (this *LocalCmd) Stdout() io.Reader {
	return this.StdoutPipe
}

func (this *LocalCmd) Stderr() io.Reader {
	return this.StderrPipe
}

func (this *RemoteCmd) Run() ([]string, error) {
	if err := this.Start(); err != nil {
		return nil, err
	}
	bOut, err := ioutil.ReadAll(this.StdoutPipe)
	if err != nil {
		return nil, err
	}
	if err := this.Wait(); err != nil {
		return nil, err
	}
	if len(bOut) > 0 {
		return strings.Split(strings.Trim(string(bOut), "\n"), "\n"), nil
	}
	return nil, nil
}

func (this *RemoteCmd) Start() error {
	var err error
	this.StdinPipe, err = this.session.StdinPipe()
	if err != nil {
		return err
	}
	this.StdoutPipe, err = this.session.StdoutPipe()
	if err != nil {
		return err
	}
	this.StderrPipe, err = this.session.StderrPipe()
	if err != nil {
		return err
	}
	return this.session.Start(this.cmd)
}

func (this *RemoteCmd) Wait() error {
	defer this.session.Close()
	bErr, err := ioutil.ReadAll(this.StderrPipe)
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

func (this *RemoteCmd) Stdin() io.WriteCloser {
	return this.StdinPipe
}

func (this *RemoteCmd) Stdout() io.Reader {
	return this.StdoutPipe
}

func (this *RemoteCmd) Stderr() io.Reader {
	return this.StderrPipe
}

func NewLocalRunner() (*Local, error) {
	return &Local{}, nil
}

func NewRemoteRunner(user, host, key string) (*Remote, error) {
	bs, err := ioutil.ReadFile(key)
	if err != nil {
		return &Remote{}, err
	}
	signer, err := ssh.ParsePrivateKey(bs)
	if err != nil {
		return &Remote{}, err
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}
	server, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return &Remote{}, err
	}
	return &Remote{server}, nil
}
