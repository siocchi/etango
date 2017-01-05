
// https://godoc.org/github.com/mjibson/goon
package main

import (
	"time"
	"net/http"
	"github.com/mjibson/goon"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	 "github.com/google/uuid"
	 "regexp"
	 "errors"
)

type WordGoon struct {
	Id    string 	`datastore:"-" goon:"id"`
	Uid    string 	`datastore:"uid" goon:"uid"`
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

func init() {

}
// var _ LinkDb = &linkDbCloud{}
var check_uid = false

func newDbGoon() *wordDbGoon {
	return &wordDbGoon{goon: ""}
}

func (db *wordDbGoon) GetWord(key string, uid string, r *http.Request) (Word, error) {
	g := goon.NewGoon(r)

	w := new(WordGoon)
	w.Id = key

	if err := g.Get(w); err != nil {
		c := appengine.NewContext(r)
		log.Infof(c, "%v", err)
		return Word{}, err
	}

	if check_uid && w.Uid != uid {
		return Word{}, errors.New("uid invalid")
	}

	v := Word{
		Id: key,
		Uid: uid,
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

func (db *wordDbGoon) GetAll(uid string, r *http.Request) ([]Word, error) {
	g := goon.NewGoon(r)

	words := []WordGoon{}
	if _, err := g.GetAll(datastore.NewQuery("WordGoon")/*.Filter("Uid =", uid)*/, &words); err != nil {
		c := appengine.NewContext(r)
		log.Infof(c, "%v", err)
		return []Word{}, err
	}


	ws := []Word{}
	for _, w := range words {
		v := Word{
			Id: w.Id,
			Uid: w.Uid,
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

func (db *wordDbGoon) AddWord(uid string, w PostWord, r *http.Request) (string, error) {

  reg, _ := regexp.Compile("/ /")
  replaced := reg.ReplaceAllString(w.Text, "_")

	uuid, err1 := uuid.NewUUID()
 	if err1 != nil {
		c := appengine.NewContext(r)
    log.Infof(c, "%v", err1)
		return "", err1
	}
	key := replaced + "_" + string(uuid.String()[0:5])

	wg := WordGoon{
		Id:      key,
		Uid:		uid,
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
	c := appengine.NewContext(r)
	log.Infof(c, "%v", wg)


	return key, nil
}

func (db *wordDbGoon) EditWord(id string, uid string, ew EditWord, r *http.Request) (Word, error) {

	w, err := db.GetWord(id, uid, r)
	if err != nil {
		c := appengine.NewContext(r)
		log.Infof(c, "%v", err)
		return Word{}, err
	}

	if check_uid && w.Uid != uid {
		return Word{}, errors.New("uid invalid")
	}

	if (ew.Kind!="memo") {
		ew.Memo = w.Memo
	}
	if (ew.Kind!="tag") {
		ew.Tag = w.Tag
	}
	if (ew.Kind!="is_review") {
		ew.IsReview = w.IsReview
	}
	if (ew.Kind!="is_input") {
		ew.IsInput = w.IsInput
	}

	wg := WordGoon{
		Id:      id,
		Uid:		uid,
		Text: w.Text,
		Memo: ew.Memo,
		Tag: ew.Tag,
		IsReview: ew.IsReview,
		IsInput: ew.IsInput,
		Count: ew.Count,
		Priority: ew.Priority,
		UpdatedAt: time.Now(),
	}

	g := goon.NewGoon(r)
	if _, err := g.Put(&wg); err != nil {
		c := appengine.NewContext(r)
		log.Infof(c, "%v", err)
		return Word{}, err
	}

	w2, err := db.GetWord(id, uid, r)
	return w2, err
}


func (db *wordDbGoon) Delete(id string, uid string, r *http.Request) error {
	g := goon.NewGoon(r)

	w, err := db.GetWord(id, uid, r)
	if err != nil {
		c := appengine.NewContext(r)
		log.Infof(c, "%v", err)
		return err
	}

	if check_uid && w.Uid != uid {
		return errors.New("uid invalid")
	}

	wkey := new(WordGoon)
	wkey.Id = id
	key, err := g.KeyError(wkey)
	if err != nil {
		return err
	}

	err2 := g.Delete(key)
	return err2
}


func (db *wordDbGoon) Close() error {
	return nil
}
