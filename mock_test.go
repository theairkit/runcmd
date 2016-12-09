package runcmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockRunner_ReturnsStdoutAndStderr(t *testing.T) {
	test := assert.New(t)

	runner := MockRunner{
		Stdout: []byte("some stdout"),
		Stderr: []byte("some stderr"),
	}

	command := runner.Command("test")
	test.NotNil(command)

	stdout, stderr, err := command.Output()
	test.NoError(err)
	test.Equal(runner.Stdout, stdout)
	test.Equal(runner.Stderr, stderr)
}

func TestMockRunner_ReadsInputStream(t *testing.T) {
	test := assert.New(t)

	runner := MockRunner{}

	command := runner.Command("test")
	test.NotNil(command)

	stdin := bytes.NewBufferString("hello")

	command.SetStdin(stdin)

	err := command.Run()
	test.NoError(err)

	test.Equal(0, stdin.Len())
}

func TestMockRunner_WritesOutputStreams(t *testing.T) {
	test := assert.New(t)

	runner := MockRunner{
		Stdout: []byte("some stdout"),
		Stderr: []byte("some stderr"),
	}

	command := runner.Command("test")
	test.NotNil(command)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	command.SetStdout(stdout)
	command.SetStderr(stderr)

	err := command.Run()
	test.NoError(err)

	test.Equal(runner.Stdout, stdout.Bytes())
	test.Equal(runner.Stderr, stderr.Bytes())
}

func TestMockRunner_ReturnsSetUserError(t *testing.T) {
	test := assert.New(t)

	runner := MockRunner{
		Error: errors.New("test error"),
	}

	command := runner.Command("test")
	test.NotNil(command)

	err := command.Run()
	test.Error(err)
	test.Equal(runner.Error, err)
}
