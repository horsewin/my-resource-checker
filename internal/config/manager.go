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

	// リソースタイプからファイル名にマッピング
	fileMap := map[string]string{
		"AWS::EC2::VPC":                             "vpc.yaml",
		"AWS::EC2::Subnet":                          "subnet.yaml",
		"AWS::EC2::SecurityGroup":                   "security_group.yaml",
		"AWS::EC2::InternetGateway":                 "internet_gateway.yaml",
		"AWS::EC2::VPCEndpoint":                     "vpce.yaml",
		"AWS::ECR::Repository":                      "ecr.yaml",
		"AWS::ECS::Cluster":                         "ecs.yaml",
		"AWS::ECS::TaskDefinition":                  "ecs_task_definition.yaml",
		"AWS::ECS::Service":                         "ecs_service.yaml",
		"AWS::ElasticLoadBalancingV2::LoadBalancer": "alb.yaml",
		"AWS::ElasticLoadBalancingV2::TargetGroup":  "target_group.yaml",
		"AWS::RDS::DBCluster":                       "aurora.yaml",
		"AWS::RDS::DBInstance":                      "rds_instance.yaml",
		"AWS::RDS::DBSubnetGroup":                   "rds_subnet_group.yaml",
		"AWS::IAM::Role":                            "iam_role.yaml",
	}

	yamlFile, ok := fileMap[resourceType]
	if !ok {
		// マッピングがない場合はデフォルトで空のルールを返す
		m.resources[resourceType] = &ResourceConfig{
			Type:            resourceType,
			ValidationRules: []ValidationRule{},
		}
		return m.resources[resourceType], nil
	}

	filename := filepath.Join("internal", "config", "configs", "resources", yamlFile)
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
