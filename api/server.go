package main

import (
	"net/http"
	"gopkg.in/gin-gonic/gin.v1"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
	"time"
)

type WordDb interface {
	GetAll(*http.Request) ([]Word, error)

	AddWord(PostWord, *http.Request) (string, error)

	Close() error
}

type (
	Word struct {
		Id    int 	`json:"id"`
		Text string	`json:"text"`
		Memo string `json:"memo"`
		Count int	 `json:"count"`
		Priority int `json:"priority"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	PostWord struct {
		Text     string `form:"text" json:"text" binding:"required"`
	}

	EditWord struct {
		Memo     string `form:"memo" json:"memo"`
		Count     string `form:"count" json:"count"`
		Priority  string `form:"priority" json:"priority"`
	}
)

var (
	db WordDb
)

func words(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	if all, err := db.GetAll(c.Request); err == nil {
		c.JSON(http.StatusOK, all)
	} else {
		c.JSON(http.StatusInternalServerError, "error")
	}
}

func create(c *gin.Context) {
  var json PostWord
	c.Header("Access-Control-Allow-Origin", "*")
	if c.BindJSON(&json) == nil {
		log.Infof(appengine.NewContext(c.Request), "post:%v", json)
		db.AddWord(json, c.Request)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
	}
}

func init() {
	db = newDbMem() // newDbGoon()

	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.GET("/v1/words.json", words)
	r.OPTIONS("/v1/word.json", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
	})
	r.POST("/v1/word.json", create)

	http.Handle("/v1/", r)
}
