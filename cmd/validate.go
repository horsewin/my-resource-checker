package cmd

import (
	"fmt"
	"sbcntr2-test-tool/internal/aws"
	"sbcntr2-test-tool/internal/config"
	"sbcntr2-test-tool/internal/reporter"
	"sbcntr2-test-tool/internal/validator"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var step int
var allSteps bool

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate AWS resources for a specific step or all steps",
	Long:  `Validates that AWS resources have been correctly created and configured according to the hands-on documentation.`,
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().IntVarP(&step, "step", "s", 0, "Step number to validate (1-6)")
	validateCmd.Flags().BoolVarP(&allSteps, "all", "a", false, "Validate all steps")
}

func runValidate(cmd *cobra.Command, args []string) error {
	if !allSteps && (step < 1 || step > 6) {
		return fmt.Errorf("please specify a valid step number (1-6) or use --all flag")
	}

	awsClient, err := aws.NewClient(
		viper.GetString("region"),
		viper.GetString("profile"),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize AWS client: %w", err)
	}

	configManager := config.NewManager()
	validationEngine := validator.NewEngine(awsClient, configManager)

	var rep reporter.Reporter
	if viper.GetString("output") == "json" {
		rep = reporter.NewJSONReporter()
	} else {
		rep = reporter.NewConsoleReporter(viper.GetBool("verbose"))
	}

	if allSteps {
		summary, err := validationEngine.ValidateAllSteps()
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		return rep.ReportSummary(summary)
	} else {
		result, err := validationEngine.ValidateStep(step)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		return rep.ReportResult(result)
	}
}
