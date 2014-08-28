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
	Run(cmd string) ([]string, error)
	Start(cmd string) (*Command, error)
	WaitCmd() error
}

type Command struct {
	Stdin  io.Writer
	Stdout io.Reader
	Stderr io.Reader
}

type Local struct {
	Cmd        *exec.Cmd
	StdStreams Command
}

type Remote struct {
	Server     *ssh.Client
	Cmd        *ssh.Session
	StdStreams Command
}

func NewLocalRunner() *Local {
	return &Local{}
}

// NB:
// Cannot implement abstract setupPipe as interface, implementing Local and Remote;
// interface mistmatch:
// ssh.Session.StderrPipe() - io.Reader()
// exec.Cmd.StderrPipe() - io.ReadCloser()

func (this *Local) Start(cmd string) (*Command, error) {
	cmdAndArgs := strings.Split(cmd, " ")
	c := exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...)
	stdin, err := c.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := c.Start(); err != nil {
		return nil, err
	}
	this.Cmd = c
	this.StdStreams = Command{stdin, stdout, stderr}
	return &this.StdStreams, nil
}

func (this *Local) Run(cmd string) ([]string, error) {
	c, err := this.Start(cmd)
	if err != nil {
		return nil, err
	}
	bOut, err := ioutil.ReadAll(c.Stdout)
	if err != nil {
		return nil, err
	}
	if err := this.WaitCmd(); err != nil {
		return nil, err
	}
	if len(bOut) > 0 {
		return strings.Split(strings.Trim(string(bOut), "\n"), "\n"), nil
	}
	return nil, nil
}

func (this *Local) WaitCmd() error {
	bErr, err := ioutil.ReadAll(this.StdStreams.Stderr)
	if err != nil {
		return errors.New(this.Cmd.Wait().Error() + "\n" + err.Error())
	}
	if len(bErr) > 0 {
		return errors.New(this.Cmd.Wait().Error() + "\n" + string(bErr))
	}
	return this.Cmd.Wait()
}

func NewRemoteRunner(user, host, key string) (*Remote, error) {
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
	return &Remote{Server: server}, nil
}

func (this *Remote) Start(cmd string) (*Command, error) {
	s, err := this.Server.NewSession()
	if err != nil {
		return nil, err
	}
	stdin, err := s.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := s.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := s.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := s.Start(cmd); err != nil {
		return nil, err
	}
	this.Cmd = s
	this.StdStreams = Command{stdin, stdout, stderr}
	return &this.StdStreams, nil
}

func (this *Remote) Run(cmd string) ([]string, error) {
	c, err := this.Start(cmd)
	if err != nil {
		return nil, err
	}
	bOut, err := ioutil.ReadAll(c.Stdout)
	if err != nil {
		return nil, err
	}
	if err := this.WaitCmd(); err != nil {
		return nil, err
	}
	if len(bOut) > 0 {
		return strings.Split(strings.Trim(string(bOut), "\n"), "\n"), nil
	}
	return nil, nil
}

func (this *Remote) WaitCmd() error {
	bErr, err := ioutil.ReadAll(this.StdStreams.Stderr)
	if err != nil {
		return errors.New(this.Cmd.Wait().Error() + "\n" + err.Error())
	}
	if len(bErr) > 0 {
		return errors.New(this.Cmd.Wait().Error() + "\n" + string(bErr))
	}
	return this.Cmd.Wait()
}
