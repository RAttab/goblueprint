// Copyright (c) 2014 Datacratic. All rights reserved.

package blueprint

import (
	"reflect"
	"testing"
)

type Basics struct {
	Text string
}

type Extended struct {
	Basics
	MoreText string
	Items    []Handler
}

type Handler interface {
	Test()
}

type HandlerA struct {
	Data string
}

func (h *HandlerA) Test() {
}

type HandlerB struct {
	Extended
	Data string
}

func (h *HandlerB) Test() {
}

func init() {
	Register(Basics{})
	Register(Extended{})
	Register(HandlerA{})
	Register(HandlerB{})
}

var scriptExamples = map[string][]byte{
	"Basics": []byte(`
{
  "Type": "Basics",
  "Text": "hello world"
}`),
	"Extended": []byte(`
{
  "Type": "Extended",
  "Text": "hello",
  "MoreText": "world",
  "Items": [
    {
      "Type": "HandlerA",
      "Data": "a"
    },
    {
      "Type": "HandlerB",
      "Data": "b",
	  "Text": "hello world"
    }
  ]
}`),
}

func TestCreate(t *testing.T) {
	basics, err := Create(scriptExamples["Basics"])
	if err != nil {
		t.Fatal(err)
	}

	if item, ok := basics.(*Basics); !ok {
		t.Fatalf("unexpected created type i.e. '%s' != Basics", reflect.TypeOf(basics))
	} else {
		if item.Text != "hello world" {
			t.Fatalf("unexpected value i.e. '%s' != 'hello world'", item.Text)
		}
	}

	extended, err := Create(scriptExamples["Extended"])
	if err != nil {
		t.Fatal(err)
	}

	if item, ok := extended.(*Extended); !ok {
		t.Fatalf("unexpected type i.e. '%s' != Extended", reflect.TypeOf(extended))
	} else {
		if item.Text != "hello" {
			t.Fatalf("unexpected value i.e. '%s' != 'hello'", item.Text)
		}

		if item.MoreText != "world" {
			t.Fatalf("unexpected value i.e. '%s' != 'world'", item.MoreText)
		}

		if len(item.Items) != 2 {
			t.Fatalf("unexpected number of items i.e. '%d' != 2", len(item.Items))
		}

		if a, ok := item.Items[0].(*HandlerA); !ok {
			t.Fatalf("unexpected type i.e. '%s' != HandlerA", reflect.TypeOf(item.Items[0]))
		} else {
			if a.Data != "a" {
				t.Fatalf("unexpected value i.e. '%s' != 'a'", a.Data)
			}
		}

		if b, ok := item.Items[1].(*HandlerB); !ok {
			t.Fatalf("unexpected type i.e. '%s' != HandlerB", reflect.TypeOf(item.Items[1]))
		} else {
			if b.Data != "b" {
				t.Fatalf("unexpected value i.e. '%s' != 'b'", b.Data)
			}

			if b.Text != "hello world" {
				t.Fatalf("unexpected value i.e. '%s' != 'hello world'", b.Text)
			}
		}
	}
}
