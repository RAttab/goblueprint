// Copyright (c) 2014 Datacratic. All rights reserved.

package blueprint

import (
	"time"
)

// DurationConverter converts string values into time.Duration values.
func DurationConverter(value interface{}) (interface{}, error) {
	if str, ok := value.(string); ok {
		return time.ParseDuration(str)
	}
	return value, nil
}

func init() {
	RegisterConverter(time.Duration(0), ConverterFn(DurationConverter))
}
