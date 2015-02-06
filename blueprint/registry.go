// Copyright (c) 2014 Datacratic. All rights reserved.

package blueprint

import (
	"github.com/datacratic/goklog/klog"

	"bytes"
	"reflect"
	"sort"
	"sync"
)

// Registry indexes the various types which will be used by the loader to create
// objects as needed.
//
// Types can either be accessed by their short form (eg. MyType) or their fully
// qualified form (eg. github.com/me/golib/MyType). Conflicts with the short
// form are resolved arbitrarily. The fully qualified is therefore recommended
// when conflicts are a possibility.
type Registry struct {
	mutex sync.Mutex
	types map[string]reflect.Type
}

// Register associates the given value's type with the short and fully qualified
// name.
func (reg *Registry) Register(value interface{}) {
	if value == nil {
		klog.KPanicf("blueprint.registry.error", "attempted to register nil")
	}

	typ := reflect.TypeOf(value)
	for typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Interface {
		typ = typ.Elem()
	}

	reg.mutex.Lock()

	if reg.types == nil {
		reg.types = make(map[string]reflect.Type)
	}

	name := typ.Name()
	if pkg := typ.PkgPath(); pkg != "" {
		name = pkg + "/" + name
	}

	if _, ok := reg.types[name]; ok {
		klog.KFatalf("blueprint.registry.error", "duplicate registration attempt for '%s'", name)
	}

	reg.types[name] = typ

	// Add a shorthand alias to make life simpler.
	if _, ok := reg.types[typ.Name()]; !ok {
		reg.types[typ.Name()] = typ
	}

	reg.mutex.Unlock()
}

// Get returns the type associated with the given name or false as the second
// parameter if no types exists for that name.
func (reg *Registry) Get(name string) (reflect.Type, bool) {
	reg.mutex.Lock()

	typ, ok := reg.types[name]

	reg.mutex.Unlock()

	return typ, ok
}

// New returns a pointer to a newly instantiated object associated with the
// given name or false as the second parameter if no types exists for that name.
func (reg *Registry) New(name string) (interface{}, bool) {
	if typ, ok := reg.Get(name); ok {
		return reflect.New(typ).Interface(), true
	}
	return nil, false
}

// String returns the string representation of the registry suitable for
// debugging.
func (reg *Registry) String() string {
	reg.mutex.Lock()

	var keys []string

	for key := range reg.types {
		keys = append(keys, key)
	}

	reg.mutex.Unlock()

	sort.Strings(keys)

	buffer := new(bytes.Buffer)
	buffer.WriteString("[")

	for _, key := range keys {
		buffer.WriteString("\n    ")
		buffer.WriteString(key)
	}

	buffer.WriteString("\n]")
	return buffer.String()
}

// DefaultRegistry is a global Registry object. Note that the global registry is
// automatically filled in with the go primitive types (eg. ints, floats,
// strings, etc.)
var DefaultRegistry Registry

// Register associates the given value's type with the short and fully qualified
// name.
func Register(value interface{}) { DefaultRegistry.Register(value) }

// Get returns the type associated with the given name or false as the second
// parameter if no types exists for that name.
func Get(name string) (reflect.Type, bool) { return DefaultRegistry.Get(name) }

// New returns a pointer to a newly instantiated object associated with the
// given name or false as the second parameter if no types exists for that name.
func New(name string) (interface{}, bool) { return DefaultRegistry.New(name) }

func init() {
	Register(int(0))
	Register(int8(0))
	Register(int16(0))
	Register(int32(0))
	Register(int64(0))

	Register(uint(0))
	Register(uint8(0))
	Register(uint16(0))
	Register(uint32(0))
	Register(uint64(0))

	Register(float32(0))
	Register(float64(0))

	Register("")
}
