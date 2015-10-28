package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func nextPort() int {
	l, err := net.Listen("tcp4", ":0")
	if err != nil {
		log.Fatal(err)
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
					time.Sleep(100)
					/*
					 *cmd := exec.Command("kill", fmt.Sprintf("%d", c.Process.Pid))
					 *if err := cmd.Run(); err != nil {
					 *  log.Fatal(err)
					 *  return
					 *}
					 */
					log.Printf("cleaning %d.\n", c.Process.Pid)
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
			log.Fatal(err)
			return
		}

	}(complete)

	time.Sleep(300 * time.Millisecond)

	req, err := http.NewRequest(r.Method, fmt.Sprintf("http://localhost:%d%s", port, r.RequestURI), r.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	client := &http.Client{}
	req.Header = r.Header

	log.Println("Making request")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	for key := range resp.Header {
		w.Header().Set(key, resp.Header.Get(key))
	}

	w.Write(data)
	complete <- true
}

func main() {
	if len(os.Args) < 1 {
		fmt.Println("Usage server filename")
		return
	}
	file := os.Args[1]
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		phpHandler(file, w, r)
	})
	http.ListenAndServe(":10000", nil)
}
