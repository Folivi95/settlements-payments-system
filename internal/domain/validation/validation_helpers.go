package validation

import (
	"reflect"
)

func isEmpty(item interface{}) bool {
	return reflect.ValueOf(item).IsZero()
}

func validateIsNotEmpty(item interface{}, typeName string) IncomingInstructionValidationResult {
	if isEmpty(item) {
		return Invalid(typeName + " should not be empty.")
	}
	return Valid
}

func validateIsEmpty(item interface{}, typeName string) IncomingInstructionValidationResult {
	if !isEmpty(item) {
		return Invalid(typeName + " should be empty.")
	}
	return Valid
}
