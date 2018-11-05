package main

import (
	"testing"
)

func TestTagExists(t *testing.T) {

	var tags []Tag

		tags = append(tags, Tag{
			Key:   "Name",
			Value: "Value",
		})
	var volume = new(Volume)

	volume.Tags = tags

	if v := volume.tagExists(); v != true {
		t.Error("Expected 'tagExist()'")
	}

}
