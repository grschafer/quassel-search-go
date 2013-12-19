# quassel-search-go

[Quassel](http://quassel-irc.org/) log text search server written in Go

This project was inspired by a friend's observation about Quassel's lack of backlog search (without having to scroll up a lot to load historical messages for a single channel) and my recent attendance of a Go talk. It was a fun small-scope project for learning Go.


## New Install Instructions (Vagrant/Ansible)

These instructions create a local VM with vagrant, install quassel-core and quassel-search-go, and run both.

0. Install [vagrant](http://www.vagrantup.com/) and [ansible](http://www.ansibleworks.com/).
1. Run `vagrant up` in the repo directory (this should also run the ansible provisioning, if not, run `vagrant provision`)
    1. Give `vagrant` as the sudo password
2. Connect to the core at localhost:4242 with a quassel client and follow the configuration wizard.
    1. Currently only sqlite is supported as a backend. Postgres coming soon, maybe.
    2. If you visit localhost:4243 before completing the above step, you'll probably just get a 500 response because tables haven't been made in the db yet.
3. Visit localhost:4243 and search your quassel logs!


## Old Install Instructions

**Note: Running this program shows your IRC logs (including private channel messages) on port 4243 by default. If you make this program/port externally visible then anyone will be able to see all your messages. An alternative is to ssh tunnel (`ssh -L 4243:<remote hostname or ip>:4243 <user>@<remote hostname or ip>`) so that traffic to localhost:4243 will go to the remote server's port 4243.**

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


## Acknowledgements

Thanks to everyone involved with Go and its documentation!
