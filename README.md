### runcmd

runcmd golang package helps you run shell commands on local or remote hosts

http://godoc.org/github.com/theairkit/runcmd

Installation:
```bash
go get github.com/theairkit/runcmd
```

### Description and examples

First, create runner: this is a type, that holds:
- for local commands: empty struct
- for remote commands: connect to remote host;
  so, you can create only one remote runner to remote host

```go
lRunner, err := runcmd.NewLocalRunner()
if err != nil {
	log.Fatal(err.Error())
}
/*
...
*/
rRunner, err := runcmd.NewRemoteKeyAuthRunner("user","127.0.0.1:22","/home/user/id_rsa")
if err != nil {
	log.Fatal(err.Error())
}
/*
...
*/
rRunner, err := runcmd.NewRemotePassAuthRunner("user","127.0.0.1:22","userpass")
if err != nil {
	log.Fatal(err.Error())
}
```

After that, create command, and run methods:
```
c, err := rRunner.Command("date")
if err != nil {
	return err
}
out, err := c.Run()
if err != nil {
	//handle error
}
```

Also, both local and remote runners implements Runner interface,
so, you can pass them as Runner

```go
func listSomeDir(r Runner) error {
	c, err := r.Command("ls -la")
	if err != nil {
		return err
	}
	out, err := c.Run()
	if err != nil {
		return err
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
	return err
}

rRunner, err := NewRemoteKeyAuthRunner(user, host, key)
if err != nil {
	return err
}

cmdLocal, err := lRunner.Command("date")
if err != nil {
	return err
}
if err = cmdLocal.Start(); err != nil {
	return err
}
cmdRemote, err := rRunner.Command("tee /tmp/tmpfile")
if err != nil {
	return err
}
if err = cmdRemote.Start(); err != nil {
	return err
}
if _, err = io.Copy(cmdRemote.StdinPipe(), cmdLocal.StdoutPipe()); err != nil {
	return err
}

// Becasue of using buffered pipe in io.Copy,
// you need correct handle finish copying:
cmdLocal.Wait()
cmdRemote.StdinPipe().Close()
cmdRemote.Wait()
```

For other examples see runcmd_test.go
