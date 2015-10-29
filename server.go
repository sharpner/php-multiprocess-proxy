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
	//TimeOut in ms waits for the php server process
	TimeOut = 130
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
	log.Println("starting server")
	if pg == nil {
		panic("pg must be set")
	}

	p := pg.next()
	if p == nil {
		// we could spawn more processes here
		// but if you have this error often
		// its better to increase the queue size
		// and decrease the timeout
		// since this only happens if you can answer your requests
		// faster on average than the time it takes to start a php server
		// it's currently only reproducible through a local DOS :P
		panic("out of processes")
	}

	defer func(p *phpProcess) {
		log.Printf("Stopping %d", p.port)
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

	fmt.Printf("%#v", req)

	log.Println("Making request")
	resp, err := client.Do(req)
	if err != nil {
		if _, ok := err.(redirectNotAnError); ok {
			if resp == nil {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte{})
				return
			}

			for key := range resp.Header {
				w.Header().Set(key, resp.Header.Get(key))
			}
			w.WriteHeader(resp.StatusCode)
			w.Write([]byte{})

			return
		}

		return
	}

	for key := range resp.Header {
		w.Header().Set(key, resp.Header.Get(key))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		w.Write([]byte{})
		return
	}

	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	w.Write(data)
}

//NewPHPHTTPHandlerFunc returns a php proxy handler
func NewPHPHTTPHandlerFunc(filename string) (http.HandlerFunc, error) {
	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	pg := newProcessGroup(filename)
	for i := 0; i < NumProcesses; i++ {
		go pg.spawn()
	}

	pg.spawn()
	defer pg.clear()
	return func(w http.ResponseWriter, r *http.Request) {
		phpHandler(pg, w, r)
	}, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage server port filename")
		return
	}

	phpFunc, err := NewPHPHTTPHandlerFunc(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", phpFunc)

	log.Printf("Serving %s on :%s\n", os.Args[2], os.Args[1])

	http.ListenAndServe(":"+os.Args[1], nil)
}
