package validation

import (
	"fmt"
	"strings"
)

type IncomingInstructionValidationResult struct {
	success bool
	errors  []string
}

func (v IncomingInstructionValidationResult) IsValid() bool {
	return v.success
}

func (v IncomingInstructionValidationResult) Failed() bool {
	return !v.success
}

func (v IncomingInstructionValidationResult) GetErrors() []string {
	return v.errors
}

func (v IncomingInstructionValidationResult) Error() string {
	return fmt.Sprintf("payment request validation failed, errors: %s", strings.Join(v.errors, "; "))
}

var Valid = IncomingInstructionValidationResult{success: true}

func Invalid(reasons ...string) IncomingInstructionValidationResult {
	return IncomingInstructionValidationResult{success: false, errors: reasons}
}

func mergeValidationErrors(results ...IncomingInstructionValidationResult) IncomingInstructionValidationResult {
	errors := make([]string, 0)
	for _, result := range results {
		if !result.IsValid() {
			errors = append(errors, result.errors...)
		}
	}

	if len(errors) > 0 {
		return IncomingInstructionValidationResult{success: false, errors: errors}
	}

	return Valid
}
