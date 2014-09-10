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

	// Valid command with valid keys:
	out, err := lRunner.Run(cmdValid)
	if err != nil {
		t.Error(err)
	}
	for _, i := range out {
		fmt.Println(i)
	}

	// Valid command with invalid keys:
	if _, err = lRunner.Run(cmdInvalidKey); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Error(cmdInvalidKey + ": no invalid keys for command, use another to pass  test")
	}

	// Invalid command:
	if _, err = lRunner.Run(cmdInvalid); err != nil {
		fmt.Println(err.Error())
		return
	}
	t.Error(errors.New(cmdInvalid + ": command exists, use another to pass test"))
}

func TestLocalStartWait(t *testing.T) {
	lRunner := NewLocalRunner()

	// Valid command with valid keys:
	cmd, err := lRunner.Start(cmdValid)
	if err != nil {
		t.Fatal(err)
	}
	bufStdOut := new(bytes.Buffer)
	bufStdOut.ReadFrom(cmd.Stdout)
	for _, s := range strings.Split(bufStdOut.String(), "\n") {
		fmt.Println(s)
	}
	if err := lRunner.WaitCmd(); err != nil {
		t.Error(err)
	}

	// Valid command with invalid keys:
	cmd, err = lRunner.Start(cmdInvalidKey)
	if err != nil {
		t.Fatal(err)
	}
	bufStdOut = new(bytes.Buffer)
	bufStdOut.ReadFrom(cmd.Stdout)
	for _, s := range strings.Split(bufStdOut.String(), "\n") {
		fmt.Println(s)
	}
	if err := lRunner.WaitCmd(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Error(cmdInvalidKey + ": no invalid keys for command, use another to pass  test")
	}

	// Invalid command:
	cmd, err = lRunner.Start(cmdInvalid)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	t.Error(errors.New(cmdInvalid + ": command exists, use another to pass test"))
}

func TestRemoteRun(t *testing.T) {
	rRunner, err := NewRemoteRunner(user, host, key)
	if err != nil {
		t.Error(err)
	}

	// Valid command with valid keys:
	out, err := rRunner.Run(cmdValid)
	if err != nil {
		t.Error(err)
	}
	for _, i := range out {
		fmt.Println(i)
	}

	// Valid command with invalid keys:
	if _, err = rRunner.Run(cmdInvalidKey); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Error(cmdInvalidKey + ": no invalid keys for command, use another to pass  test")
	}

	// Invalid command:
	if _, err = rRunner.Run(cmdInvalid); err != nil {
		fmt.Println(err.Error())
		return
	}
	t.Error(errors.New(cmdInvalid + ": command exists, use another to pass test"))
}

func TestRemoteStartWait(t *testing.T) {
	rRunner, err := NewRemoteRunner(user, host, key)
	if err != nil {
		t.Error(err)
	}

	// Valid command with valid keys:
	cmd, err := rRunner.Start(cmdValid)
	if err != nil {
		t.Error(err)
	}
	bufStdOut := new(bytes.Buffer)
	bufStdOut.ReadFrom(cmd.Stdout)
	for _, s := range strings.Split(bufStdOut.String(), "\n") {
		fmt.Println(s)
	}
	if err := rRunner.WaitCmd(); err != nil {
		t.Error(err)
	}

	// Valid command with invalid keys:
	cmd, err = rRunner.Start(cmdInvalidKey)
	if err != nil {
		t.Fatal(err)
	}
	bufStdOut = new(bytes.Buffer)
	bufStdOut.ReadFrom(cmd.Stdout)
	for _, s := range strings.Split(bufStdOut.String(), "\n") {
		fmt.Println(s)
	}
	if err := rRunner.WaitCmd(); err != nil {
		fmt.Println(err.Error())
	} else {
		t.Error(cmdInvalidKey + ": no invalid keys for command, use another to pass  test")
	}

	// Invalid command:
	cmd, err = rRunner.Start(cmdInvalid)
	if err != nil {
		t.Fatal(err)
	}
	if err := rRunner.WaitCmd(); err != nil {
		fmt.Println(err.Error())
		return
	}
	t.Error(errors.New(cmdInvalid + ": command exists, use another to pass test"))
}

func TestPipeLocal2Remote(t *testing.T) {
	lRunner := NewLocalRunner()
	rRunner, err := NewRemoteRunner(user, host, key)
	if err != nil {
		t.Error(err)
	}
	cmdLocal, err := lRunner.Start(cmdPipeOut)
	if err != nil {
		t.Fatal(err)
	}
	cmdRemote, err := rRunner.Start(cmdPipeIn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(cmdRemote.Stdin, cmdLocal.Stdout); err != nil {
		t.Error(err)
	}
	if err := lRunner.WaitCmd(); err != nil {
		t.Error(err)
	}
}

func TestPipeRemote2Local(t *testing.T) {
	lRunner := NewLocalRunner()
	rRunner, err := NewRemoteRunner(user, host, key)
	if err != nil {
		t.Error(err)
	}
	cmdLocal, err := lRunner.Start(cmdPipeIn)
	if err != nil {
		t.Fatal(err)
	}
	cmdRemote, err := rRunner.Start(cmdPipeOut)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(cmdLocal.Stdin, cmdRemote.Stdout); err != nil {
		t.Error(err)
	}
	if err := rRunner.WaitCmd(); err != nil {
		t.Error(err)
	}
}
