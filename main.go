package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Info struct {
	Region string `json:"region_name"`
	City   string `json:"city"`
}

// Echo headers and info
type Echo struct {
	Headers http.Header `json:"headers"`
	IP      string      `json:"ip"`
	Info    Info        `json:"info"`
}

var apiURL = "api.ipstack.com"
var APIKEY = "4e53216245834fa6154a2247d3165396"

func main() {
	startHTTPServer()
}

func startHTTPServer() {
	setLogging()
	fmt.Printf("Starting server,  log file: info.log")

	var PORT string
	var HOST string

	if PORT = os.Getenv("PORT"); PORT == "" {
		PORT = "80"
	}

	if HOST = os.Getenv("HOST"); HOST == "" {
		HOST = ""
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {

		s := strings.Split(r.RemoteAddr, ":")

		info := getUserLocation(s[0])
		log.Printf("[%s] %s %s %s %s", r.RemoteAddr, r.Method, r.URL, info.Region, info.City)

		for _, cookie := range r.Cookies() {
			log.Printf("    cookie: %s %s %s", cookie.Domain, cookie.Name, cookie.Value)
		}

		for _, header := range r.Header {
			log.Printf("    header: %v", header)
		}

		response := Echo{
			Headers: r.Header,
			IP:      r.RemoteAddr,
			Info:    info,
		}
		WriteLog("./static/visitors.txt", s[0], info.Region, info.City)
		responseBytes, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "unable to marshal response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Connection", "close")
		w.Write(responseBytes)
	})

	log.Printf("Starting Server %s", fmt.Sprintf("%s:%s", HOST, PORT))

	err := http.ListenAndServe(fmt.Sprintf("%s:%s", HOST, PORT), nil)
	if err != nil {
		panic(err)
	}
}

func setLogging() {
	logfile := "info.log"
	lf, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)

	if err != nil {
		log.Fatal("OpenLogfile: os.OpenFile:", err)
	}

	log.SetOutput(lf)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func getUserLocation(ip string) Info {
	info := Info{}
	url := fmt.Sprintf("http://%s/%s?access_key=%s&format=1", apiURL, ip, APIKEY)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("error getting info from %s", url)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading info from %v", resp)
	}
	err = json.Unmarshal(respBody, &info)
	log.Printf("location response %s", respBody)
	return info
}

func rhex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func WriteLog(fileName, ip, region, city string) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("failed opening file: %s", err)
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%v, %s, %s, %s\n", time.Now().String(), ip, region, city))
	if err != nil {
		log.Printf("failed writing to file: %s", err)
	}
}
