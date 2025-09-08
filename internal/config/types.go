package config

type StepConfig struct {
	Number               int                  `yaml:"number"`
	Name                 string               `yaml:"name"`
	Description          string               `yaml:"description"`
	Resources            []ResourceDefinition `yaml:"resources"`
	CloudFormationStacks []string             `yaml:"cloudformation_stacks"`
	Dependencies         []int                `yaml:"dependencies"`
}

type ResourceDefinition struct {
	Type            string   `yaml:"type"`
	Identifier      string   `yaml:"identifier"`
	Name            string   `yaml:"name"`
	Required        bool     `yaml:"required"`
	ValidationRules []string `yaml:"validation_rules"`
}

type ResourceConfig struct {
	Type            string           `yaml:"type"`
	ValidationRules []ValidationRule `yaml:"validation_rules"`
}

type ValidationRule struct {
	Name         string      `yaml:"name"`
	Type         string      `yaml:"type"`
	Property     string      `yaml:"property"`
	Expected     interface{} `yaml:"expected"`
	Operator     string      `yaml:"operator"`
	ErrorMessage string      `yaml:"error_message"`
	Severity     string      `yaml:"severity"`
}
