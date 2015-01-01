# goblueprint #

Object construction library via JSON configuration files.

The goal is to allow the configuration and composition of an executable without
recompiling or modifying the source code. This is done by creating a JSON blob
that specifies the make up of objects which is then used by goblueprint to
materialize the object.


## Installation ##

You can download the code via the usual go utilities:

```
go get github.com/datacratic/goblueprint/blueprint
```

To build the code and run the test suite along with several static analysis
tools, use the provided Makefile:

```
make test
```

Note that the usual go utilities will work just fine but we require that all
commits pass the full suite of tests and static analysis tools.


## Documentation ##

Documentation is available [here](https://godoc.org/github.com/datacratic/goblueprint/blueprint).
Usage examples are available in this [test suite](blueprint/example_test.go).


## License ##

The source code is available under the Apache License. See the LICENSE file for
more details.
