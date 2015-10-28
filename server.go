package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
)

const (
	//TimeOut waits for the php server process
	TimeOut = 140
	//NumProcesses is the total number of waiting php servers
	NumProcesses = 7
	//BindIP on which the server listens
	BindIP = "0.0.0.0"
	//BindProtocol is the listening protocol
	BindProtocol = "tcp4"
	//ProxyIP on which the php server listens
	ProxyIP = "localhost"
	//ProxyProtocol the proxied protocol
	ProxyProtocol = "http"
)

type redirectNotAnError error

func noRedirect(req *http.Request, via []*http.Request) error {
	var err redirectNotAnError
	err = errors.New("do not follow redirect please")
	return err
}

func phpHandler(pg *phpProcessGroup, w http.ResponseWriter, r *http.Request) {
	complete := make(chan bool)

	log.Println("starting server")

	p := pg.next()
	if p == nil {
		panic("out of processes")
	}

	defer func(p *phpProcess) {
		go p.stop()
	}(p)

	requestURI := fmt.Sprintf("%s://%s:%d%s", ProxyProtocol, ProxyIP, p.port, r.RequestURI)
	log.Printf("child request url %s \n", requestURI)

	req, err := http.NewRequest(r.Method, requestURI, r.Body)
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
		Jar:           jarCopy,
		CheckRedirect: noRedirect,
	}
	req.Header = r.Header

	log.Println("Making request")
	resp, err := client.Do(req)
	if err != nil {
		if _, ok := err.(redirectNotAnError); ok {
			for key := range resp.Header {
				w.Header().Set(key, resp.Header.Get(key))
			}

			w.WriteHeader(resp.StatusCode)
			w.Write([]byte{})
			return
		}

		defer func(complete chan bool) {
			complete <- true
		}(complete)

		log.Fatal(err)
		return
	}

	for key := range resp.Header {
		w.Header().Set(key, resp.Header.Get(key))
	}

	defer p.stop()

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
		fmt.Println("Usage server port filename")
		return
	}
	file := os.Args[2]

	pg := newProcessGroup(file)
	for i := 0; i < NumProcesses; i++ {
		go pg.spawn()
	}

	pg.spawn()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		phpHandler(pg, w, r)
	})
	port := os.Args[1]

	log.Printf("Serving %s on :%s\n", os.Args[2], os.Args[1])

	http.ListenAndServe(":"+port, nil)
}
