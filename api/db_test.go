package main

import (
	"testing"
)

func testDB(t *testing.T, db WordDb) {
	defer db.Close()

	if db.Text != nil {
		t.Fatal(err)
	}
}
