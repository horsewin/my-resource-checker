package reporter

import (
	"fmt"
	"os"
	"sbcntr2-test-tool/internal/validator"
	"strings"
)

type ConsoleReporter struct {
	verbose bool
}

func NewConsoleReporter(verbose bool) *ConsoleReporter {
	return &ConsoleReporter{
		verbose: verbose,
	}
}

func (r *ConsoleReporter) ReportResult(result *validator.ValidationResult) error {
	r.printHeader(result)
	r.printResources(result.Resources)
	r.printErrors(result.Errors)
	r.printWarnings(result.Warnings)
	r.printFooter(result)

	return nil
}

func (r *ConsoleReporter) ReportSummary(summary *validator.ValidationSummary) error {
	fmt.Println("\n╔══════════════════════════════════════════════════════╗")
	fmt.Println("║            VALIDATION SUMMARY REPORT                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝\n")

	fmt.Printf("Total Steps: %d\n", summary.TotalSteps)
	fmt.Printf("✅ Passed: %d\n", summary.PassedSteps)
	fmt.Printf("❌ Failed: %d\n", summary.FailedSteps)
	fmt.Printf("⏭️  Skipped: %d\n\n", summary.SkippedSteps)

	for _, result := range summary.Results {
		r.printSummaryStep(&result)
	}

	r.printOverallStatus(summary)

	return nil
}

func (r *ConsoleReporter) printHeader(result *validator.ValidationResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("STEP %d: %s\n", result.StepNumber, result.StepName)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Status: %s\n", r.getStatusIcon(result.Status))
	fmt.Printf("Duration: %v\n\n", result.Duration)
}

func (r *ConsoleReporter) printResources(resources []validator.ResourceResult) {
	if len(resources) == 0 {
		return
	}

	fmt.Println("Resources Checked:")
	fmt.Println(strings.Repeat("-", 40))

	for _, resource := range resources {
		icon := r.getResourceStatusIcon(resource.Status)
		fmt.Printf("%s %s (%s)\n", icon, resource.Name, resource.Type)

		// Verbose output disabled to keep output clean

		for _, err := range resource.Errors {
			fmt.Printf("  ❌ %s\n", err)
		}

		for _, warn := range resource.Warnings {
			fmt.Printf("  ⚠️  %s\n", warn)
		}
	}
	fmt.Println()
}

func (r *ConsoleReporter) printErrors(errors []validator.ValidationError) {
	if len(errors) == 0 {
		return
	}

	fmt.Println("❌ Errors:")
	fmt.Println(strings.Repeat("-", 40))

	for _, err := range errors {
		fmt.Printf("• %s\n", err.Message)
		if err.Suggestion != "" {
			fmt.Printf("  💡 Suggestion: %s\n", err.Suggestion)
		}
		if err.DocumentRef != "" {
			fmt.Printf("  📖 Reference: %s\n", err.DocumentRef)
		}
		fmt.Println()
	}
}

func (r *ConsoleReporter) printWarnings(warnings []validator.ValidationWarning) {
	if len(warnings) == 0 {
		return
	}

	fmt.Println("⚠️  Warnings:")
	fmt.Println(strings.Repeat("-", 40))

	for _, warn := range warnings {
		fmt.Printf("• %s: %s\n", warn.Resource, warn.Message)
	}
	fmt.Println()
}

func (r *ConsoleReporter) printFooter(result *validator.ValidationResult) {
	fmt.Println(strings.Repeat("=", 60))

	switch result.Status {
	case validator.StatusPassed:
		fmt.Println("✅ All checks passed! You can proceed to the next step.")
	case validator.StatusWarning:
		fmt.Println("⚠️  Validation completed with warnings. Please review the warnings above.")
	case validator.StatusFailed:
		fmt.Println("❌ Validation failed. Please fix the errors above before proceeding.")
	case validator.StatusSkipped:
		fmt.Println("⏭️  Validation was skipped.")
	}
	fmt.Println()
}

func (r *ConsoleReporter) printSummaryStep(result *validator.ValidationResult) {
	icon := r.getStatusIcon(result.Status)
	fmt.Printf("%s Step %d: %s\n", icon, result.StepNumber, result.StepName)

	if result.Status == validator.StatusFailed && len(result.Errors) > 0 {
		for _, err := range result.Errors {
			fmt.Printf("   - %s\n", err.Message)
		}
	}
}

func (r *ConsoleReporter) printOverallStatus(summary *validator.ValidationSummary) {
	fmt.Println("\n" + strings.Repeat("=", 60))

	if summary.FailedSteps == 0 && summary.SkippedSteps == 0 {
		fmt.Println("🎉 Congratulations! All steps validated successfully!")
		fmt.Println("Your hands-on environment is correctly configured.")
	} else if summary.FailedSteps > 0 {
		fmt.Printf("❌ %d step(s) failed validation.\n", summary.FailedSteps)
		fmt.Println("Please review and fix the errors before proceeding.")
	} else {
		fmt.Println("⚠️  Some steps were skipped.")
		fmt.Println("Run individual step validations for more details.")
	}

	fmt.Println(strings.Repeat("=", 60))
}

func (r *ConsoleReporter) getStatusIcon(status validator.ValidationStatus) string {
	switch status {
	case validator.StatusPassed:
		return "✅ PASSED"
	case validator.StatusFailed:
		return "❌ FAILED"
	case validator.StatusWarning:
		return "⚠️  WARNING"
	case validator.StatusSkipped:
		return "⏭️  SKIPPED"
	default:
		return "⏸️  PENDING"
	}
}

func (r *ConsoleReporter) getResourceStatusIcon(status validator.ResourceStatus) string {
	switch status {
	case validator.ResourceExists:
		return "✅"
	case validator.ResourceNotFound:
		return "❌"
	case validator.ResourceMisconfigured:
		return "⚠️ "
	default:
		return "⏸️ "
	}
}

func (r *ConsoleReporter) Error(err error) {
	fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
}
