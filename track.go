package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// 配置信息
type Config struct {
	BatchInsertSeconds int      `json:"batchInsertSeconds"`
	Port               int      `json:"port"`
	DbConfig           DbConfig `json:"database"`
}

// 配置文件的目录
var configFilePath string

// DB 配置信息
type DbConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
	Name string `json:"name"`
}

type Click struct {
	IPAddress string `json:"ipAddress"`
	URL       string `json:"url"`
	Href      string `json:"href"`
	UserAgent string `json:"userAgent"`
}

var Clicks []Click

func listenForRecords(db *sql.DB, seconds time.Duration) {
	// 每个几秒执行一次
	for _ = range time.Tick(seconds) {
		newClicks := make([]Click, len(Clicks))
		copy(newClicks, Clicks)
		go SetClicks(db, newClicks)
		Clicks = Clicks[0:0]
	}
}

func IPAddress(remoteAddr string) string {
	arr := strings.Split(remoteAddr, ":")
	return arr[0]
}

func clickHandler(w http.ResponseWriter, r *http.Request, body []byte) {
	Click := Click{}
	if err := json.Unmarshal(body, &Click); err != nil {
		log.Println("配置文件解析失败: ", err)
	}
	Click.IPAddress = IPAddress(r.RemoteAddr)
	Clicks = append(Clicks, Click)
	w.WriteHeader(201)
}

func readConfig(configFilePath string) Config {
	config := Config{}
	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatal("配置文件读取失败: ", err)
	}
	if err = json.Unmarshal(configFile, &config); err != nil {
		log.Fatal("配置文件解析失败: ", err)
	}
	return config
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, []byte)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "x-requested-with, x-requested-by, Content-Type")
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Unable to read requeset body: ", err)
		}
		fn(w, r, body)
	}
}

func init() {
	//goPath := os.Getenv("GOPATH")
	//defaultConfigPath := fmt.Sprintf("%s/src/github.com/roberttstephens/webanalytics/config.json", goPath)
	defaultConfigPath := "./config.json"
	//defaultConfigPath := "D:/go/src/webanalytics/config.json"
	fmt.Println(defaultConfigPath)
	flag.StringVar(&configFilePath, "config", defaultConfigPath, "path to config.json")
}

func main() {
	// Read the config, initialize the database and listen for records.
	flag.Parse()
	config := readConfig(configFilePath)
	db := Db(config.DbConfig)
	fmt.Println(db)
	seconds := time.Duration(config.BatchInsertSeconds) * time.Second
	go listenForRecords(db, seconds)

	http.HandleFunc("/click-track", makeHandler(clickHandler))
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
