// https://godoc.org/github.com/mjibson/goon
package main

import (
	"errors"
	"github.com/google/uuid"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"regexp"
	"time"
)

type ContentGoon struct {
	Id         string         `datastore:"-" goon:"id"`
	Uid        *datastore.Key `datastore:"-" goon:"parent"`
	Text       string         `datastore:"text"`
	Memo       string         `datastore:"memo"`
	IsReview   bool           `datastore:"is_review"`
	IsInput    bool           `datastore:"is_input"`
	Count      int            `datastore:"count"`
	Priority   int            `datastore:"priority"`
	CreatedAt  time.Time      `datastore:"created_at"`
	UpdatedAt  time.Time      `datastore:"updated_at"`
	ReviewedAt time.Time      `datastore:"reviewed_at"`
}

type ContentDb struct {
}

func (db *ContentDb) GetProfileKey(uid string, c context.Context) (*datastore.Key, error) {
	g := goon.FromContext(c)
	pkey := ProfileGoon{Uid: uid}
	if uid_key, err := g.KeyError(&pkey); err != nil {
		log.Debugf(c, "%v", err)
		return nil, err
	} else {
		return uid_key, nil
	}
}

func (db *ContentDb) Get(key string, uid string, c context.Context) (Content, error) {
	g := goon.FromContext(c)

	uid_key, err := db.GetProfileKey(uid, c)
	if err != nil {
		return Content{}, err
	}

	w := ContentGoon{
		Id: key,
		Uid: uid_key,
	}
	if err := g.Get(w); err != nil {
		log.Debugf(c, "%v", err)
		return Content{}, err
	}

	if w.Uid != uid_key {
		return Content{}, errors.New("uid invalid")
	}

	v := Content{
		Id:         key,
		Text:       w.Text,
		Memo:       w.Memo,
		IsReview:   w.IsReview,
		IsInput:    w.IsInput,
		Count:      w.Count,
		Priority:   w.Priority,
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  w.UpdatedAt,
		ReviewedAt: w.ReviewedAt,
	}

	return v, nil
}

func (db *ContentDb) GetAll(uid string, is_review bool, duration_s string, c context.Context) ([]Content, error) {

	uid_key, err := db.GetProfileKey(uid, c)
	if err != nil {
		return []Content{}, err
	}

	filter := datastore.NewQuery("ContentGoon").Ancestor(uid_key)
	if is_review {
		filter = filter.Filter("is_review =", true)
	}

	if duration_s != "" {
		d, err := time.ParseDuration(duration_s)
		if err != nil {
			log.Debugf(c, "%v duration:%v", err, duration_s)
			return []Content{}, err
		}
		filter = filter.Filter("reviewed_at <", time.Now().Add(time.Duration(-1)*d)).Order("reviewed_at")
	}

	filter = filter.Order("-created_at").Limit(100).Offset(0)

	contents := []ContentGoon{}
	g := goon.FromContext(c)
	if _, err := g.GetAll(filter, &contents); err != nil {
		log.Debugf(c, "%v", err)
		return []Content{}, err
	}

	ws := []Content{}
	for _, w := range contents {
		v := Content{
			Id:         w.Id,
			Text:       w.Text,
			Memo:       w.Memo,
			IsReview:   w.IsReview,
			IsInput:    w.IsInput,
			Count:      w.Count,
			Priority:   w.Priority,
			CreatedAt:  w.CreatedAt,
			UpdatedAt:  w.UpdatedAt,
			ReviewedAt: w.ReviewedAt,
		}
		ws = append(ws, v)
	}

	return ws, nil
}

func (db *ContentDb) GetPublicAll(uid string, c context.Context) ([]Content, error) {
	if all, err := db.GetAll(uid, false, "", c); err != nil {
		return []Content{}, err
	} else {
		ws := []Content{}
		for _, w := range all {
			v := Content{
				Id:         w.Id,
				Text:       w.Text,
				Memo:       "",
				IsReview:   false,
				IsInput:    false,
				Count:      w.Count,
				Priority:   w.Priority,
				CreatedAt:  w.CreatedAt,
				UpdatedAt:  w.UpdatedAt,
				ReviewedAt: w.ReviewedAt,
			}
			ws = append(ws, v)
		}

		return ws, nil
	}
}

func (db *ContentDb) GenId(content string, c context.Context) (string, error) {
	reg, _ := regexp.Compile("/ /")
	replaced := reg.ReplaceAllString(content, "_")

	uuid, err1 := uuid.NewUUID()
	if err1 != nil {
		log.Debugf(c, "%v", err1)
		return "", err1
	}
	key := replaced + "_" + string(uuid.String()[0:5])

	return key, nil
}

func (db *ContentDb) Add(uid string, w PostContent, c context.Context) (string, error) {

	if w.Text == "" {
		return "", errors.New("empty")
	}

	key, err1 := db.GenId(w.Text, c)
	if err1 != nil {
		log.Debugf(c, "%v", err1)
		return "", err1
	}

	g := goon.FromContext(c)

	uid_key, err := db.GetProfileKey(uid, c)
	if err != nil {
		return "", err
	}

	wg := ContentGoon{
		Id:         key,
		Uid:        uid_key,
		Text:       w.Text,
		Memo:       "",
		IsReview:   true,
		IsInput:    true,
		Count:      0,
		Priority:   0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ReviewedAt: time.Now(),
	}

	if _, err := g.Put(&wg); err != nil {
		log.Debugf(c, "%v", err)
		return "", err
	}
	log.Debugf(c, "%v", wg)

	return key, nil
}

func (db *ContentDb) Edit(id string, uid string, ew EditContent, c context.Context) (Content, error) {

	g := goon.FromContext(c)

	uid_key, err := db.GetProfileKey(uid, c)
	if err != nil {
		return Content{}, err
	}

	w := ContentGoon{
		Id: id,
		Uid: uid_key,
	}
	if err := g.Get(w); err != nil {
		log.Debugf(c, "edit:%v", err)
		return Content{}, err
	}

	if w.Uid != uid_key {
		return Content{}, errors.New("uid invalid")
	}

	if ew.Kind != "memo" {
		ew.Memo = w.Memo
	}
	if ew.Kind != "is_review" {
		ew.IsReview = w.IsReview
	}
	if ew.Kind != "is_input" {
		ew.IsInput = w.IsInput
	}
	if ew.Kind != "reviewed_at" {
		ew.ReviewedAt = w.ReviewedAt
	} else {
		ew.ReviewedAt = time.Now()
	}

	wg := ContentGoon{
		Id:         id,
		Uid:        uid_key,
		Text:       w.Text,
		Memo:       ew.Memo,
		IsReview:   ew.IsReview,
		IsInput:    ew.IsInput,
		Count:      ew.Count,
		Priority:   ew.Priority,
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  time.Now(),
		ReviewedAt: ew.ReviewedAt,
	}

	if _, err := g.Put(&wg); err != nil {
		log.Debugf(c, "%v", err)
		return Content{}, err
	}

	w2, err := db.Get(id, uid, c)
	log.Debugf(c, "updated:%v", w2)
	return w2, err
}

func (db *ContentDb) Copy(sid string, suid string, duid string, c context.Context) (Content, error) {

	g := goon.FromContext(c)

	suid_key, err := db.GetProfileKey(suid, c)
	if err != nil {
		return Content{}, err
	}

	w := ContentGoon{
		Id: sid,
		Uid: suid_key,
	}
	if err := g.Get(w); err != nil {
		log.Debugf(c, "edit:%v", err)
		return Content{}, err
	}

	new_id, err1 := db.GenId(w.Text, c)
	if err1 != nil {
		log.Debugf(c, "%v", err1)
		return Content{}, err1
	}

	duid_key, err := db.GetProfileKey(duid, c)
	if err != nil {
		return Content{}, err
	}

	wg := ContentGoon{
		Id:         new_id,
		Uid:        duid_key,
		Text:       w.Text,
		Memo:       "",
		IsReview:   false,
		IsInput:    false,
		Priority:   w.Priority,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ReviewedAt: time.Now(),
	}

	if _, err := g.Put(&wg); err != nil {
		log.Debugf(c, "%v", err)
		return Content{}, err
	}

	w2, err := db.Get(new_id, duid, c)
	log.Debugf(c, "updated:%v", w2)
	return w2, err
}

func (db *ContentDb) Delete(id string, uid string, c context.Context) error {
	g := goon.FromContext(c)

	uid_key, err := db.GetProfileKey(uid, c)
	if err != nil {
		return err
	}

	w := ContentGoon{
		Id: id,
		Uid: uid_key,
	}
	if err := g.Get(w); err != nil {
		log.Debugf(c, "couldn't find:%v", err)
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
