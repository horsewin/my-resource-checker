package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
)

type CloudControlResource struct {
	Type       string
	Identifier string
	Properties map[string]interface{}
	Status     string
}

func (c *Client) GetResource(ctx context.Context, resourceType, resourceID string) (*CloudControlResource, error) {
	input := &cloudcontrol.GetResourceInput{
		TypeName:   &resourceType,
		Identifier: &resourceID,
	}

	result, err := c.CloudControl.GetResource(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource %s/%s: %w", resourceType, resourceID, err)
	}

	var properties map[string]interface{}
	if result.ResourceDescription != nil && result.ResourceDescription.Properties != nil {
		err = json.Unmarshal([]byte(*result.ResourceDescription.Properties), &properties)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal resource properties: %w", err)
		}
	}

	resource := &CloudControlResource{
		Type:       resourceType,
		Identifier: resourceID,
		Properties: properties,
		Status:     "ACTIVE", // Cloud Control API doesn't provide status directly
	}

	return resource, nil
}

func (c *Client) ListResources(ctx context.Context, resourceType string) ([]*CloudControlResource, error) {
	input := &cloudcontrol.ListResourcesInput{
		TypeName: &resourceType,
	}

	var resources []*CloudControlResource
	paginator := cloudcontrol.NewListResourcesPaginator(c.CloudControl, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list resources of type %s: %w", resourceType, err)
		}

		for _, desc := range page.ResourceDescriptions {
			var properties map[string]interface{}
			if desc.Properties != nil {
				err = json.Unmarshal([]byte(*desc.Properties), &properties)
				if err != nil {
					continue
				}
			}

			resource := &CloudControlResource{
				Type:       resourceType,
				Properties: properties,
			}

			if desc.Identifier != nil {
				resource.Identifier = *desc.Identifier
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}
