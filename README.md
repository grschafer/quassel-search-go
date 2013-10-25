# quassel-search-go

[Quassel](http://quassel-irc.org/) log text search server written in Go

This project was inspired by a friend's complaint about Quassel's lack of backlog search and my recent attendance of a Go talk. It was a fun small-scope project for learning Go.


## To Run

The included binary is 64-bit linux. If that binary will work for you, skep to step 3.

0. Install Go: <http://golang.org/doc/install>
1. Get the dependencies:
    1. `go get "github.com/mattn/go-sqlite3"`
    2. `go get "github.com/lib/pq"`
    3. `go get "code.google.com/p/gcfg"`
2. Build with `go build`
3. Edit the config (`conf.gcfg`) file to use your database version (and credentials)
4. Run the binary
    * Note: If you're using sqlite3, then you will probably need to run the binary as the quassel user with a command like the following:

        ```sudo su - quasselcore -s "/bin/sh" -c "cd /path/to/binary; ./quassel-search-go"```


## Development Potential Todos

1. Add authentication (http basic with TLS?)
2. System startup script?
3. Cross-compile binaries

## Acknowledgements

Thanks to everyone involved with Go and its documentation!
