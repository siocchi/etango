
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

type ContentGoon struct {
	Id    string 	`datastore:"-" goon:"id"`
	Uid  *datastore.Key `datastore:"-" goon:"parent"`
	Text string	`datastore:"text"`
	Memo string `datastore:"memo"`
	IsReview bool `datastore:"is_review"`
	IsInput bool `datastore:"is_input"`
	Count int	 `datastore:"count"`
	Priority int `datastore:"priority"`
	CreatedAt time.Time `datastore:"created_at"`
	UpdatedAt time.Time `datastore:"updated_at"`
	ReviewedAt time.Time `datastore:"reviewed_at"`
}

type ContentDb struct {
}

func (db *ContentDb) GetProfileKey(uid string, r *http.Request) (*datastore.Key, error) {
	g := goon.NewGoon(r)
	pkey := ProfileGoon{Uid: uid}
	if uid_key, err := g.KeyError(&pkey); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return nil, err
	} else {
		return uid_key, nil
	}
}

func (db *ContentDb) Get(key string, uid string, r *http.Request) (Content, error) {
	g := goon.NewGoon(r)

	uid_key, err := db.GetProfileKey(uid, r)
	if err != nil {
		return Content{}, err
	}

	w := new(ContentGoon)
	w.Id = key
	w.Uid = uid_key

	if err := g.Get(w); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return Content{}, err
	}

	if w.Uid != uid_key {
		return Content{}, errors.New("uid invalid")
	}

	v := Content{
		Id: key,
		Text: w.Text,
		Memo: w.Memo,
		IsReview: w.IsReview,
		IsInput: w.IsInput,
		Count: w.Count,
		Priority: w.Priority,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
		ReviewedAt: w.ReviewedAt,
	}

	return v, nil
}

func (db *ContentDb) GetAll(uid string, is_review bool, duration_s string, r *http.Request) ([]Content, error) {

	uid_key, err := db.GetProfileKey(uid, r)
	if err != nil {
		return []Content{}, err
	}

	filter := datastore.NewQuery("ContentGoon").Ancestor(uid_key)
	if (is_review) {
		filter = filter.Filter("is_review =", true)
	}

	if (duration_s != "") {
		d, err := time.ParseDuration(duration_s)
		if err!=nil {
			log.Debugf(appengine.NewContext(r), "%v duration:%v", err, duration_s)
			return []Content{}, err
		}
		filter = filter.Filter("reviewed_at <", time.Now().Add(time.Duration(-1)*d)).Order("reviewed_at")
	}

	filter = filter.Order("-created_at").Limit(100).Offset(0)

	contents := []ContentGoon{}
	g := goon.NewGoon(r)
	if _, err := g.GetAll(filter, &contents); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return []Content{}, err
	}

	ws := []Content{}
	for _, w := range contents {
		v := Content{
			Id: w.Id,
			Text: w.Text,
			Memo: w.Memo,
			IsReview: w.IsReview,
			IsInput: w.IsInput,
			Count: w.Count,
			Priority: w.Priority,
			CreatedAt: w.CreatedAt,
			UpdatedAt: w.UpdatedAt,
			ReviewedAt: w.ReviewedAt,
		}
		ws = append(ws, v)
	}

	return ws, nil
}

func (db *ContentDb) GetPublicAll(uid string, r *http.Request) ([]Content, error) {
	if all, err := db.GetAll(uid, false, "", r); err != nil {
		return []Content{}, err
	} else {
		ws := []Content{}
		for _, w := range all {
			v := Content{
				Id: w.Id,
				Text: w.Text,
				Memo: "",
				IsReview: false,
				IsInput: false,
				Count: w.Count,
				Priority: w.Priority,
				CreatedAt: w.CreatedAt,
				UpdatedAt: w.UpdatedAt,
				ReviewedAt: w.ReviewedAt,
			}
			ws = append(ws, v)
		}

		return ws, nil
	}
}

func (db *ContentDb) GenId(content string, r *http.Request) (string, error) {
	reg, _ := regexp.Compile("/ /")
	replaced := reg.ReplaceAllString(content, "_")

	uuid, err1 := uuid.NewUUID()
	if err1 != nil {
		log.Debugf(appengine.NewContext(r), "%v", err1)
		return "", err1
	}
	key := replaced + "_" + string(uuid.String()[0:5])

	return key, nil
}

func (db *ContentDb) Add(uid string, w PostContent, r *http.Request) (string, error) {

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

	wg := ContentGoon{
		Id:   key,
		Uid:  uid_key,
		Text: w.Text,
		Memo: "",
		IsReview: true,
		IsInput: true,
		Count: 0,
		Priority: 0,
		CreatedAt: time.Now(),
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

func (db *ContentDb) Edit(id string, uid string, ew EditContent, r *http.Request) (Content, error) {

	g := goon.NewGoon(r)

	uid_key, err := db.GetProfileKey(uid, r)
	if err != nil {
		return Content{}, err
	}

	w := new(ContentGoon)
	w.Id = id
	w.Uid = uid_key
	if err := g.Get(w); err != nil {
		log.Debugf(appengine.NewContext(r), "edit:%v", err)
		return Content{}, err
	}

	if w.Uid != uid_key {
		return Content{}, errors.New("uid invalid")
	}

	if (ew.Kind!="memo") {
		ew.Memo = w.Memo
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

	wg := ContentGoon{
		Id:   id,
		Uid:  uid_key,
		Text: w.Text,
		Memo: ew.Memo,
		IsReview: ew.IsReview,
		IsInput: ew.IsInput,
		Count: ew.Count,
		Priority: ew.Priority,
		CreatedAt: w.CreatedAt,
		UpdatedAt: time.Now(),
		ReviewedAt: ew.ReviewedAt,
	}

	if _, err := g.Put(&wg); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return Content{}, err
	}

	w2, err := db.Get(id, uid, r)
	log.Debugf(appengine.NewContext(r), "updated:%v", w2)
	return w2, err
}

func (db *ContentDb) Copy(id string, uid string, r *http.Request) (Content, error) {

	g := goon.NewGoon(r)

	uid_key, err := db.GetProfileKey(uid, r)
	if err != nil {
		return Content{}, err
	}

	w := new(ContentGoon)
	w.Id = id
	w.Uid = uid_key
	if err := g.Get(w); err != nil {
		log.Debugf(appengine.NewContext(r), "edit:%v", err)
		return Content{}, err
	}

	new_id, err1 := db.GenId(w.Text, r)
	if err1 != nil {
		log.Debugf(appengine.NewContext(r), "%v", err1)
		return Content{}, err1
	}


	if w.Uid != uid_key {
		return Content{}, errors.New("uid invalid")
	}

	wg := ContentGoon{
		Id:   new_id,
		Uid:  uid_key,
		Text: w.Text,
		Memo: "",
		IsReview: false,
		IsInput: false,
		Priority: w.Priority,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ReviewedAt: time.Now(),
	}

	if _, err := g.Put(&wg); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return Content{}, err
	}

	w2, err := db.Get(new_id, uid, r)
	log.Debugf(appengine.NewContext(r), "updated:%v", w2)
	return w2, err
}

func (db *ContentDb) Delete(id string, uid string, r *http.Request) error {
	g := goon.NewGoon(r)

	uid_key, err := db.GetProfileKey(uid, r)
	if err != nil {
		return err
	}

	w := new(ContentGoon)
	w.Id = id
	w.Uid = uid_key
	if err := g.Get(w); err != nil {
		log.Debugf(appengine.NewContext(r), "delete:%v", err)
		return err
	}

	if w.Uid != uid_key {
		return errors.New("uid invalid")
	}

	wkey := new(ContentGoon)
	wkey.Id = id
	wkey.Uid = uid_key
	key, err := g.KeyError(wkey)
	if err != nil {
		return err
	}

	err2 := g.Delete(key)
	return err2
}
