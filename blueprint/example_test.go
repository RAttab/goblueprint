// Copyright (c) 2014 Datacratic. All rights reserved.
//
// In this example we'll develop a very simple service which will be customized
// and configured using a JSON blob.

package blueprint_test

import (
	"github.com/datacratic/goblueprint/blueprint"

	"fmt"
)

// Handler defines a generic interface for our service.x
type Handler interface {
	Handle()
}

// Printer is a specialization of our handler that prints to the console. Note
// that it has a configuration parameter (ie. Prefix) to be configured in our
// JSON blob.
type Printer struct{ Value string }

func (printer *Printer) Handle() { fmt.Print(printer.Value) }

// MultiHandler contains and executes multiple handlers. What we want is to be
// able to customize which handler is executed without having to recompile.
type MultiHandler struct{ Handlers []Handler }

func (multi *MultiHandler) Handle() {
	for _, handler := range multi.Handlers {
		handler.Handle()
	}
}

// Our first step is to register al the relevant types with the blueprint
// registry. This is necessary because we'll later want to specify types but
// golang doesn't currently provide a global registry of types.
func init() {
	blueprint.Register(Printer{})
	blueprint.Register(MultiHandler{})
}

// This is the meat of our example where we'll build our MultiHandler from
// multiple Handlers using a simple JSON blob.
//
// First, we create two printers. The special '!' notation is used to indicate
// the type of the object to be created. Note that in most cases the types can
// be infered from the object being created and can therefore be ignored.
//
// We then further customize our printers by specifying the content of the Value
// field. Since goblueprint relies on gopath to configure objects, this technic
// can be used to customize just about anything about a program so long as the
// fields are public.
//
// Finally, we create our MultiHandler and fill in its Handlers field by linking
// in the two handlers we previously created using the '#' notation. Note that
// the links can contain arbitrary paths and therefore can be used to avoid
// duplicating configuration parameters across multiple object.
var schema = `
{
    "hello!Printer": { "Value": "Hello " },
    "world!Printer": { "Value": "World!" },

    "multi!MultiHandler": {
        "#Handlers": [ "hello", "world" ]
    }
}
`

func ExampleBlueprint() {

	// Now all we need to do is load-up our JSON schema into a
	// map[string]interface{} object. Note that it's also possible to construct
	// objects directly using the LoadJSONInto function.
	values, err := blueprint.LoadJSON([]byte(schema))
	if err != nil {
		panic(err)
	}

	// We can then access our constructed multi object and execute it to get the
	// expected result. Magic!
	multi := values["multi"].(*MultiHandler)
	multi.Handle()

	// Output:
	// Hello World!
}
