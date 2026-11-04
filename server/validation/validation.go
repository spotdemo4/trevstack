// Package validation checks request fields. Each check returns an
// InvalidArgument status that names the field, or nil.
package validation

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/timestamppb"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
)

// First returns the first error that is not nil.
func First(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// Required checks that an optional field is set.
func Required(field string, set bool) error {
	if !set {
		return invalid("%s is required", field)
	}
	return nil
}

// Length checks that value has between minimum and maximum characters.
func Length(field, value string, minimum, maximum int) error {
	length := utf8.RuneCountInString(value)
	if length < minimum {
		return invalid("%s must be at least %d characters", field, minimum)
	}
	if length > maximum {
		return invalid("%s must be at most %d characters", field, maximum)
	}
	return nil
}

// MaxBytes checks that value has at most maximum bytes.
func MaxBytes(field, value string, maximum int) error {
	if len(value) > maximum {
		return invalid("%s must be at most %d bytes", field, maximum)
	}
	return nil
}

// Pattern checks that value matches pattern, which description describes.
func Pattern(field, value string, pattern *regexp.Regexp, description string) error {
	if !pattern.MatchString(value) {
		return invalid("%s must %s", field, description)
	}
	return nil
}

// Between checks that value is between minimum and maximum, inclusive.
func Between(field string, value, minimum, maximum uint32) error {
	if value < minimum || value > maximum {
		return invalid("%s must be between %d and %d", field, minimum, maximum)
	}
	return nil
}

// Less checks that the lower value is less than the upper one when both are
// set.
func Less(lowerField string, lower *uint32, upperField string, upper *uint32) error {
	if lower != nil && upper != nil && *lower >= *upper {
		return invalid("%s must be less than %s", lowerField, upperField)
	}
	return nil
}

// TimeRange checks that the timestamps are valid and end is after start when
// both are set.
func TimeRange(start, end *timestamppb.Timestamp) error {
	if start != nil {
		if err := start.CheckValid(); err != nil {
			return invalid("start is not a valid timestamp")
		}
	}
	if end != nil {
		if err := end.CheckValid(); err != nil {
			return invalid("end is not a valid timestamp")
		}
	}
	if start != nil && end != nil && !end.AsTime().After(start.AsTime()) {
		return invalid("end must be after start")
	}
	return nil
}

func invalid(format string, args ...any) error {
	return trevrpc.InvalidArgument(fmt.Sprintf(format, args...))
}
