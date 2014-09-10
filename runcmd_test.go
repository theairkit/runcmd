package runcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

var (
	cmdValid      = "ls -la"
	cmdInvalid    = "blah-blah"
	cmdInvalidKey = "uname -blah"
	cmdPipeOut    = "date"
	cmdPipeIn     = "cat - > /tmp/d"
	user          = "mike"
	host          = "127.0.0.1:22"
	key           = "/home/mike/.ssh/id_rsa"
)

func TestLocalRun(t *testing.T) {
	lRunner := NewLocalRunner()
	if err := testRun(lRunner); err != nil {
		t.Error(err)
	}
}

func TestRemoteRun(t *testing.T) {
	rRunner, err := NewRemoteRunner(user, host, key)
	if err != nil {
		t.Error(err)
	}
	if err := testRun(rRunner); err != nil {
		t.Error(err)
	}
}

func TestLocalStartWait(t *testing.T) {
	lRunner := NewLocalRunner()
	if err := testStartWait(lRunner); err != nil {
		t.Error(err)
	}
	return
}

func TestRemoteStartWait(t *testing.T) {
	rRunner, err := NewRemoteRunner(user, host, key)
	if err != nil {
		t.Error(err)
	}
	if err := testStartWait(rRunner); err != nil {
		t.Error(err)
	}
}

func TestPipeLocal2Remote(t *testing.T) {
	if err := testPipe(true); err != nil {
		t.Error(err)
	}
}

func TestPipeRemote2Local(t *testing.T) {
	if err := testPipe(false); err != nil {
		t.Error(err)
	}
}

func testRun(runner Runner) error {
	// Valid command with valid keys:
	out, err := runner.Run(cmdValid)
	if err != nil {
		return err
	}
	for _, i := range out {
		fmt.Println(i)
	}

	// Valid command with invalid keys:
	if _, err = runner.Run(cmdInvalidKey); err != nil {
		fmt.Println(err.Error())
	} else {
		return errors.New(cmdInvalidKey + ": no invalid keys for command, use another to pass  test")
	}

	// Invalid command:
	if _, err = runner.Run(cmdInvalid); err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return errors.New(cmdInvalid + ": command exists, use another to pass test")
}

func testStartWait(runner Runner) error {
	// Valid command with valid keys:
	cmd, err := runner.Start(cmdValid)
	if err != nil {
		return err
	}
	bufStdOut := new(bytes.Buffer)
	bufStdOut.ReadFrom(cmd.Stdout)
	for _, s := range strings.Split(bufStdOut.String(), "\n") {
		fmt.Println(s)
	}
	if err := runner.WaitCmd(); err != nil {
		return err
	}

	// Valid command with invalid keys:
	cmd, err = runner.Start(cmdInvalidKey)
	if err != nil {
		return err
	}
	bufStdOut = new(bytes.Buffer)
	bufStdOut.ReadFrom(cmd.Stdout)
	for _, s := range strings.Split(bufStdOut.String(), "\n") {
		fmt.Println(s)
	}
	if err := runner.WaitCmd(); err != nil {
		fmt.Println(err.Error())
	} else {
		return errors.New(cmdInvalidKey + ": no invalid keys for command, use another to pass  test")
	}

	// Invalid command:
	cmd, err = runner.Start(cmdInvalid)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	if err := runner.WaitCmd(); err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return errors.New(cmdInvalid + ": command exists, use another to pass test")
}

func testPipe(d bool) error {
	lRunner := NewLocalRunner()
	rRunner, err := NewRemoteRunner(user, host, key)
	if err != nil {
		return err
	}

	// local2remote:
	if d {
		cmdLocal, err := lRunner.Start(cmdPipeOut)
		if err != nil {
			return err
		}
		cmdRemote, err := rRunner.Start(cmdPipeIn)
		if err != nil {
			return err
		}
		if _, err = io.Copy(cmdRemote.Stdin, cmdLocal.Stdout); err != nil {
			return err
		}
		return lRunner.WaitCmd()
	}

	// remote2local:
	cmdLocal, err := lRunner.Start(cmdPipeIn)
	if err != nil {
		return err
	}
	cmdRemote, err := rRunner.Start(cmdPipeOut)
	if err != nil {
		return err
	}
	if _, err = io.Copy(cmdLocal.Stdin, cmdRemote.Stdout); err != nil {
		return err
	}
	return rRunner.WaitCmd()
}
