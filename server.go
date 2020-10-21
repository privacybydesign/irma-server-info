package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type Conf struct {
	Port   string
	DbHost string
	DbUser string
	DbPass string
	DbName string
}

type Entry struct {
	Email   string `json:"email"`
	Version string `json:"version"`
}

var (
	conf = Conf{
		Port:   "8080",
		DbHost: "localhost",
		DbUser: "serverinfo",
		DbPass: "serverinfo",
		DbName: "serverinfo",
	}
	db       *sql.DB
	dbDriver = "mysql"
)

func readConfig(confPath string) {
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		fmt.Println("Error: could not find configuration file: ", confPath)
		fmt.Println("Example configuration file:")
		buf, err := yaml.Marshal(&conf)
		if err != nil {
			log.Fatalln("Couldn't marshal config: ", err)
		}
		fmt.Println(buf)
		return
	}
	buf, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Fatalln("Could not read configuration path: ", err)
	}
	err = yaml.Unmarshal(buf, &conf)
	if err != nil {
		log.Fatalln("Could not parse config files:", err)
	}
}

func connectToDatabase() {
	var err error
	db, err = sql.Open(dbDriver,
		fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", conf.DbUser, conf.DbPass, conf.DbHost, conf.DbName))
	if err != nil {
		log.Fatalln("Could not connect to the DB ", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalln("Database could not be pinged: ", err)
	}
}

func handleServerInfo(w http.ResponseWriter, r *http.Request) {
	if r.UserAgent() != "irmaserver" {
		log.Printf("User-agent %v is not \"irmaserver\"", r.UserAgent())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var entry Entry
	entryBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading received data: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(entryBytes, &entry)
	if err != nil {
		log.Println("Error parsing received data: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO servers (email, version) VALUES (?, ?)")
	if err != nil {
		log.Println("Error in statement: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(entry.Email, entry.Version)
	if err != nil {
		log.Println("Failed to store entry: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	var confPath string

	flag.StringVar(&confPath, "config", "conf.yaml", "path to configuration file")
	flag.Parse()

	readConfig(confPath)
	connectToDatabase()

	http.HandleFunc("/serverinfo", handleServerInfo)

	log.Println("Listening on port ", conf.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), nil); err != nil {
		log.Println("Server error: ", err)
	}
}
