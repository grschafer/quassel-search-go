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
  "regexp"
  "strconv"
  "encoding/json"
  "path"
  "net/http"
  "html/template"
  "log"
  "time"
  "fmt"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
  _ "github.com/lib/pq"
  "code.google.com/p/gcfg"
)

type resultSet struct {
  Needle string
  ChannelResults []channelResult
}
type channelResult struct {
  Channel string
  Messages []message
}
type sender struct {
  Username string
  FullIdent string
}
var re = regexp.MustCompile("(.*)!~?(.*)")
func makeSender(rawSender string) sender {
  fmt.Println(rawSender)
  m := re.FindStringSubmatch(rawSender)
  fmt.Println(m)
  if len(m) == 3 {
    return sender{m[1], m[2]}
  }
  return sender{rawSender, rawSender}
}
type message struct {
  MessageId int
  Time time.Time
  Channel string
  Sender sender
  Text string
}

// Database
var db *sql.DB

func messagesFromRows(rows *sql.Rows) chan message {
  messages := make(chan message)

  go func() {
    defer rows.Close()
    defer close(messages)
    for rows.Next() {
      var messageid int
      var msgTime interface{}
      var convertedTime time.Time
      var buffercname, msgSender, msg string
      err := rows.Scan(&messageid, &buffercname, &msgSender, &msgTime, &msg)
      if err != nil {
        log.Print(err)
      }
      // sqlite returns date as int64, postgres returns as time.Time
      switch msgTime := msgTime.(type) {
      case int64:
        convertedTime = time.Unix(msgTime, 0)
      case time.Time:
        convertedTime = msgTime
      }
      m := message{MessageId: messageid,
                   Time: convertedTime,
                   Channel: buffercname,
                   Sender: makeSender(msgSender),
                   Text: msg}
      messages <- m
    }
  }()

  return messages
}

// TODO: return error
func searchResults(needle string) (resultSet) {
  sql_needle := "%" + needle + "%"
  rows, err := db.Query("select messageid,buffercname,sender,time,message from backlog natural join sender natural join buffer where type = 1 and message like $1", sql_needle)
  // TODO: more error checking! return error to handler
  if err != nil {
    log.Fatal(err)
  }

  results := resultSet{Needle: needle, ChannelResults: make([]channelResult, 0)}
  channels := make(map[string][]message)
  for m := range messagesFromRows(rows) {
    channels[m.Channel] = append(channels[m.Channel], m)
  }
  for channel,messages := range channels {
    cr := channelResult{Channel: channel, Messages: messages}
    results.ChannelResults = append(results.ChannelResults, cr)
  }
  //fmt.Println("searched for ", needle, " got results: ", len(results.Messages))
  return results
}

/*
TODO: make these into constants
direction values: -1 = before, 1 = after
*/
var dirComparators = map[int]string{
  -1: "<",
  1: ">",
}
var dirSort = map[int]string{
  -1: "desc",
  1: "asc",
}
func messageContext(messageId int, linesToFetch int, direction int) []message {
  row := db.QueryRow("select bufferid from backlog where messageid = $1", messageId)
  var bufferid int
  err := row.Scan(&bufferid)
  if err != nil {
    // TODO: return error (ie - no rows in result set)
    log.Fatal(err)
  }

  query := fmt.Sprintf(`select messageid,buffercname,sender,time,message
                        from backlog
                        natural join sender
                        natural join buffer
                        where messageid %s $1
                          and bufferid = $2
                          and type = 1
                        order by messageid %s
                        limit $3`, dirComparators[direction], dirSort[direction])
  rows, err := db.Query(query, messageId, bufferid, linesToFetch)
  if err != nil {
    log.Fatal(err)
  }

  results := make([]message, 0)
  for m := range messagesFromRows(rows) {
    results = append(results, m)
  }
  return results
}

// Web server
// TODO: accept var num args, return list of all
func tmplPath(tmpl string) string {
  return path.Join("templates", tmpl)
}
var templatePaths = []string{tmplPath("index.html"), tmplPath("search.html")}
var templates = template.Must(template.ParseFiles(templatePaths...))
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
  err := templates.ExecuteTemplate(w, tmpl + ".html", data)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("indexHandler")
  renderTemplate(w, "index", nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("searchHandler")
  needle := r.FormValue("n")
  // call db stuff
  results := searchResults(needle)
  fmt.Println(r)
  fmt.Println(results)
  fmt.Println()
  renderTemplate(w, "search", results)
}

func ajaxContextHandler(w http.ResponseWriter, r *http.Request) {
  // get request data: direction, from msgId, # lines
  // buffer to get data from is implied in msgId
  // TODO: error handling!!
  fmt.Println("ajaxContextHandler")
  messageId, _ := strconv.Atoi(r.FormValue("messageId"))
  linesToFetch, _ := strconv.Atoi(r.FormValue("linesToFetch"))
  direction, _ := strconv.Atoi(r.FormValue("direction"))
  // call db stuff
  results := messageContext(messageId, linesToFetch, direction)
  fmt.Println(r)
  fmt.Println(results)
  jsres, err := json.Marshal(results)
  fmt.Println("err:", err)
  fmt.Println(jsres)
  fmt.Println()
  w.Header().Set("Content-Type", "application/json")
  w.Write(jsres)
}



// Main
type DbConfig struct {
  DbType string // postgres or sqlite3
  DbPath string // path to database file (sqlite only)
  DbName string // name of database (postgres only)
  DbUser string // name of database user (postgres only)
  DbPass string // password of database user (postgres only)
}
type WebserverConfig struct {
  Port int // port to serve log-search website at (default: 4243)
}
type Configuration struct {
  Database DbConfig
  Webserver WebserverConfig
}

// TODO: config file path from cmdline flags?
func main() {
  var err error
  var configuration Configuration
  err = gcfg.ReadFileInto(&configuration, "conf.gcfg")
  if err != nil {
    log.Fatal(err)
  }

  var dataSourceName string
  switch configuration.Database.DbType {
  case "sqlite3":
    dataSourceName = configuration.Database.DbPath
  case "postgres":
    dataSourceName = fmt.Sprintf("user=%s password=%s dbname=%s",
      configuration.Database.DbUser, configuration.Database.DbPass, configuration.Database.DbName)
  default:
    log.Fatalf("%v not a recognized database type (sqlite3 or postgres)", configuration.Database.DbType)
  }

  db, err = sql.Open(configuration.Database.DbType, dataSourceName)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  http.HandleFunc("/", indexHandler)
  http.HandleFunc("/search/", searchHandler)
  http.HandleFunc("/context/", ajaxContextHandler)
  http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
  http.ListenAndServe(fmt.Sprintf(":%d", configuration.Webserver.Port), nil)
}
