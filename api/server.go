package main

import (
	"net/http"
	"gopkg.in/gin-gonic/gin.v1"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"

)

type LinkDb interface {
	GetLink(string, *http.Request) (string, error)

	AddLink(string, *http.Request) (string, error)

	Close() error
}

type (
	Word struct {
		Id    int 	`json:"id"`
		Text string	`json:"text"`
	}

	PostWord struct {
		Text     string `form:"text" json:"text" binding:"required"`
	}
)

var (
	db LinkDb
)


func words(c *gin.Context) {
	var js = []word {
		word {
			Id: 1,
			Text: "test",
		},
	}
func create(c *gin.Context) {
  var json PostWord
	c.Header("Access-Control-Allow-Origin", "*")
	if c.BindJSON(&json) == nil {
			log.Infof(appengine.NewContext(c.Request), "post:%v", json)
			db.AddLink(json.Text, c.Request)
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
