package validator

import "time"

type ValidationStatus int

const (
	StatusPending ValidationStatus = iota
	StatusPassed
	StatusFailed
	StatusWarning
	StatusSkipped
)

type ResourceStatus int

const (
	ResourceNotFound ResourceStatus = iota
	ResourceExists
	ResourceMisconfigured
	ResourcePending
)

type ErrorType int

const (
	ErrorResourceNotFound ErrorType = iota
	ErrorPropertyMismatch
	ErrorConfigurationInvalid
	ErrorAWSAPIFailure
	ErrorAuthenticationFailure
	ErrorPermissionDenied
	ErrorNetworkFailure
)

type ValidationResult struct {
	StepNumber int
	StepName   string
	Status     ValidationStatus
	Resources  []ResourceResult
	Errors     []ValidationError
	Warnings   []ValidationWarning
	Duration   time.Duration
}

type ResourceResult struct {
	Type     string
	ID       string
	Name     string
	Status   ResourceStatus
	Expected map[string]interface{}
	Actual   map[string]interface{}
	Errors   []string
	Warnings []string
}

type ValidationError struct {
	Type        ErrorType
	Resource    string
	Property    string
	Expected    interface{}
	Actual      interface{}
	Message     string
	Suggestion  string
	DocumentRef string
}

type ValidationWarning struct {
	Resource string
	Message  string
}

type ValidationSummary struct {
	TotalSteps   int
	PassedSteps  int
	FailedSteps  int
	SkippedSteps int
	Results      []ValidationResult
}

func (s ValidationStatus) String() string {
	switch s {
	case StatusPending:
		return "PENDING"
	case StatusPassed:
		return "PASSED"
	case StatusFailed:
		return "FAILED"
	case StatusWarning:
		return "WARNING"
	case StatusSkipped:
		return "SKIPPED"
	default:
		return "UNKNOWN"
	}
}

func (s ResourceStatus) String() string {
	switch s {
	case ResourceNotFound:
		return "NOT_FOUND"
	case ResourceExists:
		return "EXISTS"
	case ResourceMisconfigured:
		return "MISCONFIGURED"
	case ResourcePending:
		return "PENDING"
	default:
		return "UNKNOWN"
	}
}
