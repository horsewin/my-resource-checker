package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type CloudFormationStack struct {
	Name       string
	Status     string
	Resources  []CloudFormationResource
	Outputs    map[string]string
	Parameters map[string]string
}

type CloudFormationResource struct {
	LogicalID  string
	PhysicalID string
	Type       string
	Status     string
}

func (c *Client) GetCloudFormationStack(ctx context.Context, stackName string) (*CloudFormationStack, error) {
	describeInput := &cloudformation.DescribeStacksInput{
		StackName: &stackName,
	}

	result, err := c.CloudFormation.DescribeStacks(ctx, describeInput)
	if err != nil {
		return nil, fmt.Errorf("failed to describe stack %s: %w", stackName, err)
	}

	if len(result.Stacks) == 0 {
		return nil, fmt.Errorf("stack %s not found", stackName)
	}

	stack := result.Stacks[0]

	cfStack := &CloudFormationStack{
		Name:       stackName,
		Status:     string(stack.StackStatus),
		Outputs:    make(map[string]string),
		Parameters: make(map[string]string),
	}

	for _, output := range stack.Outputs {
		if output.OutputKey != nil && output.OutputValue != nil {
			cfStack.Outputs[*output.OutputKey] = *output.OutputValue
		}
	}

	for _, param := range stack.Parameters {
		if param.ParameterKey != nil && param.ParameterValue != nil {
			cfStack.Parameters[*param.ParameterKey] = *param.ParameterValue
		}
	}

	resources, err := c.ListCloudFormationResources(ctx, stackName)
	if err != nil {
		return nil, err
	}
	cfStack.Resources = resources

	return cfStack, nil
}

func (c *Client) ListCloudFormationResources(ctx context.Context, stackName string) ([]CloudFormationResource, error) {
	input := &cloudformation.ListStackResourcesInput{
		StackName: &stackName,
	}

	var resources []CloudFormationResource
	paginator := cloudformation.NewListStackResourcesPaginator(c.CloudFormation, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list stack resources: %w", err)
		}

		for _, res := range page.StackResourceSummaries {
			resource := CloudFormationResource{
				Status: string(res.ResourceStatus),
			}

			if res.LogicalResourceId != nil {
				resource.LogicalID = *res.LogicalResourceId
			}
			if res.PhysicalResourceId != nil {
				resource.PhysicalID = *res.PhysicalResourceId
			}
			if res.ResourceType != nil {
				resource.Type = *res.ResourceType
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

func (c *Client) StackExists(ctx context.Context, stackName string) bool {
	_, err := c.GetCloudFormationStack(ctx, stackName)
	return err == nil
}

func (c *Client) GetStackStatus(ctx context.Context, stackName string) (types.StackStatus, error) {
	stack, err := c.GetCloudFormationStack(ctx, stackName)
	if err != nil {
		return "", err
	}
	return types.StackStatus(stack.Status), nil
}
