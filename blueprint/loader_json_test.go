// Copyright (c) 2014 Datacratic. All rights reserved.

package blueprint

import (
	"testing"
)

type Blah struct{ A []string }

func init() { Register(Blah{}) }

func TestLoader_JSON(t *testing.T) {
	json := `{
        "string": "blah",
        "int": 10,
        "float32": 10.10,

        "#link-a": "link-b",
        "#link-b": "string",
        "#link-c": "link-b",

        "blah!Blah": { "A": [ "a", "b", "c" ] },
        "bleh!Blah": { "#A": [ "blah.A.2", "blah.A.0", "blah.A.1" ] },

        "#W": "X.Base",

        "X!github.com/RAttab/goblueprint/blueprint/Struct": {
            "I": 20,
            "#S": "string",
            "Base!Impl": {
                "I": 30,
                "#S": "X.S"
            }
        },

        "#Y": "X",

        "Z!Impl": {
            "#I": "Y.I",
            "#S": "Y.Base.S"
        }
    }`

	x := &Struct{
		I: 20,
		S: "blah",
		Base: &Impl{
			I: 30,
			S: "blah",
		},
	}

	CheckLoadJSON(t, json, map[string]interface{}{
		"string":  "blah",
		"int":     float64(10),
		"float32": float64(10.10),

		"link-a": "blah",
		"link-b": "blah",
		"link-c": "blah",

		"blah": &Blah{A: []string{"a", "b", "c"}},
		"bleh": &Blah{A: []string{"c", "a", "b"}},

		"W": x.Base,
		"X": x,
		"Y": x,
		"Z": &Impl{
			I: x.I,
			S: "blah",
		},
	})
}

func CheckLoadJSON(t *testing.T, json string, exp map[string]interface{}) {
	values, err := LoadJSON([]byte(json))

	if err != nil {
		t.Errorf("FAIL: unable to load json\n%v", err)
		return
	}

	CheckValues(t, values, exp)
}
