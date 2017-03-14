package main

import (

	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"cloud.google.com/go/datastore"

	"golang.org/x/net/context"
	"testing"
)

func testDB(t *testing.T, db WordDb) {
	defer db.Close()

	if db.Text != nil {
		t.Fatal(err)
	}
}
