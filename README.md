### runcmd

runcmd golang package helps you run shell commands on local or remote hosts

http://godoc.org/github.com/theairkit/runcmd

Installation:
```bash
go get github.com/theairkit/runcmd
```

### Description and examples:

First, create runner: this is a type, that holds:
- for local commands: empty struct
- for remote commands: connect to remote host;
  so, you can create only one remote runner to remote host

Note: there are no ability to set connection timeout in golang-ssh package

(track my request: https://codereview.appspot.com/158760043/)

```go
lRunner, err := runcmd.NewLocalRunner()
if err != nil {
	//handle error
}

rRunner, err := runcmd.NewRemoteKeyAuthRunner(
			"user",
			"127.0.0.1:22",
			"/home/user/id_rsa",
			)
if err != nil {
	//handle error
}

rRunner, err := runcmd.NewRemotePassAuthRunner(
			"user",
			"127.0.0.1:22",
			"userpass",
			)
if err != nil {
	//handle error
}
```

After that, create command, and run methods:
```
c, err := rRunner.Command("date")
if err != nil {
	//handle error
}
out, err := c.Run()
if err != nil {
	//handle error
}
```

Both local and remote runners implements Runner interface,
so, you can work with them as Runner:

```go
func listSomeDir(r Runner) error {
	c, err := r.Command("ls -la")
	if err != nil {
		//handle error
	}
	out, err := c.Run()
	if err != nil {
		//handle error
	}
	for _, i := range out {
		fmt.Println(i)
	}
}

// List some dir on local host:
if err := listSomeDir(lRunner); err != nil {
	//handle error
}

// List some dir on remote host:
if err := listSomeDir(rRunner); err != nil {
	//handle error
}
```

Another useful code snippet: pipe from local to remote command:

```
lRunner, err := NewLocalRunner()
if err != nil {
	//handle error
}

rRunner, err := NewRemoteKeyAuthRunner(user, host, key)
if err != nil {
	//handle error
}

cLocal, err := lRunner.Command("date")
if err != nil {
	//handle error
}
if err = cmdLocal.Start(); err != nil {
	//handle error
}
cRemote, err := rRunner.Command("tee /tmp/tmpfile")
if err != nil {
	//handle error
}
if err = cRemote.Start(); err != nil {
	//handle error
}
if _, err = io.Copy(cRemote.StdinPipe(),cLocal.StdoutPipe(),); err != nil {
	//handle error
}

// Correct handle end of copying:
cmdLocal.Wait()
cmdRemote.StdinPipe().Close()
cmdRemote.Wait()
```

For other examples see runcmd_test.go
Before running 'go test', change next variables in runcmd_test.go:
```go
	//Change it before running the tests:
	user = "user"
	host = "127.0.0.1:22"
	key  = "/home/user/.ssh/id_rsa"
	pass = "somepass"
```
