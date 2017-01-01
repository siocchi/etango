package main

import (
	"net/http"
	"sync/atomic"
	"fmt"
	"strconv"
)


type linkDbMem struct{
	links map[string]string
}

var counter int32 = 0

// var _ LinkDb = &linkDbMem{}

func newDbMem() *linkDbMem {
	return &linkDbMem{
		links:  make(map[string]string),
	}
}

func values(m map[string]string) []Word {
    vs := []Word{}
    for i, t := range m {
			ii, err := strconv.Atoi(i)
			if err!=nil {
					return vs
			}
			v := Word{
				Id: ii,
				Text: t,
			}
      vs = append(vs, v)
    }
    return vs
}


func (db *linkDbMem) GetAll(r *http.Request) ([]Word, error) {
	return values(db.links), nil
}

func (db *linkDbMem) AddLink(l string, r *http.Request) (string, error) {
	ikey := atomic.AddInt32(&counter, 1)
	key := fmt.Sprint(ikey)
	db.links[key] = l
	return key, nil
}

func (db *linkDbMem) Close() error {
	return nil
}
