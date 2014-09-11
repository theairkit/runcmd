package runcmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	lRunner, err := NewLocalRunner()
	if err != nil {
		t.Error(err)
	}
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
	lRunner, err := NewLocalRunner()
	if err != nil {
		t.Error(err)
	}
	if err := testStartWait(lRunner); err != nil {
		t.Error(err)
	}
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
	cmd := runner.Command(cmdValid)
	out, err := cmd.Run()
	if err != nil {
		return err
	}
	for _, i := range out {
		fmt.Println(i)
	}

	// Valid command with invalid keys:
	cmd = runner.Command(cmdInvalidKey)
	if _, err = cmd.Run(); err != nil {
		fmt.Println(err.Error())
	} else {
		return errors.New(cmdInvalidKey + ": no invalid keys for command, use another to pass  test")
	}

	// Invalid command:
	cmd = runner.Command(cmdInvalid)
	if _, err = cmd.Run(); err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return errors.New(cmdInvalid + ": command exists, use another to pass test")
}

func testStartWait(runner Runner) error {
	// Valid command with valid keys:
	cmd := runner.Command(cmdValid)
	if err := cmd.Start(); err != nil {
		return err
	}
	bOut, err := ioutil.ReadAll(cmd.Stdout())
	for _, s := range strings.Split(strings.Trim(string(bOut), "\n"), "\n") {
		fmt.Println(s)
	}
	if err := cmd.Wait(); err != nil {
		return err
	}

	// Valid command with invalid keys:
	cmd = runner.Command(cmdInvalidKey)
	if err = cmd.Start(); err != nil {
		return err
	}
	bOut, err = ioutil.ReadAll(cmd.Stdout())
	for _, s := range strings.Split(strings.Trim(string(bOut), "\n"), "\n") {
		fmt.Println(s)
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println(err.Error())
	} else {
		return errors.New(cmdInvalidKey + ": no invalid keys for command, use another to pass  test")
	}

	// Invalid command:
	cmd = runner.Command(cmdInvalid)
	if err = cmd.Start(); err != nil {
		fmt.Println(err.Error())
		return nil
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return errors.New(cmdInvalid + ": command exists, use another to pass test")
}

func testPipe(d bool) error {
	lRunner, err := NewLocalRunner()
	if err != nil {
		return err
	}
	rRunner, err := NewRemoteRunner(user, host, key)
	if err != nil {
		return err
	}

	// local2remote:
	if d {
		cmdLocal := lRunner.Command(cmdPipeOut)
		if err = cmdLocal.Start(); err != nil {
			return err
		}
		cmdRemote := rRunner.Command(cmdPipeIn)
		if err = cmdRemote.Start(); err != nil {
			return err
		}
		if _, err = io.Copy(cmdRemote.Stdin(), cmdLocal.Stdout()); err != nil {
			return err
		}
		return cmdLocal.Wait()
	}
	// remote2local:
	cmdLocal := lRunner.Command(cmdPipeIn)
	if err = cmdLocal.Start(); err != nil {
		return err
	}
	cmdRemote := rRunner.Command(cmdPipeOut)
	if err = cmdRemote.Start(); err != nil {
		return err
	}
	if _, err = io.Copy(cmdLocal.Stdin(), cmdRemote.Stdout()); err != nil {
		return err
	}
	return cmdRemote.Wait()
}
