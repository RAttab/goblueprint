// Copyright (c) 2014 Datacratic. All rights reserved.

package blueprint

import (
	"bytes"
)

// Errors aggregates multiple errors encountered during loading and reports them
// as a single error.
type Errors []error

// Error aggregates all the errors into a single string seperated by the '\n'
// character.
func (errors Errors) Error() string {
	buffer := new(bytes.Buffer)

	for _, err := range errors {
		buffer.WriteString(err.Error())
		buffer.WriteString("\n")
	}

	return buffer.String()
}
