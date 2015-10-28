package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"time"
)

func nextPort() int {
	l, err := net.Listen("tcp4", ":0")
	if err != nil {
		log.Print(err)
	}
	defer l.Close()

	x, _ := l.Addr().(*net.TCPAddr)

	return x.Port
}

func clean(complete chan bool, c *exec.Cmd) {
	for {
		select {
		case <-complete:
			{
				go func() {
					if c.Process != nil {
						c.Process.Kill()
					}
				}()
			}
		}
	}
}

func phpHandler(script string, w http.ResponseWriter, r *http.Request) {
	port := nextPort()
	log.Println(port)
	args := []string{"-S", fmt.Sprintf("0.0.0.0:%d", port), script}

	complete := make(chan bool)

	log.Printf("child request url http://localhost:%d%s\n", port, r.RequestURI)
	log.Println("starting server")
	go func(complete chan bool) {
		cmd := exec.Command("php", args...)
		go clean(complete, cmd)
		if err := cmd.Run(); err != nil {
			log.Println(err.Error())
			return
		}

	}(complete)

	time.Sleep(300 * time.Millisecond)

	req, err := http.NewRequest(r.Method, fmt.Sprintf("http://localhost:%d%s", port, r.RequestURI), r.Body)
	if err != nil {
		log.Print(err)
		return
	}

	jarCopy, err := cookiejar.New(nil)
	if err != nil {
		log.Print(err)
		return
	}

	jarCopy.SetCookies(r.URL, r.Cookies())

	client := &http.Client{
		Jar: jarCopy,
	}
	req.Header = r.Header

	log.Println("Making request")
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return
	}

	for key := range resp.Header {
		w.Header().Set(key, resp.Header.Get(key))
	}

	defer func(complete chan bool) {
		complete <- true
	}(complete)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		w.Write([]byte{})
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(data)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage server port filename ")
		return
	}
	file := os.Args[2]
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		phpHandler(file, w, r)
	})
	port := os.Args[1]
	http.ListenAndServe(":"+port, nil)
}
