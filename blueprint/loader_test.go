// Copyright (c) 2014 Datacratic. All rights reserved.

package blueprint

import (
	"github.com/RAttab/gopath/path"

	"fmt"
	"github.com/RAttab/goset"
	"reflect"
	"testing"
)

type Base interface {
	Eq(Base) bool
	String() string
}

type Impl struct {
	I int
	S string
}

func (impl *Impl) Eq(o Base) bool {
	if other, ok := o.(*Impl); ok {
		return impl.I == other.I && impl.S == other.S
	}
	return false
}

func (impl *Impl) String() string {
	return fmt.Sprintf("Impl{I: %d, S:'%s'}", impl.I, impl.S)
}

type Struct struct {
	I    int
	S    string
	Base Base
}

func (s *Struct) Eq(o Base) bool {
	if other, ok := o.(*Struct); ok {
		return s.I == other.I && s.S == other.S && s.Base.Eq(other.Base)
	}
	return false
}

func (s *Struct) String() string {
	return fmt.Sprintf("Struct{I: %d, S:'%s', Base:%s}", s.I, s.S, s.Base)
}

func init() {
	Register(Impl{})
	Register(Struct{})
}

func TestLoader(t *testing.T) {
	loader := NewLoader()

	loader.TestAdd(t, "string", "blah")
	loader.TestAdd(t, "int", int(10))
	loader.TestAdd(t, "float32", float32(10.10))

	loader.TestLink(t, "link-a", "link-b")
	loader.TestLink(t, "link-b", "string")
	loader.TestLink(t, "link-c", "link-b")

	loader.TestLink(t, "W", "X.Base")

	loader.TestType(t, "X", "github.com/RAttab/goblueprint/blueprint/Struct")
	loader.TestAdd(t, "X.I", int(20))
	loader.TestLink(t, "X.S", "string")
	loader.TestType(t, "X.Base", "Impl")
	loader.TestAdd(t, "X.Base.I", int(30))
	loader.TestLink(t, "X.Base.S", "X.S")

	loader.TestLink(t, "Y", "X")

	loader.TestType(t, "Z", "Impl")
	loader.TestLink(t, "Z.I", "Y.I")
	loader.TestLink(t, "Z.S", "Y.Base.S")

	x := &Struct{
		I: 20,
		S: "blah",
		Base: &Impl{
			I: 30,
			S: "blah",
		},
	}

	loader.TestFinish(t, map[string]interface{}{
		"string":  "blah",
		"int":     int(10),
		"float32": float32(10.10),

		"link-a": "blah",
		"link-b": "blah",
		"link-c": "blah",

		"W": x.Base,
		"X": x,
		"Y": x,
		"Z": &Impl{
			I: x.I,
			S: "blah",
		},
	})
}

func NewLoader() *Loader { return &Loader{Values: make(map[string]interface{})} }

func (loader *Loader) TestAdd(t *testing.T, src string, value interface{}) {
	n := len(loader.errors)

	loader.Add(path.New(src), value)

	if i := len(loader.errors); n != i {
		t.Errorf("FAIL: error during add(%s, %v) -> %s", src, value, Errors(loader.errors[n:]))
	}
}

func (loader *Loader) TestType(t *testing.T, src, name string) {
	n := len(loader.errors)

	loader.Type(path.New(src), name)

	if i := len(loader.errors); n != i {
		t.Errorf("FAIL: error during type(%s, %s) -> %s", src, name, Errors(loader.errors[n:]))
	}
}

func (loader *Loader) TestLink(t *testing.T, src, target string) {
	n := len(loader.errors)

	loader.Link(path.New(src), path.New(target))

	if i := len(loader.errors); n != i {
		t.Errorf("FAIL: error during type(%s, %s)  -> %s", src, target, Errors(loader.errors[n:]))
	}
}

func (loader *Loader) TestFinish(t *testing.T, exp map[string]interface{}) {
	n := len(loader.errors)

	value, _ := loader.Finish()

	if i := len(loader.errors); n != i {
		t.Errorf("FAIL: error during finish() -> %s", Errors(loader.errors[n:]))
	}

	CheckValues(t, value.(map[string]interface{}), exp)
}

func CheckValues(t *testing.T, values, exp map[string]interface{}) {
	a := set.NewString()
	b := set.NewString()

	for key := range values {
		a.Put(key)
	}
	for key := range exp {
		b.Put(key)
	}

	if diff := a.Difference(b); len(diff) > 0 {
		t.Errorf("FAIL: extra keys %s", diff)
	}
	if diff := b.Difference(a); len(diff) > 0 {
		t.Errorf("FAIL: missing keys %s", diff)
	}

	for key := range a.Intersect(b) {
		if expBase, ok := exp[key].(Base); ok {
			if valueBase := values[key].(Base); !expBase.Eq(valueBase) {
				t.Errorf("FAIL(%s): value='%s' != exp='%s'", key, valueBase, expBase)
			}

		} else if !reflect.DeepEqual(values[key], exp[key]) {
			t.Errorf("FAIL(%s): {%T, %v} != {%T, %v}", key, values[key], values[key], exp[key], exp[key])
		}
	}
}
