// https://godoc.org/github.com/mjibson/goon
package main

import (
	"time"
	"net/http"
	"github.com/mjibson/goon"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"errors"
)

type ProfileGoon struct {
	Uid string `datastore:"-" goon:"id"` // id
	UserName string	`datastore:"user_name"` // user name
	CreatedAt time.Time `datastore:"created_at"`
	// TODO LastLoginedAt 
	Disabled bool `datastore:"disabled"`
}

type UserDb struct {
}

func (db *UserDb) GetUidByUser(user string, r *http.Request) (string, error) {
	g := goon.NewGoon(r)

	profiles := []ProfileGoon{}
	if _, err := g.GetAll(datastore.NewQuery("ProfileGoon").Filter("user_name =", user), &profiles); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return "", err
	}

	if len(profiles) == 0 {
		log.Debugf(appengine.NewContext(r), "not found user %v", user)
		return "", errors.New("not found user")
	}

	if profiles[0].Disabled {
		log.Debugf(appengine.NewContext(r), "user is now disabled%v", user)
		return "", errors.New("user is disabled")
	}

	return profiles[0].Uid, nil
}

func (db *UserDb) GetUser(uid string, r *http.Request) (string, error) {
	g := goon.NewGoon(r)
	p := ProfileGoon{Uid: uid}
	if err := g.Get(&p); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return "", err
	} else {
	    if p.Disabled {
			log.Debugf(appengine.NewContext(r), "user is now disabled%v", p)
			return "", errors.New("user is disabled")
		} else {
			log.Debugf(appengine.NewContext(r), "login with %v", p)
			return p.UserName, nil
		}
	}
}

func (db *UserDb) NewUser(uid string, user string, r *http.Request) error {
	g := goon.NewGoon(r)

	// TODO validate username

	profiles := []ProfileGoon{}
	if _, err := g.GetAll(datastore.NewQuery("ProfileGoon").Filter("user_name =", user), &profiles); err != nil {
		log.Debugf(appengine.NewContext(r), "%v", err)
		return err
	}

	if len(profiles) != 0 {
		return errors.New("already in")
	}

	pkey := ProfileGoon{
		Uid: uid,
		UserName: user,
		CreatedAt: time.Now(),
		Disabled: false,
	}
	_, err := g.Put(&pkey)

	return err
}
