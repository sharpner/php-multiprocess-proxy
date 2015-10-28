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
	TimeOut = 240
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
		panic("no more processes")
	}

	defer p.stop()
	log.Printf("child request url http://localhost:%d%s\n", p.port, r.RequestURI)

	req, err := http.NewRequest(r.Method, fmt.Sprintf("http://localhost:%d%s", p.port, r.RequestURI), r.Body)
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

	pg := newProcessGroup(file)
	for i := 0; i < 5; i++ {
		pg.spawn()
	}

	pg.spawn()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		phpHandler(pg, w, r)
	})
	port := os.Args[1]
	http.ListenAndServe(":"+port, nil)
}
