// Copyright (c) 2014 Datacratic. All rights reserved.

package store

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
)

var (
	registry map[string]reflect.Type
)

// Register adds the type of the specified value to the global registry so that instances can be created.
func Register(value interface{}) {
	if registry == nil {
		registry = make(map[string]reflect.Type)
	}

	t := reflect.TypeOf(value)
	if t == nil {
		log.Fatalf("cannot register '%v'", value)
	}

	s := t.Name()
	if _, ok := registry[s]; ok {
		log.Fatalf("type '%s' is already registered", s)
	}

	registry[s] = t
}

type context struct {
	names map[string]interface{}
}

// Create uses the JSON data to create and populate instances of reflected or registered types.
func Create(data json.RawMessage) (result interface{}, err error) {
	parser := context{
		names: make(map[string]interface{}),
	}

	value, err := parser.readInstance(data)
	if err != nil {
		return
	}

	result = value.Interface()
	return
}

func (parser *context) readInstance(data json.RawMessage) (result reflect.Value, err error) {
	items := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &items); err != nil {
		return
	}

	text, ok := items["Type"]
	if !ok {
		err = fmt.Errorf("missing 'type' field")
		return
	}

	kind := ""
	if err = json.Unmarshal(text, &kind); err != nil {
		return
	}

	t, ok := registry[kind]
	if !ok {
		err = fmt.Errorf("type '%s' has not been registered", kind)
		return
	}

	p := reflect.New(t)
	if err = parser.readFields(p.Elem(), items); err != nil {
		return
	}

	result = p
	return
}

func (parser *context) readFields(value reflect.Value, items map[string]json.RawMessage) (err error) {
	t := value.Type()

	for i := 0; i != t.NumField(); i++ {
		f := t.Field(i)
		k := f.Type.Kind()

		switch k {
		case reflect.Struct:
			if f.Anonymous {
				if err = parser.readFields(value.Field(i), items); err != nil {
					err = fmt.Errorf("%s in anonymous field '%s' of struct '%s'", err, f.Name, t.Name())
					return
				}
			} else {
				if err = parser.readStructField(value.Field(i), f, items); err != nil {
					return
				}
			}

		case reflect.Slice:
			if err = parser.readSliceField(value.Field(i), f, items); err != nil {
				err = fmt.Errorf("%s in slice field '%s' of struct '%s'", err, f.Name, t.Name())
				return
			}

		case reflect.String:
			if err = parser.readStringField(value.Field(i), f, items); err != nil {
				err = fmt.Errorf("%s in field '%s' of struct '%s'", err, f.Name, t.Name())
				return
			}

		default:
			err = fmt.Errorf("unhandled kind of field '%s'", k)
		}
	}

	return
}

func (parser *context) readSliceField(value reflect.Value, field reflect.StructField, data map[string]json.RawMessage) (err error) {
	if item, ok := data[field.Name]; ok {
		t := field.Type.Elem()
		k := t.Kind()

		switch k {
		case reflect.Interface:
			var items []json.RawMessage
			if err = json.Unmarshal(item, &items); err != nil {
				return
			}

			n := len(items)
			slice := reflect.MakeSlice(field.Type, n, n)
			for i := 0; i != len(items); i++ {
				var item reflect.Value
				item, err = parser.readInstance(items[i])
				if err != nil {
					return
				}

				slice.Index(i).Set(item)
			}

			value.Set(slice)

		default:
			err = fmt.Errorf("unhandled kind of slice '%s'", k)
		}
	}

	return
}

func (parser *context) readStructField(value reflect.Value, field reflect.StructField, data map[string]json.RawMessage) (err error) {
	if item, ok := data[field.Name]; ok {
		items := make(map[string]json.RawMessage)
		if err = json.Unmarshal(item, &items); err != nil {
			return
		}

		err = parser.readFields(value, items)
		if err != nil {
			err = fmt.Errorf("%s in field '%s' of struct '%s'", err, field.Name, field.Type.Name())
			return
		}
	}

	return
}

func (parser *context) readStringField(value reflect.Value, field reflect.StructField, data map[string]json.RawMessage) (err error) {
	if item, ok := data[field.Name]; ok {
		s := ""
		if err = json.Unmarshal(item, &s); err != nil {
			return
		}

		value.Set(reflect.ValueOf(s))
	}

	return
}
