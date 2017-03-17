package main

import (
	"testing"
	"google.golang.org/appengine/aetest"
)

func testDB(t *testing.T) {
	dummyId := "testprofile"
	var db ContentDb

	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatalf("aetest: %v", err)
	}
	defer done()

	js := PostContent{
		Text: "test",
	}

	if _, err := db.Add(dummyId, js, ctx); err!=nil {
		t.Errorf("%v", err)
	}


	all, err := db.GetAll(dummyId, false, "", ctx)
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(all) == 0 {
		t.Errorf("%v", err)
	}
}
