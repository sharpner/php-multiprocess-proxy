package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"sync"
	"time"
)

//ProcessGroup can control
type ProcessGroup interface {
	Clear()
	Spawn()
	Next() Process
}

//Process interface
type Process interface {
	Port() int
	Stop()
}

type phpProcessGroup struct {
	sync.Mutex
	processes []phpProcess
	script    string
}

func newProcessGroup(script string) ProcessGroup {
	return &phpProcessGroup{script: script}
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

func (pg *phpProcessGroup) Spawn() {
	p := phpProcess{}
	p.done = make(chan bool)
	port := nextPort()
	p.port = port
	log.Printf("starting new process %d\n", port)
	args := []string{"-S", fmt.Sprintf("%s:%d", BindIP, port), pg.script}
	go func(complete chan bool) {
		cmd := exec.Command("php", args...)
		go clean(complete, cmd)
		if err := cmd.Run(); err != nil {
			log.Println(err.Error())
			return
		}

	}(p.done)
	pg.Lock()
	defer pg.Unlock()
	time.Sleep(TimeOut * time.Millisecond)
	a := append(pg.processes, p)
	pg.processes = a
}

func (pg *phpProcessGroup) Next() (p Process) {
	pg.Lock()
	defer pg.Unlock()
	defer func() {
		if r := recover(); r != nil {
			p = nil
		}
	}()

	p, a := &pg.processes[0], pg.processes[1:]
	pg.processes = a
	go pg.Spawn()
	return p
}

func (pg *phpProcessGroup) Clear() {
	pg.Lock()
	defer pg.Unlock()
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	for _, p := range pg.processes {
		p.Stop()
	}
}

type phpProcess struct {
	process *exec.Cmd
	done    chan bool
	port    int
}

func (p phpProcess) Port() int {
	return p.port
}

func (p *phpProcess) Stop() {
	p.done <- true
}

func nextPort() int {
	l, err := net.Listen(BindProtocol, ":0")
	if err != nil {
		log.Print(err)
	}
	defer l.Close()

	x, _ := l.Addr().(*net.TCPAddr)

	return x.Port
}
