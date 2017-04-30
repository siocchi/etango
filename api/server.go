package main

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

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

	Status struct {
		Status  string
	}

	UserNameAndScreenName struct {
		UserName string
		ScreenName string
	}
)

var (
	db     ContentDb
	userDb UserDb
)

func contents(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	profile := profileFromSession(ctx)
	if profile == nil {
		return c.JSON(http.StatusUnauthorized, Status{Status: "unauthorized",})
	}
	is_review := c.QueryParam("is_review") == "true"
	duration := c.QueryParam("duration")

	if all, err := db.GetAll(profile.ID, is_review, duration, ctx); err == nil {
		return c.JSON(http.StatusOK, all)
	} else {
		return c.JSON(http.StatusBadRequest, "error")
	}
}

func publicContents(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	uid, err := userDb.GetUidByUser(c.Param("user"), ctx)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Status{Status: "not found user"})
	}

	if all, err := db.GetPublicAll(uid, ctx); err != nil {
		return c.JSON(http.StatusBadRequest, Status{Status: "error"})
	} else {
		return c.JSON(http.StatusOK, all)
	}
}

func create(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())

	profile := profileFromSession(ctx)
	if profile == nil {
		return c.JSON(http.StatusUnauthorized, Status{Status: "unauthorized"})
	}

	var json PostContent
	if c.Bind(&json) == nil {
		log.Debugf(ctx, "post:%v", json)
		db.Add(profile.ID, json, ctx)
		return c.JSON(http.StatusOK, Status{Status: "ok"})
	} else {
		return c.JSON(http.StatusBadRequest, Status{Status: "parse error"})
	}
}

func edit(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	profile := profileFromSession(ctx)
	if profile == nil {
		return c.JSON(http.StatusUnauthorized, Status{Status: "unauthorized"})
	}

	var json EditContent
	if c.Bind(&json) != nil {
		return c.JSON(http.StatusBadRequest, Status{Status: "parse error"})
	}
	if w, err := db.Edit(c.Param("id"), profile.ID, json, ctx); err != nil {
		return c.JSON(http.StatusBadRequest, Status{Status: "something wrong"})
	} else {
		return c.JSON(http.StatusOK, w)
	}
}

func delete(c echo.Context) error {
	id := c.Param("id")
	ctx := appengine.NewContext(c.Request())

	profile := profileFromSession(ctx)
	if profile == nil {
		return c.JSON(http.StatusUnauthorized, Status{Status: "unauthorized"})
	}

	err := db.Delete(id, profile.ID, ctx)
	if err != nil {
		log.Debugf(ctx, "delete error:%v", err)
		return c.JSON(http.StatusBadRequest, Status{Status: "error"})
	} else {
		log.Debugf(ctx, "delete:%v", id)
		return c.JSON(http.StatusOK, Status{Status: "ok"})
	}
}

func createUser(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	profile := profileFromSession(ctx)
	if profile == nil {
		return c.JSON(http.StatusUnauthorized, Status{Status: "unauthorized"})
	}

	var json PostUser
	if c.Bind(&json) != nil {
		return c.JSON(http.StatusBadRequest, Status{Status: "bad request; json error"})
	}

	if err := userDb.NewUser(profile.ID, json.User, ctx); err != nil {
		log.Debugf(ctx, "new user :%v", err)
		return c.JSON(http.StatusBadRequest, Status{Status: "bad request"})
	}

	return c.JSON(http.StatusOK, Status{Status: "success"})
}

func deleteUser(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	profile := profileFromSession(ctx)
	if profile == nil {
		return c.JSON(http.StatusUnauthorized, Status{Status: "unauthorized"})
	}

	if err := userDb.DisableUser(profile.ID, ctx); err != nil {
		return c.JSON(http.StatusBadRequest, Status{Status: "bad request"})
	}

	return c.JSON(http.StatusOK, Status{Status: "success"})
}

func profile(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())

	var profile = profileFromSession(ctx)

	if profile == nil {
		return c.JSON(http.StatusUnauthorized, Status{Status: "unauthorized"})
	}

	user, err := userDb.GetUserName(profile.ID, ctx)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, Status{Status: "unregistered"})
	}
	return c.JSON(http.StatusOK, UserNameAndScreenName{UserName: user, ScreenName: profile.DisplayName})
}

func login(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, "/signup")
		return c.Redirect(http.StatusMovedPermanently, url)
	}

	_, err := userDb.GetUserName(u.ID, ctx)
	if err != nil {
		return c.Redirect(http.StatusMovedPermanently, "/signup")
	}

	if err := userDb.Login(u.ID, ctx); err != nil {
		return c.JSON(http.StatusUnauthorized, Status{Status: "unauthorized"})
	}

	return c.Redirect(http.StatusMovedPermanently, "/home")
}

func logout(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	u := user.Current(ctx)
	if u == nil {
		return c.Redirect(http.StatusMovedPermanently, "/")
	}

	url, _ := user.LogoutURL(ctx, "/")
	return c.Redirect(http.StatusMovedPermanently, url)
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
	e := echo.New()
	g := e.Group("")
	g.Use(middleware.CORS())

	g.GET("/v1/words.json", contents)
	g.POST("/v1/word.json", create)

	g.POST("/v1/word/:id/edit.json", edit)
	g.DELETE("/v1/word/:id/edit.json", delete)

	g.POST("/v1/create_user.json", createUser)
	g.POST("/v1/delete_user.json", deleteUser)
	g.GET("/v1/profile.json", profile)
	g.GET("/v1/user/:user/words.json", publicContents)

	g.GET("/v1/login", login)
	g.GET("/v1/logout", logout)

	http.Handle("/v1/", e)
}
