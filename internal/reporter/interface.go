package reporter

import "sbcntr2-test-tool/internal/validator"

type Reporter interface {
	ReportResult(result *validator.ValidationResult) error
	ReportSummary(summary *validator.ValidationSummary) error
}
