
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
	Uid  *datastore.Key `datastore:"-" goon:"parent"`
	Text string	`datastore:"text"`
	Memo string `datastore:"memo"`
	Tag	 string `datastore:"tag"`
	IsReview bool `datastore:"is_review"`
	IsInput bool `datastore:"is_input"`
	Count int	 `datastore:"count"`
	Priority int `datastore:"priority"`
	UpdatedAt time.Time `datastore:"updated_at"`
	ReviewedAt time.Time `datastore:"reviewed_at"`
}

type ProfileGoon struct {
	Uid string `datastore:"-" goon:"id"` // id
	UserName string	`datastore:"user_name"` // user name
}

type wordDbGoon struct {
	goon string
}

func newDbGoon() *wordDbGoon {
	return &wordDbGoon{goon: ""}
}

func (db *wordDbGoon) GetProfileKey(uid string, r *http.Request) (*datastore.Key, error) {
	g := goon.NewGoon(r)
	pkey := ProfileGoon{Uid: uid}
	if uid_key, err := g.KeyError(&pkey); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return nil, err
	} else {
		return uid_key, nil
	}
}

func (db *wordDbGoon) GetWord(key string, uid string, r *http.Request) (Word, error) {
	g := goon.NewGoon(r)

	uid_key, err := db.GetProfileKey(uid, r)
	if err != nil {
		return Word{}, err
	}

	w := new(WordGoon)
	w.Id = key
	w.Uid = uid_key

	if err := g.Get(w); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return Word{}, err
	}

	if w.Uid != uid_key {
		return Word{}, errors.New("uid invalid")
	}

	v := Word{
		Id: key,
		Text: w.Text,
		Memo: w.Memo,
		Tag: w.Tag,
		IsReview: w.IsReview,
		IsInput: w.IsInput,
		Count: w.Count,
		Priority: w.Priority,
		UpdatedAt: w.UpdatedAt,
		ReviewedAt: w.ReviewedAt,
	}

	return v, nil
}

func (db *wordDbGoon) GetAll(uid string, is_review bool, duration_s string, r *http.Request) ([]Word, error) {

	uid_key, err := db.GetProfileKey(uid, r)
	if err != nil {
		return []Word{}, err
	}

	filter := datastore.NewQuery("WordGoon").Ancestor(uid_key)
	if (is_review) {
		filter = filter.Filter("is_review =", true)
	}

	if (duration_s != "") {
		d, err := time.ParseDuration(duration_s)
		if err!=nil {
			log.Debugf(appengine.NewContext(r), "%v duration:%v", err, duration_s)
			return []Word{}, err
		}
		filter = filter.Filter("reviewed_at <", time.Now().Add(time.Duration(-1)*d))
	}

	words := []WordGoon{}
	g := goon.NewGoon(r)
	if _, err := g.GetAll(filter, &words); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return []Word{}, err
	}

	ws := []Word{}
	for _, w := range words {
		v := Word{
			Id: w.Id,
			Text: w.Text,
			Memo: w.Memo,
			Tag: w.Tag,
			IsReview: w.IsReview,
			IsInput: w.IsInput,
			Count: w.Count,
			Priority: w.Priority,
			UpdatedAt: w.UpdatedAt,
			ReviewedAt: w.ReviewedAt,
		}
		ws = append(ws, v)
	}

	return ws, nil
}

func (db *wordDbGoon) GenId(word string, r *http.Request) (string, error) {
	reg, _ := regexp.Compile("/ /")
	replaced := reg.ReplaceAllString(word, "_")

	uuid, err1 := uuid.NewUUID()
	if err1 != nil {
		log.Debugf(appengine.NewContext(r), "%v", err1)
		return "", err1
	}
	key := replaced + "_" + string(uuid.String()[0:5])

	return key, nil
}

func (db *wordDbGoon) AddWord(uid string, w PostWord, r *http.Request) (string, error) {

	key, err1 := db.GenId(w.Text, r)
	if err1 != nil {
		log.Debugf(appengine.NewContext(r), "%v", err1)
		return "", err1
	}

	g := goon.NewGoon(r)

	uid_key, err := db.GetProfileKey(uid, r)
	if err != nil {
		return "", err
	}

	wg := WordGoon{
		Id:   key,
		Uid:  uid_key,
		Text: w.Text,
		Memo: "",
		Tag:  "",
		IsReview: true,
		IsInput: true,
		Count: 0,
		Priority: 0,
		UpdatedAt: time.Now(),
		ReviewedAt: time.Now(),
	}

	if _, err := g.Put(&wg); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return "", err
	}
	log.Debugf(appengine.NewContext(r), "%v", wg)

	return key, nil
}

func (db *wordDbGoon) EditWord(id string, uid string, ew EditWord, r *http.Request) (Word, error) {

	g := goon.NewGoon(r)

	uid_key, err := db.GetProfileKey(uid, r)
	if err != nil {
		return Word{}, err
	}

	w := new(WordGoon)
	w.Id = id
	w.Uid = uid_key
	if err := g.Get(w); err != nil {
		log.Debugf(appengine.NewContext(r), "edit:%v", err)
		return Word{}, err
	}

	if w.Uid != uid_key {
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
	if (ew.Kind!="reviewed_at") {
		ew.ReviewedAt = w.ReviewedAt
	} else {
		ew.ReviewedAt = time.Now()
	}

	wg := WordGoon{
		Id:   id,
		Uid:  uid_key,
		Text: w.Text,
		Memo: ew.Memo,
		Tag: ew.Tag,
		IsReview: ew.IsReview,
		IsInput: ew.IsInput,
		Count: ew.Count,
		Priority: ew.Priority,
		UpdatedAt: time.Now(),
		ReviewedAt: ew.ReviewedAt,
	}

	if _, err := g.Put(&wg); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return Word{}, err
	}

	w2, err := db.GetWord(id, uid, r)
	log.Debugf(appengine.NewContext(r), "updated:%v", w2)
	return w2, err
}

func (db *wordDbGoon) Delete(id string, uid string, r *http.Request) error {
	g := goon.NewGoon(r)

	uid_key, err := db.GetProfileKey(uid, r)
	if err != nil {
		return err
	}

	w := new(WordGoon)
	w.Id = id
	w.Uid = uid_key
	if err := g.Get(w); err != nil {
		log.Debugf(appengine.NewContext(r), "delete:%v", err)
		return err
	}

	if w.Uid != uid_key {
		return errors.New("uid invalid")
	}

	wkey := new(WordGoon)
	wkey.Id = id
	wkey.Uid = uid_key
	key, err := g.KeyError(wkey)
	if err != nil {
		return err
	}

	err2 := g.Delete(key)
	return err2
}

func (db *wordDbGoon) GetUser(uid string, r *http.Request) (string, error) {
	g := goon.NewGoon(r)
	p := ProfileGoon{Uid: uid}
	if err := g.Get(&p); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return "", err
	} else {
		return p.UserName, nil
	}

}


// TODO rename
func (db *wordDbGoon) SignUp(uid string, user string, r *http.Request) error {
	g := goon.NewGoon(r)

	// TODO validate username

	pkey := ProfileGoon{Uid: uid, UserName: user}
	_, err := g.Put(&pkey)

	return err
}

func (db *wordDbGoon) Close() error {
	return nil
}
