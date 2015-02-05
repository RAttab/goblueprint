// Copyright (c) 2014 Datacratic. All rights reserved.

package blueprint

import (
	"github.com/datacratic/goklog/klog"
	"github.com/datacratic/gopath/path"

	"fmt"
	"reflect"
)

// MaxLinksDepth is used to protect against cyclic links by limitting how far a
// link can be resolved.
const MaxLinksDepth = 128

// Loader gradually constructs an object using calls to Add, Type and Link and
// finalizes the result using Finish. Construction is achieved through the
// gopath library  which drives the core of Loader.
//
// Note that during construction no errors are reported. Instead they are
// accumulated into an Errors object and reported only when Finish is
// called.
type Loader struct {

	// Values is the value to be constructed.
	Values interface{}

	links  map[string]path.P
	errors Errors
}

// Add sets the object at the given path to value.
func (loader *Loader) Add(src path.P, value interface{}) {
	klog.KPrintf("blueprint.loader.add.debug", "src=%s, value={%T, %v}", src, value, value)

	err := src.Set(loader.Values, value)
	if err == nil {
		return
	}

	if err == path.ErrInvalidType {
		var typ reflect.Type
		if typ, err = src.Type(loader.Values); err == nil {
			if value, err = convert(typ, value); err == nil {
				err = src.Set(loader.Values, value)
			}
		}
	}

	loader.ErrorAt(err, src)
}

// Type asserts the type of an object at the given path.  This is useful when
// dealing with interfaces which can't be pathed through unless they're
// associated with a concrete type.
func (loader *Loader) Type(src path.P, name string) {
	klog.KPrintf("blueprint.loader.type.debug", "src=%s, name=%s", src, name)

	if value, ok := New(name); !ok {
		loader.ErrorAt(fmt.Errorf("unknown type '%s'", name), src)

	} else if err := src.Set(loader.Values, value); err != nil {
		loader.ErrorAt(err, src)
	}
}

// Link indicates that the object at the given src path should be equal to the
// value at the given target path. All links are resolved when calling Finish
// so there are no ordering constraints on links.
func (loader *Loader) Link(src, target path.P) {
	klog.KPrintf("blueprint.loader.link.debug", "src=%s, target=%s", src, target)

	if loader.links == nil {
		loader.links = make(map[string]path.P)
	}

	loader.links[src.String()] = target
}

// ErrorAt is used to report an error while loading the given path. Errors are
// accumulated during loading and only reported back to the user when Finish is
// called.
func (loader *Loader) ErrorAt(err error, src path.P) {
	if err != nil {
		loader.errors = append(loader.errors, fmt.Errorf("%s at '%s'", err, src))
	}
}

// Finish completes and returns the object. If errors were encountered during
// loading, they're all returned here as type Errors.
func (loader *Loader) Finish() (interface{}, error) {
	if loader.links != nil {
		for src, target := range loader.links {
			var value interface{}

			dst, err := loader.resolve(target, 0)

			klog.KPrintf("blueprint.loader.finish.debug", "src=%s, target=%s", src, dst)

			if err == nil {
				if value, err = dst.Get(loader.Values); err == nil && value == nil {
					err = fmt.Errorf("unable to link '%s' to nil value '%s'", src, dst)
				}
			}

			if err == nil {
				err = path.New(src).Set(loader.Values, value)
			}

			if err != nil {
				loader.ErrorAt(err, path.New(src))
			}
		}
	}

	// Required otherwise we set the type param on the error interface which
	// makes the error non-nil. One of those fun parts of the go language.
	if loader.errors != nil {
		return nil, loader.errors
	}
	return loader.Values, nil
}

func (loader *Loader) resolve(target path.P, depth int) (path.P, error) {
	if depth > MaxLinksDepth {
		return nil, fmt.Errorf("reached max links depth for '%s'", target)
	}

	for i := 1; i <= len(target); i++ {
		if redirect, ok := loader.links[target[:i].String()]; ok {
			return loader.resolve(append(redirect, target[i:]...), depth+1)
		}
	}

	return target, nil
}
