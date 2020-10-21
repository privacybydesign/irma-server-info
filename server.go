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
	Email   string
	Version string
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
		fmt.Printf("Error: could not find configuration file: %s\n", confPath)
		fmt.Printf("Example configuration file:\n")
		buf, _ := yaml.Marshal(&conf)
		fmt.Printf("%s\n", buf)
		return
	} else {
		buf, err := ioutil.ReadFile(confPath)
		if err != nil {
			log.Fatalf("Could not read %s: %v", confPath, err)
		}
		err = yaml.Unmarshal(buf, &conf)
		if err != nil {
			log.Fatalf("Could not parse config files: %v", err)
		}
	}
}

func connectToDatabase() {
	var err error
	db, err = sql.Open(dbDriver,
		fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", conf.DbUser, conf.DbPass, conf.DbHost, conf.DbName))
	if err != nil {
		log.Fatalf("Could not connect to the DB %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Database could not be pinged: %v", err)
	}
}

func handleServerInfo(w http.ResponseWriter, r *http.Request) {
	if r.UserAgent() != "irmaserver" {
		log.Printf("User-agent %v is not \"irmaserver\"", r.UserAgent())
		return
	}

	var entry Entry
	entryBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading received data: %v", err)
		return
	}
	err = json.Unmarshal(entryBytes, &entry)
	if err != nil {
		log.Printf("Error parsing received data: %v", err)
		return
	}

	stmt, err := db.Prepare("INSERT into servers (email, version) VALUES (?, ?)")
	if err != nil {
		log.Printf("Error in statement: %v", err)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(entry.Email, entry.Version)
	if err != nil {
		log.Printf("Failed to store entry: %v", err)
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

	log.Printf("Listening on port %s", conf.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), nil); err != nil {
		log.Printf("Server error: %v", err)
	}
}
