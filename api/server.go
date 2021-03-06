package main

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
	"gopkg.in/gin-gonic/gin.v1"
	"net/http"
	"time"
)

type (
	Content struct {
		Id         string    `json:"id"`
		Text       string    `json:"text"`
		Memo       string    `json:"memo"`
		IsReview   bool      `json:"is_review"`
		IsInput    bool      `json:"is_input"`
		Count      int       `json:"count"`
		Priority   int       `json:"priority"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
		ReviewedAt time.Time `json:"reviewed_at"`
	}

	PostContent struct {
		Text string `form:"text" json:"text" binding:"required"`
	}

	EditContent struct {
		Kind       string    `form:"kind" json:"kind"`
		Memo       string    `form:"memo" json:"memo"`
		IsReview   bool      `json:"is_review"`
		IsInput    bool      `json:"is_input"`
		Count      int       `form:"count" json:"count"`
		Priority   int       `form:"priority" json:"priority"`
		ReviewedAt time.Time `json:"reviewed_at"`
	}

	PostUser struct {
		User string `form:"user" json:"user" binding:"required"`
	}

	Profile struct {
		ID, DisplayName string
	}
)

var (
	db     ContentDb
	userDb UserDb
)

func contents(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)
	profile := profileFromSession(ctx)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}
	is_review := c.Query("is_review") == "true"
	duration := c.Query("duration")

	c.Header("Access-Control-Allow-Origin", "*")
	if all, err := db.GetAll(profile.ID, is_review, duration, ctx); err == nil {
		c.JSON(http.StatusOK, all)
	} else {
		c.JSON(http.StatusBadRequest, "error")
	}
}

func publicContents(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)
	uid, err := userDb.GetUidByUser(c.Param("user"), ctx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "not found user"})
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	if all, err := db.GetPublicAll(uid, ctx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
	} else {
		c.JSON(http.StatusOK, all)
	}
}

func create(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	ctx := appengine.NewContext(c.Request)

	profile := profileFromSession(ctx)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	var json PostContent
	if c.BindJSON(&json) == nil {
		log.Debugf(ctx, "post:%v", json)
		db.Add(profile.ID, json, ctx)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "parse error"})
	}
}

func edit(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	ctx := appengine.NewContext(c.Request)
	profile := profileFromSession(ctx)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	var json EditContent
	if c.BindJSON(&json) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "parse error"})
		return
	}
	if w, err := db.Edit(c.Param("id"), profile.ID, json, ctx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "something wrong"})
	} else {
		c.JSON(http.StatusOK, w)
	}
}

func delete(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	id := c.Param("id")
	ctx := appengine.NewContext(c.Request)

	profile := profileFromSession(ctx)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	err := db.Delete(id, profile.ID, ctx)
	if err != nil {
		log.Debugf(ctx, "delete error:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
	} else {
		log.Debugf(ctx, "delete:%v", id)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func createUser(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)
	profile := profileFromSession(ctx)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	var json PostUser
	if c.BindJSON(&json) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request; json error"})
		return
	}

	if err := userDb.NewUser(profile.ID, json.User, ctx); err != nil {
		log.Debugf(ctx, "new user :%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func deleteUser(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)
	profile := profileFromSession(ctx)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	if err := userDb.DisableUser(profile.ID, ctx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func profile(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	ctx := appengine.NewContext(c.Request)

	var profile = profileFromSession(ctx)

	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	user, err := userDb.GetUserName(profile.ID, ctx)
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

	_, err := userDb.GetUserName(u.ID, ctx)
	if err != nil {
		c.Redirect(http.StatusMovedPermanently, "/signup")
		return
	}

	if err := userDb.Login(u.ID, ctx); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/home")
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

func profileFromSession(ctx context.Context) *Profile {
	u := user.Current(ctx)
	if u == nil {
		return nil
	}

	return &Profile{
		ID:          u.ID,
		DisplayName: u.String(),
	}
}

func init() {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.GET("/v1/words.json", contents)
	r.POST("/v1/word.json", create)

	r.POST("/v1/word/:id/edit.json", edit)
	r.DELETE("/v1/word/:id/edit.json", delete)

	r.POST("/v1/create_user.json", createUser)
	r.POST("/v1/delete_user.json", deleteUser)
	r.GET("/v1/profile.json", profile)
	r.GET("/v1/user/:user/words.json", publicContents)

	r.GET("/v1/login", login)
	r.GET("/v1/logout", logout)

	http.Handle("/v1/", r)
}
