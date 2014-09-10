runcmd golang package helps you run shell commands on local or remote host.

Note: for remote commands only ssh-key-auth (rsa/dsa) supported

http://godoc.org/github.com/theairkit/runcmd

Example of usage below (see also runcmd_test.go):


```go
package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/theairkit/runcmd"
)

func main() {

	// Run local command:
	lRunner := runcmd.NewLocalRunner()
	out, err := lRunner.Run("ls -la")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Output is a slice of strings; it is pretty for parsing etc.
	for _, i := range out {
		fmt.Println(i)
	}

	// Run remote command:
	rRunner, err := runcmd.NewRemoteRunner("mike", "10.10.0.3:22", "/Users/mike/.ssh/id_rsa")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	out, err = rRunner.Run("ls -la")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Output of remote also slice of strings:
	for _, i := range out {
		fmt.Println(i)
	}

	// Start remote command:
	c, err := rRunner.Start("ls -la")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	/*
		Now we can work with:
		c.Stdout as io.Reader
		c.Stdin as io.Writer
		c.Stderr as io.Reader

		Below we read from c.Stdout to bytes.Buffer,
		convert it to slice of strings and iterate slice:
		(of course, we always need to know, which data in stdout:
		text, binary etc.)
	*/
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Stdout)
	for _, s := range strings.Split(buf.String(), "\n") {
		fmt.Println(s)
	}

	// WaitCmd return the exit code
	// and release resources once the command exits:
	if err := rRunner.WaitCmd(); err != nil {
		fmt.Println(err.Error())
		return
	}
}
```
