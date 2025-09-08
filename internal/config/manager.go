package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Manager struct {
	steps     map[int]*StepConfig
	resources map[string]*ResourceConfig
}

func NewManager() *Manager {
	return &Manager{
		steps:     make(map[int]*StepConfig),
		resources: make(map[string]*ResourceConfig),
	}
}

func (m *Manager) LoadStepConfig(stepNumber int) (*StepConfig, error) {
	if config, exists := m.steps[stepNumber]; exists {
		return config, nil
	}

	filename := filepath.Join("internal", "config", "configs", "steps", fmt.Sprintf("step%d.yaml", stepNumber))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read step config file: %w", err)
	}

	var config StepConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal step config: %w", err)
	}

	m.steps[stepNumber] = &config
	return &config, nil
}

func (m *Manager) LoadResourceConfig(resourceType string) (*ResourceConfig, error) {
	if config, exists := m.resources[resourceType]; exists {
		return config, nil
	}

	filename := filepath.Join("internal", "config", "configs", "resources", fmt.Sprintf("%s.yaml", resourceType))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource config file: %w", err)
	}

	var config ResourceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource config: %w", err)
	}

	m.resources[resourceType] = &config
	return &config, nil
}

func (m *Manager) GetValidationRules(resourceType string) ([]ValidationRule, error) {
	config, err := m.LoadResourceConfig(resourceType)
	if err != nil {
		return nil, err
	}
	return config.ValidationRules, nil
}

func (m *Manager) GetAllSteps() []*StepConfig {
	for i := 1; i <= 5; i++ {
		m.LoadStepConfig(i)
	}

	var steps []*StepConfig
	for i := 1; i <= 5; i++ {
		if step, exists := m.steps[i]; exists {
			steps = append(steps, step)
		}
	}
	return steps
}
