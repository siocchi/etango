package main

import (
	"net/http"
	"gopkg.in/gin-gonic/gin.v1"
)

type LinkDb interface {
	GetLink(string, *http.Request) (string, error)

	AddLink(string, *http.Request) (string, error)

	Close() error
}

type word struct {
		Id    int 	`json:"id"`
		Text string	`json:"text"`
}

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
	c.Header("Access-Control-Allow-Origin", "*")
    c.JSON(http.StatusOK, js)
}

func init() {
	db = newDbMem() // newDbGoon()

	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.GET("/v1/words.json", words)

	http.Handle("/v1/", r)
}
