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

	EditWord(string, EditWord, *http.Request) (Word, error)

	Delete(id string, r *http.Request) error

	Close() error
}

type (
	Word struct {
		Id    string 	`json:"id"`
		Text string	`json:"text"`
		Memo string `json:"memo"`
		Tag	 string `json:"tag"`
		IsReview bool `json:"is_review"`
		IsInput bool `json:"is_input"`
		Count int	 `json:"count"`
		Priority int `json:"priority"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	PostWord struct {
		Text     string `form:"text" json:"text" binding:"required"`
	}

	EditWord struct {
		Kind     string `form:"kind" json:"kind"`
		Memo     string `form:"memo" json:"memo"`
		Tag     string `form:"tag" json:"tag"`
		IsReview bool `json:"is_review"`
		IsInput bool `json:"is_input"`
		Count     int `form:"count" json:"count"`
		Priority  int `form:"priority" json:"priority"`
	}
)

var (
	db WordDb
)

func words(c *gin.Context) {

	profile := profileFromSession(c.Request)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "parse error"})
		return
	}
	
	c.Header("Access-Control-Allow-Origin", "*")
	if all, err := db.GetAll(c.Request); err == nil {
		if review := c.Query("is_review"); review!="" {
			ws := []Word{}
			for _, w := range all {
				if review == "true" && w.IsReview {
					ws = append(ws, w)
				} else if review == "false" && !w.IsReview {
						ws = append(ws, w)
					}
			}
			c.JSON(http.StatusOK, ws)
		} else {
			c.JSON(http.StatusOK, all)
		}
	} else {
		c.JSON(http.StatusInternalServerError, "error")
	}
}

func create(c *gin.Context) {
  var json PostWord
	c.Header("Access-Control-Allow-Origin", "*")

	profile := profileFromSession(c.Request)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "parse error"})
		return
	}

	if c.BindJSON(&json) == nil {
		log.Debugf(appengine.NewContext(c.Request), "post:%v", json)
		db.AddWord(json, c.Request)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "parse error"}) // 400
	}
}

func edit(c *gin.Context) {
  var json EditWord

	id := c.Param("id")

	profile := profileFromSession(c.Request)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "parse error"})
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	if c.BindJSON(&json) == nil {
		w, err := db.EditWord(id, json, c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error"})
		} else {
			c.JSON(http.StatusOK, w)
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "parse error"}) // 400
	}
}

func delete(c *gin.Context) {
	id := c.Param("id")

	profile := profileFromSession(c.Request)
	if profile == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "parse error"})
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	err := db.Delete(id, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
	} else {
		log.Debugf(appengine.NewContext(c.Request), "delete:%v", id)
		c.JSON(http.StatusOK, gin.H{"status":"ok"})
	}
}


func init() {
	db = newDbGoon() // newDbMem()

	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.GET("/v1/words.json", words)
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


	http.HandleFunc("/v1/login", loginHandler)
	http.HandleFunc("/v1/logout", logoutHandler)
	http.HandleFunc("/v1/oauth2callback", oauthCallbackHandler)

	http.Handle("/v1/", r)
}
