package validator

import (
	"context"
	"fmt"
	"sbcntr2-test-tool/internal/aws"
	"sbcntr2-test-tool/internal/config"
	"time"
)

type Engine struct {
	awsClient     *aws.Client
	configManager *config.Manager
	cache         map[string]interface{}
}

func NewEngine(awsClient *aws.Client, configManager *config.Manager) *Engine {
	return &Engine{
		awsClient:     awsClient,
		configManager: configManager,
		cache:         make(map[string]interface{}),
	}
}

func (e *Engine) ValidateStep(stepNumber int) (*ValidationResult, error) {
	startTime := time.Now()

	stepConfig, err := e.configManager.LoadStepConfig(stepNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to load step config: %w", err)
	}

	result := &ValidationResult{
		StepNumber: stepNumber,
		StepName:   stepConfig.Name,
		Status:     StatusPending,
		Resources:  []ResourceResult{},
		Errors:     []ValidationError{},
		Warnings:   []ValidationWarning{},
	}

	ctx := context.Background()

	for _, cfStack := range stepConfig.CloudFormationStacks {
		if !e.awsClient.StackExists(ctx, cfStack) {
			result.Errors = append(result.Errors, ValidationError{
				Type:        ErrorResourceNotFound,
				Resource:    cfStack,
				Message:     fmt.Sprintf("CloudFormation stack '%s' not found", cfStack),
				Suggestion:  fmt.Sprintf("Please create the stack '%s' as described in the handbook", cfStack),
				DocumentRef: fmt.Sprintf("Step %d", stepNumber),
			})
		}
	}

	for _, resource := range stepConfig.Resources {
		resResult := e.validateResource(ctx, resource)
		result.Resources = append(result.Resources, resResult)

		if resResult.Status == ResourceNotFound && resource.Required {
			result.Errors = append(result.Errors, ValidationError{
				Type:        ErrorResourceNotFound,
				Resource:    resource.Name,
				Message:     fmt.Sprintf("Required resource '%s' not found", resource.Name),
				Suggestion:  fmt.Sprintf("Please create the resource '%s' as described in step %d", resource.Name, stepNumber),
				DocumentRef: fmt.Sprintf("Step %d", stepNumber),
			})
		}
	}

	result.Status = e.determineStatus(result)
	result.Duration = time.Since(startTime)

	return result, nil
}

func (e *Engine) ValidateAllSteps() (*ValidationSummary, error) {
	summary := &ValidationSummary{
		TotalSteps:   6,
		PassedSteps:  0,
		FailedSteps:  0,
		SkippedSteps: 0,
		Results:      []ValidationResult{},
	}

	for i := 1; i <= 6; i++ {
		result, err := e.ValidateStep(i)
		if err != nil {
			summary.SkippedSteps++
			continue
		}

		summary.Results = append(summary.Results, *result)

		switch result.Status {
		case StatusPassed:
			summary.PassedSteps++
		case StatusFailed:
			summary.FailedSteps++
		case StatusSkipped:
			summary.SkippedSteps++
		}
	}

	return summary, nil
}

func (e *Engine) validateResource(ctx context.Context, resource config.ResourceDefinition) ResourceResult {
	result := ResourceResult{
		Type:     resource.Type,
		ID:       resource.Identifier,
		Name:     resource.Name,
		Status:   ResourceNotFound,
		Expected: make(map[string]interface{}),
		Actual:   make(map[string]interface{}),
		Errors:   []string{},
		Warnings: []string{},
	}

	validator := NewResourceValidator(e.awsClient, e.configManager)

	exists, actualProps, err := validator.CheckResourceExists(ctx, resource.Type, resource.Name)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to check resource: %v", err))
		return result
	}

	if !exists {
		return result
	}

	result.Status = ResourceExists
	result.Actual = actualProps


	for _, ruleName := range resource.ValidationRules {
		rules, err := e.configManager.GetValidationRules(resource.Type)
		if err != nil {
			continue
		}


		for _, rule := range rules {
			if rule.Name == ruleName {

				result.Expected[rule.Property] = rule.Expected

				if err := validator.ValidateRule(actualProps, rule); err != nil {
					if rule.Severity == "error" {
						result.Status = ResourceMisconfigured
						result.Errors = append(result.Errors, err.Error())
					} else {
						result.Warnings = append(result.Warnings, err.Error())
					}
				}
			}
		}
	}

	return result
}

func (e *Engine) determineStatus(result *ValidationResult) ValidationStatus {
	if len(result.Errors) > 0 {
		return StatusFailed
	}

	for _, resource := range result.Resources {
		if resource.Status == ResourceNotFound || resource.Status == ResourceMisconfigured {
			return StatusFailed
		}
	}

	if len(result.Warnings) > 0 {
		return StatusWarning
	}

	return StatusPassed
}
