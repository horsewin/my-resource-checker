package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type Client struct {
	cfg            aws.Config
	CloudControl   *cloudcontrol.Client
	CloudFormation *cloudformation.Client
	EC2            *ec2.Client
	ECR            *ecr.Client
	ECS            *ecs.Client
	ELBv2          *elasticloadbalancingv2.Client
	IAM            *iam.Client
}

func NewClient(region string, profile string) (*Client, error) {
	ctx := context.Background()

	var optFns []func(*config.LoadOptions) error

	if region != "" {
		optFns = append(optFns, config.WithRegion(region))
	}

	if profile != "" {
		optFns = append(optFns, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &Client{
		cfg:            cfg,
		CloudControl:   cloudcontrol.NewFromConfig(cfg),
		CloudFormation: cloudformation.NewFromConfig(cfg),
		EC2:            ec2.NewFromConfig(cfg),
		ECR:            ecr.NewFromConfig(cfg),
		ECS:            ecs.NewFromConfig(cfg),
		ELBv2:          elasticloadbalancingv2.NewFromConfig(cfg),
		IAM:            iam.NewFromConfig(cfg),
	}, nil
}

func (c *Client) GetRegion() string {
	return c.cfg.Region
}
