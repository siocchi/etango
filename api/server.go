package main

import (
	"net/http"
	"gopkg.in/gin-gonic/gin.v1"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
	"time"
	"google.golang.org/appengine/user"
)

type ContentDb interface {
	GetAll(string, bool, string, *http.Request) ([]Content, error)

	GetPublicAll(string, *http.Request) ([]Content, error)

	Add(string, PostContent, *http.Request) (string, error)

	Edit(string, string, EditContent, *http.Request) (Content, error)

	Delete(string, string, *http.Request) error

	Copy(string, string, *http.Request) (Content, error)
}

type UserDb interface {
	NewUser(string, string, *http.Request) error

	GetUidByUser(user string, r *http.Request) (string, error)

	GetUser(string, *http.Request) (string, error)
}

type (
	Content struct {
		Id    string 	`json:"id"`
		Text string	`json:"text"`
		Memo string `json:"memo"`
		IsReview bool `json:"is_review"`
		IsInput bool `json:"is_input"`
		Count int	 `json:"count"`
		Priority int `json:"priority"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		ReviewedAt time.Time `json:"reviewed_at"`
	}

	PostContent struct {
		Text     string `form:"text" json:"text" binding:"required"`
	}

	EditContent struct {
		Kind     string `form:"kind" json:"kind"`
		Memo     string `form:"memo" json:"memo"`
		IsReview bool `json:"is_review"`
		IsInput bool `json:"is_input"`
		Count     int `form:"count" json:"count"`
		Priority  int `form:"priority" json:"priority"`
		ReviewedAt time.Time `json:"reviewed_at"`
	}

	PostUser struct {
		User     string `form:"user" json:"user" binding:"required"`
	}

	Profile struct {
       ID, DisplayName string
	}
)

var (
	db ContentDb
	userDb UserDb
)

func contents(c *gin.Context) {

	profile := profileFromSession(c.Request)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}
	is_review := c.Query("is_review") == "true";
	duration := c.Query("duration");

	c.Header("Access-Control-Allow-Origin", "*")
	if all, err := db.GetAll(profile.ID, is_review, duration, c.Request); err == nil {
		c.JSON(http.StatusOK, all)
	} else {
		c.JSON(http.StatusBadRequest, "error")
	}
}

func publicContents(c *gin.Context) {

	uid, err := userDb.GetUidByUser(c.Param("user"), c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "not found user"})
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	if all, err := db.GetPublicAll(uid, c.Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
	} else {
		c.JSON(http.StatusOK, all)
	}
}

func create(c *gin.Context) {
	var json PostContent
	c.Header("Access-Control-Allow-Origin", "*")

	profile := profileFromSession(c.Request)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	if c.BindJSON(&json) == nil {
		log.Debugf(appengine.NewContext(c.Request), "post:%v", json)
		db.Add(profile.ID, json, c.Request)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "parse error"})
	}
}

func edit(c *gin.Context) {
	var json EditContent

	profile := profileFromSession(c.Request)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	if c.BindJSON(&json) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "parse error"})
		return
	}
	if w, err := db.Edit(c.Param("id"), profile.ID, json, c.Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "something wrong"})
	} else {
		c.JSON(http.StatusOK, w)
	}
}

func delete(c *gin.Context) {
	id := c.Param("id")

	profile := profileFromSession(c.Request)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	err := db.Delete(id, profile.ID, c.Request)
	if err != nil {
		log.Debugf(appengine.NewContext(c.Request), "delete error:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
	} else {
		log.Debugf(appengine.NewContext(c.Request), "delete:%v", id)
		c.JSON(http.StatusOK, gin.H{"status":"ok"})
	}
}

func createUser(c *gin.Context) {
	profile := profileFromSession(c.Request)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	var json PostUser
	if c.BindJSON(&json) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request"})
		return
	}

	if err := userDb.NewUser(profile.ID, json.User, c.Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func profile(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

	var profile = profileFromSession(c.Request)

	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	user, err := userDb.GetUser(profile.ID, c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unregistered"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_name": user, "screen_name": profile.DisplayName})
}

func login(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)
	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, "/signup")
		c.Redirect(http.StatusMovedPermanently, url)
		return
	}
	log.Debugf(appengine.NewContext(c.Request), "user: %v %s", u, u.ID)

	_, err := userDb.GetUser(u.ID, c.Request)
	if err != nil {
		c.Redirect(http.StatusMovedPermanently, "/signup")
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/home")

	// url, _ := user.LogoutURL(ctx, "/v1/logout")
	// fmt.Fprintf(w, `Welcome, %s! (<a href="%s">sign out</a>)`, u, url)
}

func logout(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)
	u := user.Current(ctx)
	if u == nil {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	url, _ := user.LogoutURL(ctx, "/")
	c.Redirect(http.StatusMovedPermanently, url)
}

func profileFromSession(r *http.Request) *Profile {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil {
		return nil
	}

	return &Profile{
		ID: u.ID,
		DisplayName: u.String(),
	}
}

func init() {
	db = newDbGoon()
	userDb = newUserDbGoon()

	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.GET("/v1/words.json", contents)
	r.OPTIONS("/v1/word.json", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
	})
	r.POST("/v1/word.json", create)

	r.OPTIONS("/v1/word/:id/edit.json", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Allow-Methods", "POST, DELETE, OPTIONS")
	})
	r.POST("/v1/word/:id/edit.json", edit)
	r.DELETE("/v1/word/:id/edit.json", delete)

	r.POST("/v1/create_user.json", createUser)
	r.GET("/v1/profile.json", profile)
	r.GET("/v1/user/:user/words.json", publicContents)

	r.GET("/v1/login", login)
	r.GET("/v1/logout", logout)

	http.Handle("/v1/", r)
}
