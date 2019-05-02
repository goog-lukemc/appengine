package datastore

import (
	"reflect"
	"testing"
)

var testCasesKeyCompat = []struct {
	desc       string
	key        *Key
	encodedKey string
}{
	{
		desc: "A control test for legacy to legacy key conversion int as the key",
		key: &Key{
			kind:  "Person",
			intID: 1,
			appID: "glibrary",
		},
		encodedKey: "aghnbGlicmFyeXIMCxIGUGVyc29uGAEM",
	},
	{
		desc: "A control test for legacy to legacy key conversion string as the key",
		key: &Key{
			kind:     "Graph",
			stringID: "graph:7-day-active",
			appID:    "glibrary",
		},
		encodedKey: "aghnbGlicmFyeXIdCxIFR3JhcGgiEmdyYXBoOjctZGF5LWFjdGl2ZQw",
	},

	// These are keys encoded with cloud.google.com/go/datastore
	// Standard int as the key
	{
		desc: "Convert new key format to old key with int id",
		key: &Key{
			kind:  "WordIndex",
			intID: 1033,
			appID: "glibrary",
		},
		encodedKey: "Eg4KCVdvcmRJbmRleBCJCA",
	},
	// These are keys encoded with cloud.google.com/go/datastore
	// Standard string
	{
		desc: "Convert new key format to old key with string id",
		key: &Key{
			kind:     "WordIndex",
			stringID: "IAmAnID",
			appID:    "glibrary",
		},
		encodedKey: "EhQKCVdvcmRJbmRleBoHSUFtQW5JRA",
	},

	// These are keys encoded with cloud.google.com/go/datastore
	// ID String with parent as string
	{
		desc: "Convert new key format to old key with string id with a parent",
		key: &Key{
			kind:     "WordIndex",
			stringID: "IAmAnID",
			appID:    "glibrary",
			parent: &Key{
				kind:     "LetterIndex",
				stringID: "IAmAnotherID",
				appID:    "glibrary",
			},
		},
		encodedKey: "EhsKC0xldHRlckluZGV4GgxJQW1Bbm90aGVySUQSFAoJV29yZEluZGV4GgdJQW1BbklE",
	},
}

func TestKeyCoversion(t *testing.T) {
	// Simulate the key converter enablement
	keyConversionProject = "glibrary"

	for _, tc := range testCasesKeyCompat {
		dk, err := DecodeKey(tc.encodedKey)
		if err != nil {
			t.Fatalf("%v", err.Error())
		}
		if !reflect.DeepEqual(dk, tc.key) {
			t.Errorf("%s: got %+v, want %+v", tc.desc, dk, tc.key)
		}

	}
}
