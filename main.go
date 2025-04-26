package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

type MetricList struct {
	Port           int    `yaml:"port"`
	Timeout_server int    `yaml:"timeout_server"`
	Certfile       string `yaml:"certfile"`
	Keyfile        string `yaml:"keyfile"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Metrics        []struct {
		ProcessName string `yaml:"process_name"`
		Command     string `yaml:"command"`
		Type        string `yaml:"type"`
	} `yaml:"metrics"`
}

func checkFileExist(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (srv *MetricList) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(srv.Username))
			expectedPasswordHash := sha256.Sum256([]byte(srv.Password))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func Exporter(srv MetricList) {
	var wg sync.WaitGroup
	for _, value := range srv.Metrics {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prometheus.MustRegister(newUsageCollector(value.ProcessName, value.Command))
		}()
	}
	wg.Wait()
	prometheus.Unregister(collectors.NewGoCollector())
	mux := http.NewServeMux()
	handler := &http.Server{
		Handler:     mux,
		ReadTimeout: time.Duration(srv.Timeout_server) * time.Second,
		Addr:        ":" + strconv.Itoa(srv.Port),
	}

	if srv.Username != "" || srv.Password != "" {
		mux.HandleFunc("/metrics", srv.basicAuth(promhttp.Handler().ServeHTTP))
	} else {
		mux.Handle("/metrics", promhttp.Handler())
	}
	log.Println("Starting at port " + strconv.Itoa(srv.Port))
	if checkFileExist(srv.Certfile) && checkFileExist(srv.Keyfile) {
		go log.Fatal(handler.ListenAndServeTLS(srv.Certfile, srv.Keyfile))
	} else {
		go log.Fatal(handler.ListenAndServe())
	}
}

func ExecCommand(command string) float64 {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	var stdout bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		return 0
	}

	output := stdout.Bytes()
	if output == nil {
		return 0
	}
	filter := regexp.MustCompile(`[^0-9e+.]+`).ReplaceAllString(string(output), "")
	value, err := strconv.ParseFloat(filter, 64)
	if err != nil {
		return 0
	}
	return value
}

func main() {
	var fileName string
	flag.StringVar(&fileName, "config.file", "", "Location of config file")
	flag.Parse()
	if len(fileName) == 0 {
		log.Println("Usage: \n./cmd-exporter --config.file=server.yaml")
		flag.PrintDefaults()
		os.Exit(1)
	}
	data, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
		return
	}
	var server MetricList
	err = yaml.Unmarshal(data, &server)
	if err != nil {
		log.Fatal(err)
		return
	}
	Exporter(server)
}
