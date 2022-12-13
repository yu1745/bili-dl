package util

import "sync"

type GoLimit struct {
	c    chan struct{}
	wait *sync.WaitGroup
}

func NewGoLimit(max int) *GoLimit {
	return &GoLimit{
		c:    make(chan struct{}, max),
		wait: &sync.WaitGroup{},
	}
}

func (g *GoLimit) Add() {
	g.c <- struct{}{}
	g.wait.Add(1)
}

func (g *GoLimit) Done() {
	<-g.c
	g.wait.Done()
}

func (g *GoLimit) Wait() {
	g.wait.Wait()
}
