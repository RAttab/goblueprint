// Copyright (c) 2014 Datacratic. All rights reserved.

package blueprint

import (
	"github.com/datacratic/goklog/klog"

	"reflect"
	"sync"
)

// Converter are used to convert one value representation into another before
// attempting to load it into an object.
type Converter interface {
	Convert(interface{}) (interface{}, error)
}

// ConverterFn is a convenience wrappers for functions to implement the
// Converter interface.
type ConverterFn func(interface{}) (interface{}, error)

// Convert uses the underlying function to convert the value given value into
// another value.
func (fn ConverterFn) Convert(value interface{}) (interface{}, error) {
	return fn(value)
}

var converters map[reflect.Type]Converter
var convertersMutex sync.Mutex

// RegisterConverter makes the given converter for the type of the given object.
func RegisterConverter(obj interface{}, conv Converter) {
	convertersMutex.Lock()
	defer convertersMutex.Unlock()

	typ := reflect.TypeOf(obj)

	if converters == nil {
		converters = make(map[reflect.Type]Converter)
	}

	if _, ok := converters[typ]; ok {
		klog.KFatalf("blueprint.converters.register.error", "duplicate converters for type '%s'", typ)
	}

	converters[typ] = conv
}

func convert(typ reflect.Type, value interface{}) (interface{}, error) {
	convertersMutex.Lock()

	var err error

	if reflect.TypeOf(value) != typ {
		if conv, ok := converters[typ]; ok {
			value, err = conv.Convert(value)
		}
	}

	convertersMutex.Unlock()
	return value, err
}
