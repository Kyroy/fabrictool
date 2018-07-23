# fabrictool

An example project for a command-line tool in Go.

```bash
vgo build -ldflags "-X github.com/kyroy/fabrictool/cmd.version=0.0.1 -X github.com/kyroy/fabrictool/cmd.gitCommit=$(git rev-parse HEAD)"
```
