package main

import (
	"google.golang.org/appengine/aetest"
	"testing"
)

func testDBNormal(t *testing.T) {
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

	if _, err := db.Add(dummyId, js, ctx); err != nil {
		t.Errorf("%v", err)
	}

	all, err := db.GetAll(dummyId, false, "", ctx)
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(all) == 0 {
		t.Errorf("failed to write in DB")
	}

	if all[0].Text != "test" {
		t.Errorf("Text is not equal")
	}
}

func testDBEmptyText(t *testing.T) {
	dummyId := "testprofile"
	var db ContentDb

	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatalf("aetest: %v", err)
	}
	defer done()

	js := PostContent{
		Text: "",
	}

	if _, err := db.Add(dummyId, js, ctx); err == nil {
		t.Errorf("db.Add must be error")
	}
}
