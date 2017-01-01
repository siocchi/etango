package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
)


type linkDbMem struct{
	links map[string]string
}

// var _ LinkDb = &linkDbMem{}

func newDbMem() *linkDbMem {
	return &linkDbMem{
		links:  make(map[string]string),
	}
}

func (db *linkDbMem) GetLink(key string, r *http.Request) (string, error) {
	l, err := db.links[key]
	if err == false {
		return "", nil // TODO !!
	}

	return l, nil
}

func (db *linkDbMem) AddLink(l string, r *http.Request) (string, error) {
	key :=  fmt.Sprintf("%x", md5.Sum([]byte(l)))
	db.links[key] = l
	return key, nil
}

func (db *linkDbMem) Close() error {
	return nil
}
