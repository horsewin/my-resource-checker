package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"sbcntr2-test-tool/internal/validator"
)

type JSONReporter struct{}

func NewJSONReporter() *JSONReporter {
	return &JSONReporter{}
}

func (r *JSONReporter) ReportResult(result *validator.ValidationResult) error {
	output := r.formatResult(result)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func (r *JSONReporter) ReportSummary(summary *validator.ValidationSummary) error {
	output := r.formatSummary(summary)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func (r *JSONReporter) formatResult(result *validator.ValidationResult) map[string]interface{} {
	resources := make([]map[string]interface{}, 0, len(result.Resources))
	for _, res := range result.Resources {
		resources = append(resources, map[string]interface{}{
			"type":     res.Type,
			"id":       res.ID,
			"name":     res.Name,
			"status":   res.Status.String(),
			"expected": res.Expected,
			"actual":   res.Actual,
			"errors":   res.Errors,
			"warnings": res.Warnings,
		})
	}

	errors := make([]map[string]interface{}, 0, len(result.Errors))
	for _, err := range result.Errors {
		errors = append(errors, map[string]interface{}{
			"type":        err.Type,
			"resource":    err.Resource,
			"property":    err.Property,
			"expected":    err.Expected,
			"actual":      err.Actual,
			"message":     err.Message,
			"suggestion":  err.Suggestion,
			"documentRef": err.DocumentRef,
		})
	}

	warnings := make([]map[string]interface{}, 0, len(result.Warnings))
	for _, warn := range result.Warnings {
		warnings = append(warnings, map[string]interface{}{
			"resource": warn.Resource,
			"message":  warn.Message,
		})
	}

	return map[string]interface{}{
		"stepNumber": result.StepNumber,
		"stepName":   result.StepName,
		"status":     result.Status.String(),
		"duration":   result.Duration.String(),
		"resources":  resources,
		"errors":     errors,
		"warnings":   warnings,
	}
}

func (r *JSONReporter) formatSummary(summary *validator.ValidationSummary) map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(summary.Results))
	for _, res := range summary.Results {
		results = append(results, r.formatResult(&res))
	}

	return map[string]interface{}{
		"totalSteps":   summary.TotalSteps,
		"passedSteps":  summary.PassedSteps,
		"failedSteps":  summary.FailedSteps,
		"skippedSteps": summary.SkippedSteps,
		"results":      results,
	}
}
