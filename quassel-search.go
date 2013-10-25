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
	"code.google.com/p/gcfg"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"time"
)

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

// ===== Models =====

type resultSet struct {
	Needle         string
	ChannelResults []channelResult
}
type channelResult struct {
	Channel  string
	Messages []message
}
type sender struct {
	Username  string
	FullIdent string
}

var re = regexp.MustCompile("(.*)!~?(.*)")

func makeSender(rawSender string) sender {
	m := re.FindStringSubmatch(rawSender)
	if len(m) == 3 {
		return sender{m[1], m[2]}
	}
	return sender{rawSender, rawSender}
}

type message struct {
	MessageId int
	Time      time.Time
	Channel   string
	Sender    sender
	Text      string
}
type frontpageStats struct {
	NumMessages int
	NumSenders  int
	NumChannels int
}

// ===== Database =====
var db *sql.DB

// messagesFromRows returns a synchronous channel of messages
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
			panicIfError(err)

			// sqlite returns date as int64, postgres returns as time.Time
			switch msgTime := msgTime.(type) {
			case int64:
				convertedTime = time.Unix(msgTime, 0)
			case time.Time:
				convertedTime = msgTime
			default:
				panic(fmt.Errorf("unrecognized datetime format for backlog messageid=%d?", messageid))
			}

			m := message{MessageId: messageid,
				Time:    convertedTime,
				Channel: buffercname,
				Sender:  makeSender(msgSender),
				Text:    msg}
			messages <- m
		}
	}()

	return messages
}

func searchResults(needle string) resultSet {
	sql_needle := "%" + needle + "%"
	rows, err := db.Query("select messageid,buffercname,sender,time,message from backlog natural join sender natural join buffer where type = 1 and message like $1", sql_needle)
	panicIfError(err)

	results := resultSet{Needle: needle, ChannelResults: make([]channelResult, 0)}
	channels := make(map[string][]message)
	for m := range messagesFromRows(rows) {
		channels[m.Channel] = append(channels[m.Channel], m)
	}
	for channel, messages := range channels {
		cr := channelResult{Channel: channel, Messages: messages}
		results.ChannelResults = append(results.ChannelResults, cr)
	}
	return results
}

// returns counts for # messages, # users, # channels
func getFrontpageStats() frontpageStats {
	var stats frontpageStats
	var row *sql.Row
	var err error

	row = db.QueryRow("select count(*) from backlog")
	err = row.Scan(&stats.NumMessages)
	panicIfError(err)

	row = db.QueryRow("select count(*) from sender")
	err = row.Scan(&stats.NumSenders)
	panicIfError(err)

	row = db.QueryRow("select count(*) from buffer")
	err = row.Scan(&stats.NumChannels)
	panicIfError(err)

	return stats
}

// direction in which to fetch contextual messages
const (
	BeforeDirection = -1
	AfterDirection  = 1
)

var dirComparators = map[int]string{
	BeforeDirection: "<",
	AfterDirection:  ">",
}
var dirSort = map[int]string{
	BeforeDirection: "desc",
	AfterDirection:  "asc",
}

func messageContext(messageId int, linesToFetch int, direction int) []message {
	if direction != -1 && direction != 1 {
		panic(fmt.Errorf("direction argument must be 1 or -1"))
	}
	row := db.QueryRow("select bufferid from backlog where messageid = $1", messageId)
	var bufferid int
	err := row.Scan(&bufferid)
	panicIfError(err)

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
	panicIfError(err)

	results := make([]message, 0)
	for m := range messagesFromRows(rows) {
		results = append(results, m)
	}
	return results
}

// ===== Web server =====
// matches "/", "/search/", and "/context/"
var validPath = regexp.MustCompile(`^/($|(search|context)\/)`)

func tmplPaths(tmpl ...string) []string {
	paths := make([]string, len(tmpl))
	for i, p := range tmpl {
		paths[i] = path.Join("templates", p)
	}
	return paths
}

var templatePaths = tmplPaths("index.html", "search.html")
var templates = template.Must(template.ParseFiles(templatePaths...))

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	panicIfError(err)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	stats := getFrontpageStats()
	renderTemplate(w, "index", stats)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	needle := r.FormValue("n")
	if needle == "" {
		panic(fmt.Errorf("no search term provided"))
	}
	results := searchResults(needle)
	renderTemplate(w, "search", results)
}

func ajaxContextHandler(w http.ResponseWriter, r *http.Request) {
	messageId, err := strconv.Atoi(r.FormValue("messageId"))
	panicIfError(err)
	linesToFetch, err := strconv.Atoi(r.FormValue("linesToFetch"))
	panicIfError(err)
	direction, err := strconv.Atoi(r.FormValue("direction"))
	panicIfError(err)

	results := messageContext(messageId, linesToFetch, direction)
	jsres, err := json.Marshal(results)
	panicIfError(err)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsres)
	panicIfError(err)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isValidPath := validPath.MatchString(r.URL.Path)
		if !isValidPath {
			http.NotFound(w, r)
			return
		}

		// recover from handler panics
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					log.Fatalf("panic didn't return a string? got %T with value: %v", r, r)
				}
				log.Print(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		fn(w, r)
	}
}

// ===== Config Structs =====
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
	Database  DbConfig
	Webserver WebserverConfig
}

// ===== Main =====
func main() {
	var cfgPath = flag.String("config", "conf.gcfg", "Path to config file")
	flag.Parse()
	log.Print("Reading config from ", *cfgPath)

	var err error
	var configuration Configuration
	err = gcfg.ReadFileInto(&configuration, *cfgPath)
	panicIfError(err)

	var dataSourceName string
	switch configuration.Database.DbType {
	case "sqlite3":
		dataSourceName = configuration.Database.DbPath
	case "postgres":
		dataSourceName = fmt.Sprintf("user=%s password=%s dbname=%s",
			configuration.Database.DbUser,
			configuration.Database.DbPass,
			configuration.Database.DbName)
	default:
		log.Fatalf("%v not a recognized database type (sqlite3 or postgres)", configuration.Database.DbType)
	}

	db, err = sql.Open(configuration.Database.DbType, dataSourceName)
	panicIfError(err)
	defer db.Close()
	err = db.Ping()
	panicIfError(err)

	http.HandleFunc("/", makeHandler(indexHandler))
	http.HandleFunc("/search/", makeHandler(searchHandler))
	http.HandleFunc("/context/", makeHandler(ajaxContextHandler))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	log.Print("Starting to listen on port ", configuration.Webserver.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", configuration.Webserver.Port), nil)
	panicIfError(err)
}
