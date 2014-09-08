runcmd helps you run shell commands locally ro on some remote host.

http://godoc.org/github.com/theairkit/runcmd

Example of usage:


```go
package main

import (
	"fmt"

	"github.com/theairkit/runcmd"
)

func main() {

	// Run local command:
	lRunner := runcmd.NewLocalRunner()
	if out, err := lRunner.Run("ls -la"); err != nil {
		fmt.Println(err.Error())
		return
	} else {
		for _, i := range out {
			fmt.Println(i)
		}
	}

	// Run remote command:
	rRunner, err := runcmd.NewRemoteRunner("root", "192.168.20.80:22", "/home/mike/.ssh/id_rsa")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if out, err := rRunner.Run("ls -la"); err != nil {
		fmt.Println(err.Error())
		return
	} else {
		for _, i := range out {
			fmt.Println(i)
		}
	}

	// Start remote command:
	c, err := rRunner.Start("ls -la")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	/*
	   Do some staff here;
	   We can work here with:
	   c.Stdout
	   c.Stdin
	   c.Stderr
	*/

	// WaitCmd return the exit code
	// and release resources once the command exits:
	if err := rRunner.WaitCmd(); err != nil {
		fmt.Println(err.Error())
		return
	}
}

```
