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
	Cmd *exec.Cmd
	IOE Command
}

type Remote struct {
	Server *ssh.Client
	Cmd    *ssh.Session
	IOE    Command
}

func NewLocalRunner() *Local {
	return &Local{}
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

func (local *Local) Start(cmd string) (*Command, error) {
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
	local.Cmd = c
	local.IOE = Command{stdin, stdout, stderr}
	return &local.IOE, nil
}

func (remote *Remote) Start(cmd string) (*Command, error) {
	s, err := remote.Server.NewSession()
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
	remote.Cmd = s
	remote.IOE = Command{stdin, stdout, stderr}
	return &remote.IOE, nil
}

func (local *Local) Run(cmd string) ([]string, error) {
	c, err := local.Start(cmd)
	if err != nil {
		return nil, err
	}
	bOut, err := ioutil.ReadAll(c.Stdout)
	if err != nil {
		return nil, err
	}
	if err := local.WaitCmd(); err != nil {
		return nil, err
	}
	if len(bOut) > 0 {
		return strings.Split(strings.Trim(string(bOut), "\n"), "\n"), nil
	}
	return nil, nil
}

func (remote *Remote) Run(cmd string) ([]string, error) {
	c, err := remote.Start(cmd)
	if err != nil {
		return nil, err
	}
	bOut, err := ioutil.ReadAll(c.Stdout)
	if err != nil {
		return nil, err
	}
	if err := remote.WaitCmd(); err != nil {
		return nil, err
	}
	if len(bOut) > 0 {
		return strings.Split(strings.Trim(string(bOut), "\n"), "\n"), nil
	}
	return nil, nil
}

func (this *Local) WaitCmd() error {
	bErr, err := ioutil.ReadAll(this.IOE.Stderr)
	if err != nil {
		return errors.New(this.Cmd.Wait().Error() + "\n" + err.Error())
	}
	if len(bErr) > 0 {
		return errors.New(this.Cmd.Wait().Error() + "\n" + string(bErr))
	}
	return this.Cmd.Wait()
}

func (this *Remote) WaitCmd() error {
	bErr, err := ioutil.ReadAll(this.IOE.Stderr)
	if err != nil {
		return errors.New(this.Cmd.Wait().Error() + "\n" + err.Error())
	}
	if len(bErr) > 0 {
		return errors.New(this.Cmd.Wait().Error() + "\n" + string(bErr))
	}
	return this.Cmd.Wait()
}
