package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/go-sql-driver/mysql"
)

type Conf struct {
	Port   string `yaml:"Port"`
	DbHost string `yaml:"DbHost"`
	DbUser string `yaml:"DbUser"`
	DbPass string `yaml:"DbPass"`
	DbName string `yaml:"DbName"`
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
		fmt.Println("Error: could not find configuration file:", confPath)
		fmt.Println("Example configuration file:")
		buf, err := yaml.Marshal(&conf)
		if err != nil {
			log.Fatalln("Couldn't marshal config:", err)
		}
		fmt.Println(buf)
		return
	}
	buf, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Fatalln("Could not read configuration path:", err)
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
		log.Fatalln("Could not connect to the DB", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalln("Database could not be pinged:", err)
	}
}

func handleServerInfo(w http.ResponseWriter, r *http.Request) {
	if r.UserAgent() != "irmaserver" {
		log.Println("User-agent is not \"irmaserver\", but:", r.UserAgent())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var entry Entry
	entryBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading received data:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(entryBytes, &entry)
	if err != nil {
		log.Println("Error parsing received data:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO servers (email, version) VALUES (?, ?)")
	if err != nil {
		log.Println("Error in statement:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer func() {
		if err = stmt.Close(); err != nil {
			log.Println("Failed to close statement:", err)
		}
	}()

	_, err = stmt.Exec(entry.Email, entry.Version)
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok && driverErr.Number == 1062 {
			// (email, version) was not unique, do nothing
		} else {
			log.Println("Failed to store entry:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	var confPath string

	flag.StringVar(&confPath, "config", "conf.yaml", "path to configuration file")
	flag.Parse()

	readConfig(confPath)
	connectToDatabase()

	http.HandleFunc("/", handleServerInfo)

	log.Println("Listening on port", conf.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), nil); err != nil {
		log.Println("Server error:", err)
	}
}
