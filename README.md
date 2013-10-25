# quassel-search-go

[Quassel](http://quassel-irc.org/) log text search server written in Go

This project was inspired by a friend's observation about Quassel's lack of backlog search (without having to scroll up a lot to load historical messages for a single channel) and my recent attendance of a Go talk. It was a fun small-scope project for learning Go.


## To Run

**Note: Running this program shows your IRC logs (including private channel messages) on port 4243 by default. If you make this program/port externally visible then anyone will be able to see all your messages.**

Try the appropriate binary if you're on linux (in which case skip to step 3). Otherwise, follow the first few steps to compile it yourself.

0. Install Go: <http://golang.org/doc/install>
1. Get the dependencies:
    1. `go get "github.com/mattn/go-sqlite3"`
    2. `go get "github.com/lib/pq"`
    3. `go get "code.google.com/p/gcfg"`
2. Build with `go build`
3. Edit the config (`conf.gcfg`) file to use your database type (and credentials)
4. Run the binary
    * Note: If you're using sqlite3, then you will probably need to run the binary as the quassel user with a command like the following:

        ```sudo su - quasselcore -s "/bin/sh" -c "cd /path/to/binary; ./quassel-search-go"```


## Development Potential Todos

1. Add authentication (http basic with TLS?)
2. System startup script?
3. Cross-compile binaries?


## Acknowledgements

Thanks to everyone involved with Go and its documentation!
