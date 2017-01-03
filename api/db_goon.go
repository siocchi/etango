
// https://godoc.org/github.com/mjibson/goon
package main

import (
	"strconv"
	"fmt"
	"time"
	"sync/atomic"
	"net/http"
	"github.com/mjibson/goon"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type WordGoon struct {
	Id    string 	`datastore:"-" goon:"id"`
	Text string	`datastore:"text"`
	Memo string `datastore:"memo"`
	Tag	 string `datastore:"tag"`
	IsReview bool `datastore:"is_review"`
	IsInput bool `datastore:"is_input"`
	Count int	 `datastore:"count"`
	Priority int `datastore:"priority"`
	UpdatedAt time.Time `datastore:"updated_at"`
}


type wordDbGoon struct {
	goon string
}
var counter32 int32 = 0

// var _ LinkDb = &linkDbCloud{}

func newDbGoon() *wordDbGoon {
	return &wordDbGoon{goon: ""}
}

func (db *wordDbGoon) GetWord(key string, r *http.Request) (Word, error) {
	g := goon.NewGoon(r)

	w := new(WordGoon)
	w.Id = key

	if err := g.Get(w); err != nil {
		c := appengine.NewContext(r)
		log.Infof(c, "%v", err)
		return Word{}, err
	}

	ikey, err := strconv.Atoi(key)
	if err!=nil {
			return Word{}, err
	}

	v := Word{
		Id: ikey,
		Text: w.Text,
		Memo: w.Memo,
		Tag: w.Tag,
		IsReview: w.IsReview,
		IsInput: w.IsInput,
		Count: w.Count,
		Priority: w.Priority,
		UpdatedAt: w.UpdatedAt,
	}

	return v, nil
}

func (db *wordDbGoon) GetAll(r *http.Request) ([]Word, error) {
	g := goon.NewGoon(r)

	words := []WordGoon{}

	if _, err := g.GetAll(datastore.NewQuery("WordGoon"), &words); err != nil {
		c := appengine.NewContext(r)
		log.Infof(c, "%v", err)
		return []Word{}, err
	}

	ws := []Word{}
	for _, w := range words {
		ikey, err := strconv.Atoi(w.Id)
		if err!=nil {
				return ws, nil
		}
		v := Word{
			Id: ikey,
			Text: w.Text,
			Memo: w.Memo,
			Tag: w.Tag,
			IsReview: w.IsReview,
			IsInput: w.IsInput,
			Count: w.Count,
			Priority: w.Priority,
			UpdatedAt: w.UpdatedAt,
		}
		ws = append(ws, v)
	}

	return ws, nil
}

func (db *wordDbGoon) AddWord(w PostWord, r *http.Request) (string, error) {

	ikey := atomic.AddInt32(&counter32, 1)
	key := fmt.Sprint(ikey)
	wg := WordGoon{
		Id:      key,
		Text: w.Text,
		Memo: "",
		Tag: "",
		IsReview: true,
		IsInput: true,
		Count: 0,
		Priority: 0,
		UpdatedAt: time.Now(),
	}

	g := goon.NewGoon(r)
	if _, err := g.Put(&wg); err != nil {
		c := appengine.NewContext(r)
		log.Infof(c, "%v", err)
		return "", err
	}
	return key, nil
}

func (db *wordDbGoon) Close() error {
	return nil
}
