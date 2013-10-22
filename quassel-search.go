/*
Server for searching through quassel logs
Greg Schafer
Oct. 2013

References:
http://golang.org/pkg/database/sql/
https://github.com/mattn/go-sqlite3/blob/master/_example/simple/simple.go
*/

package main

import (
  "net/http"
  "html/template"
  "log"
  "fmt"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

type Message struct {
  Channel string
  Time time
  Sender string
  Message string
}

// Database


// Web server



// Main

func main() {
  db, err := sql.Open("sqlite3", "quassel-storage.sqlite")
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  // TODO: join with senders table?
  rows, err := db.Query("select buffercname,sender,time,type,message from backlog natural join sender natural join buffer")
  if err != nil {
    log.Fatal(err)
  }
  defer rows.Close()
  for rows.Next() {
    var time, msgtype int
    var buffercname, sender, msg string
    rows.Scan(&buffercname, &sender, &time, &msgtype, &msg)
    if msgtype == 1 {
      fmt.Println(buffercname, time, sender, msg)
    }
  }

}
