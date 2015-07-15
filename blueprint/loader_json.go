// Copyright (c) 2014 Datacratic. All rights reserved.

package blueprint

import (
	"github.com/RAttab/gopath/path"

	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// LoadJSON using an extendded JSON representation to construct an object.
//
// There are two extension to the JSON format to enable object construction.
// First off, the type of an object can be specified by appending the '!'
// character followed by the type name to the key of an object. eg.
//
//     { "blah!Blah": { ... }, "bleh!string": "weeee", ... }
//
// Qualifying the type is generally required when dealing with interfaces where
// a type can't be deduced from the object being loaded.
//
// The second extension is used to link two objects toghether by prefixing the
// key of an object with the '#' character. eg.
//
//     {
//         "blah": { "bleh": "weee" },
//         "#foo": "blah.bleh",
//         "#bar": [ "foo", "bar.0" ]
//     }
//
// A link causes the object to be filled with the object at the specified
// path. Optionally, an array can also be filled in from multiple paths as
// demonstrated by the bar key.
//
// This JSON format has one major downside: it's not possible to qualify the
// type of array elements. This becomes an issue when dealing with an array of
// interface. The work-around is to construct the objects and link them in.
func LoadJSON(body []byte) (map[string]interface{}, error) {
	loader := &loaderJSON{Loader: &Loader{Values: make(map[string]interface{})}}

	values, err := loader.Load(body)
	if err != nil {
		return nil, err
	}

	return values.(map[string]interface{}), nil
}

// LoadJSONInto constructs the given value using the JSON representation
// described in LoadJSON.
func LoadJSONInto(body []byte, value interface{}) error {
	loader := &loaderJSON{Loader: &Loader{Values: value}}
	_, err := loader.Load(body)
	return err
}

type loaderJSON struct{ *Loader }

func (loader *loaderJSON) Load(body []byte) (interface{}, error) {
	var obj interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, err
	}

	loader.load(nil, obj)
	return loader.Finish()
}

func (loader *loaderJSON) load(current path.P, obj interface{}) {
	switch obj.(type) {

	case map[string]interface{}:
		loader.loadMap(current, obj.(map[string]interface{}))

	case []interface{}:
		loader.loadSlice(current, obj.([]interface{}))

	default:
		loader.Add(current, obj)
	}
}

func (loader *loaderJSON) loadMap(current path.P, obj map[string]interface{}) {
	for key, value := range obj {
		if strings.HasPrefix(key, "#") {
			loader.loadLinks(append(current, key[1:]), value)
			continue
		}

		if i := strings.Index(key, "!"); i > 0 {
			typ := key[i+1:]
			key = key[:i]
			loader.Type(append(current, key), typ)
		}

		loader.load(append(current, key), value)
	}
}

func (loader *loaderJSON) loadLinks(current path.P, obj interface{}) {
	switch obj.(type) {

	case string:
		loader.Link(current, path.New(obj.(string)))

	case []interface{}:
		for i, value := range obj.([]interface{}) {
			loader.loadLinks(append(current, strconv.Itoa(i)), value)
		}

	default:
		loader.ErrorAt(fmt.Errorf("unknown object type '%T' for links", obj), current)
	}
}

func (loader *loaderJSON) loadSlice(current path.P, obj []interface{}) {
	for i, item := range obj {
		loader.load(append(current, strconv.Itoa(i)), item)
	}
}
