package main

import (
	"net/http"
	"sync/atomic"
	"fmt"
	"strconv"
	"time"
)


type wordDbMem struct{
	words map[string]Word
}

var counter int32 = 0

func newDbMem() *wordDbMem {
	return &wordDbMem{
		words:  make(map[string]Word),
	}
}

func values(m map[string]Word) []Word {
    vs := []Word{}
    for i, w := range m {
			ii, err := strconv.Atoi(i)
			if err!=nil {
					return vs
			}
			v := Word{
				Id: ii,
				Text: w.Text,
				Memo: w.Memo,
				Tag: w.Tag,
				IsReview: w.IsReview,
				IsInput: w.IsInput,
				Count: w.Count,
				Priority: w.Priority,
				UpdatedAt: w.UpdatedAt,
			}
      vs = append(vs, v)
    }
    return vs
}


func (db *wordDbMem) GetAll(r *http.Request) ([]Word, error) {
	return values(db.words), nil
}

func (db *wordDbMem) AddWord(w PostWord, r *http.Request) (string, error) {
	ikey := atomic.AddInt32(&counter, 1)
	key := fmt.Sprint(ikey)
	db.words[key] = Word{
		Id: 0,
		Text: w.Text,
		Memo: "",
		Tag: "",
		IsReview: true,
		IsInput: true,
		Count: 0,
		Priority: 0,
		UpdatedAt: time.Now(),
	}
	return key, nil
}

func (db *wordDbMem) Close() error {
	return nil
}
