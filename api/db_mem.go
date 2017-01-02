package main

import (
	"net/http"
	"sync/atomic"
	"fmt"
	"strconv"
)


type wordDbMem struct{
	words map[string]string
}

var counter int32 = 0

func newDbMem() *wordDbMem {
	return &wordDbMem{
		words:  make(map[string]string),
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


func (db *wordDbMem) GetAll(r *http.Request) ([]Word, error) {
	return values(db.words), nil
}

func (db *wordDbMem) AddWord(l string, r *http.Request) (string, error) {
	ikey := atomic.AddInt32(&counter, 1)
	key := fmt.Sprint(ikey)
	db.words[key] = l
	return key, nil
}

func (db *wordDbMem) Close() error {
	return nil
}
