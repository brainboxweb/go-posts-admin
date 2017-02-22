package search_test

import (

	"testing"
	"github.com/brainboxweb/go-posts-admin/search"

)

func TestMusic(t *testing.T) {

	result := search.TopResult("google")
	expected := "KIViy7L_lo8"

	if result != expected {
		t.Errorf("Expected %s got  %s", expected, result)
	}
}
