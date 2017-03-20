// https://godoc.org/github.com/mjibson/goon
package main

import (
	"errors"
	"github.com/mjibson/goon"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"golang.org/x/net/context"
	"time"
)

type ProfileGoon struct {
	Uid       string    `datastore:"-" goon:"id"` // id
	UserName  string    `datastore:"user_name"`   // user name
	CreatedAt time.Time `datastore:"created_at"`
	LastLoginedAt time.Time `datastore:"last_login_at"`
	Disabled bool `datastore:"disabled"`
}

type UserDb struct {
}

func (db *UserDb) GetUidByUser(user string, c context.Context) (string, error) {
	g := goon.FromContext(c)

	profiles := []ProfileGoon{}
	if _, err := g.GetAll(datastore.NewQuery("ProfileGoon").Filter("user_name =", user), &profiles); err != nil {
		log.Debugf(c, "%v", err)
		return "", err
	}

	if len(profiles) == 0 {
		log.Debugf(c, "not found user %v", user)
		return "", errors.New("not found user")
	}

	if profiles[0].Disabled {
		log.Debugf(c, "user is now disabled%v", user)
		return "", errors.New("user is disabled")
	}

	return profiles[0].Uid, nil
}

func (db *UserDb) GetUser(uid string, c context.Context) (string, error) {
	g := goon.FromContext(c)
	p := ProfileGoon{Uid: uid}
	if err := g.Get(&p); err != nil {
		log.Debugf(c, "%v", err)
		return "", err
	} else {
		if p.Disabled {
			log.Debugf(c, "user is now disabled%v", p)
			return "", errors.New("user is disabled")
		} else {
			log.Debugf(c, "login with %v", p)
			return p.UserName, nil
		}
	}
}

func (db *UserDb) NewUser(uid string, user string, c context.Context) error {
	g := goon.FromContext(c)

	// TODO validate username

	profiles := []ProfileGoon{}
	if _, err := g.GetAll(datastore.NewQuery("ProfileGoon").Filter("user_name =", user), &profiles); err != nil {
		log.Debugf(c, "%v", err)
		return err
	}

	if len(profiles) != 0 {
		return errors.New("already in")
	}

	pkey := ProfileGoon{
		Uid:       uid,
		UserName:  user,
		CreatedAt: time.Now(),
		LastLoginedAt: time.Now(),
		Disabled:  false,
	}
	_, err := g.Put(&pkey)

	return err
}

func (db *UserDb) Login(uid string, c context.Context) error {
	g := goon.FromContext(c)

	p := ProfileGoon{Uid: uid}
	if err := g.Get(&p); err != nil {
		log.Debugf(c, "%v", err)
		return err
	}

	pkey := ProfileGoon{
		Uid:       uid,
		UserName:  p.UserName,
		CreatedAt: p.CreatedAt,
		LastLoginedAt: time.Now(),
		Disabled:  p.Disabled,
	}
	_, err := g.Put(&pkey)

	return err
}

func (db *UserDb) DisableUser(uid string, c context.Context) error {
	g := goon.FromContext(c)

	p := ProfileGoon{Uid: uid}
	if err := g.Get(&p); err != nil {
		log.Debugf(c, "%v", err)
		return err
	}

	pkey := ProfileGoon{
		Uid:       uid,
		UserName:  p.UserName,
		CreatedAt: p.CreatedAt,
		LastLoginedAt: p.LastLoginedAt,
		Disabled:  true,
	}
	_, err := g.Put(&pkey)

	return err
}
