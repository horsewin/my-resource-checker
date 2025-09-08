package validator

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sbcntr2-test-tool/internal/aws"
	"sbcntr2-test-tool/internal/config"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

type ResourceValidator struct {
	awsClient     *aws.Client
	configManager *config.Manager
}

func NewResourceValidator(awsClient *aws.Client, configManager *config.Manager) *ResourceValidator {
	return &ResourceValidator{
		awsClient:     awsClient,
		configManager: configManager,
	}
}

func (v *ResourceValidator) CheckResourceExists(ctx context.Context, resourceType, resourceID string) (bool, map[string]interface{}, error) {
	switch resourceType {
	case "AWS::EC2::VPC":
		return v.checkVPC(ctx, resourceID)
	case "AWS::EC2::Subnet":
		return v.checkSubnet(ctx, resourceID)
	case "AWS::EC2::SecurityGroup":
		return v.checkSecurityGroup(ctx, resourceID)
	case "AWS::EC2::InternetGateway":
		return v.checkInternetGateway(ctx, resourceID)
	case "AWS::EC2::VPCEndpoint":
		return v.checkVPCEndpoint(ctx, resourceID)
	case "AWS::ECR::Repository":
		return v.checkECRRepository(ctx, resourceID)
	case "AWS::ECS::Cluster":
		return v.checkECSCluster(ctx, resourceID)
	case "AWS::ECS::TaskDefinition":
		return v.checkTaskDefinition(ctx, resourceID)
	case "AWS::ECS::Service":
		return v.checkECSService(ctx, resourceID)
	case "AWS::ElasticLoadBalancingV2::LoadBalancer":
		return v.checkLoadBalancer(ctx, resourceID)
	case "AWS::ElasticLoadBalancingV2::TargetGroup":
		return v.checkTargetGroup(ctx, resourceID)
	default:
		return v.checkCloudControlResource(ctx, resourceType, resourceID)
	}
}

func (v *ResourceValidator) ValidateRule(actualProps map[string]interface{}, rule config.ValidationRule) error {
	actualValue, exists := actualProps[rule.Property]
	if !exists && rule.Type == "exists" {
		return fmt.Errorf("%s: property '%s' not found", rule.ErrorMessage, rule.Property)
	}

	switch rule.Type {
	case "property":
		return v.validateProperty(actualValue, rule)
	case "exists":
		if !exists {
			return fmt.Errorf(rule.ErrorMessage)
		}
	case "count":
		return v.validateCount(actualValue, rule)
	case "custom":
		return v.validateCustom(actualProps, rule)
	}

	return nil
}

func (v *ResourceValidator) validateProperty(actual interface{}, rule config.ValidationRule) error {
	switch rule.Operator {
	case "eq":
		if !reflect.DeepEqual(actual, rule.Expected) {
			return fmt.Errorf("%s: expected %v, got %v", rule.ErrorMessage, rule.Expected, actual)
		}
	case "ne":
		if reflect.DeepEqual(actual, rule.Expected) {
			return fmt.Errorf("%s: value should not be %v", rule.ErrorMessage, rule.Expected)
		}
	case "gt":
		if !v.compareNumbers(actual, rule.Expected, ">") {
			return fmt.Errorf("%s: %v should be greater than %v", rule.ErrorMessage, actual, rule.Expected)
		}
	case "lt":
		if !v.compareNumbers(actual, rule.Expected, "<") {
			return fmt.Errorf("%s: %v should be less than %v", rule.ErrorMessage, actual, rule.Expected)
		}
	case "contains":
		if !v.contains(actual, rule.Expected) {
			return fmt.Errorf("%s: %v should contain %v", rule.ErrorMessage, actual, rule.Expected)
		}
	case "regex":
		if !v.matchRegex(actual, rule.Expected) {
			return fmt.Errorf("%s: %v does not match pattern %v", rule.ErrorMessage, actual, rule.Expected)
		}
	}

	return nil
}

func (v *ResourceValidator) validateCount(actual interface{}, rule config.ValidationRule) error {
	count := 0
	switch val := actual.(type) {
	case []interface{}:
		count = len(val)
	case map[string]interface{}:
		count = len(val)
	default:
		return fmt.Errorf("cannot count non-collection type")
	}

	expected, ok := rule.Expected.(int)
	if !ok {
		return fmt.Errorf("expected count must be an integer")
	}

	if count != expected {
		return fmt.Errorf("%s: expected count %d, got %d", rule.ErrorMessage, expected, count)
	}

	return nil
}

func (v *ResourceValidator) validateCustom(props map[string]interface{}, rule config.ValidationRule) error {
	return nil
}

func (v *ResourceValidator) compareNumbers(actual, expected interface{}, operator string) bool {
	actualFloat, ok1 := toFloat64(actual)
	expectedFloat, ok2 := toFloat64(expected)

	if !ok1 || !ok2 {
		return false
	}

	switch operator {
	case ">":
		return actualFloat > expectedFloat
	case "<":
		return actualFloat < expectedFloat
	case ">=":
		return actualFloat >= expectedFloat
	case "<=":
		return actualFloat <= expectedFloat
	default:
		return false
	}
}

func (v *ResourceValidator) contains(actual, expected interface{}) bool {
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)
	return strings.Contains(actualStr, expectedStr)
}

func (v *ResourceValidator) matchRegex(actual, expected interface{}) bool {
	actualStr := fmt.Sprintf("%v", actual)
	pattern := fmt.Sprintf("%v", expected)

	matched, err := regexp.MatchString(pattern, actualStr)
	if err != nil {
		return false
	}

	return matched
}

func toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	default:
		return 0, false
	}
}

func (v *ResourceValidator) checkVPC(ctx context.Context, vpcID string) (bool, map[string]interface{}, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcID},
	}

	result, err := v.awsClient.EC2.DescribeVpcs(ctx, input)
	if err != nil {
		return false, nil, nil
	}

	if len(result.Vpcs) == 0 {
		return false, nil, nil
	}

	vpc := result.Vpcs[0]
	props := map[string]interface{}{
		"VpcId":     *vpc.VpcId,
		"CidrBlock": *vpc.CidrBlock,
		"State":     string(vpc.State),
	}

	return true, props, nil
}

func (v *ResourceValidator) checkSubnet(ctx context.Context, subnetID string) (bool, map[string]interface{}, error) {
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{subnetID},
	}

	result, err := v.awsClient.EC2.DescribeSubnets(ctx, input)
	if err != nil {
		return false, nil, nil
	}

	if len(result.Subnets) == 0 {
		return false, nil, nil
	}

	subnet := result.Subnets[0]
	props := map[string]interface{}{
		"SubnetId":         *subnet.SubnetId,
		"CidrBlock":        *subnet.CidrBlock,
		"AvailabilityZone": *subnet.AvailabilityZone,
		"VpcId":            *subnet.VpcId,
	}

	return true, props, nil
}

func (v *ResourceValidator) checkSecurityGroup(ctx context.Context, sgID string) (bool, map[string]interface{}, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{sgID},
	}

	result, err := v.awsClient.EC2.DescribeSecurityGroups(ctx, input)
	if err != nil {
		return false, nil, nil
	}

	if len(result.SecurityGroups) == 0 {
		return false, nil, nil
	}

	sg := result.SecurityGroups[0]
	props := map[string]interface{}{
		"GroupId":   *sg.GroupId,
		"GroupName": *sg.GroupName,
		"VpcId":     *sg.VpcId,
	}

	var ingressRules []map[string]interface{}
	for _, rule := range sg.IpPermissions {
		ingressRule := map[string]interface{}{
			"FromPort":   rule.FromPort,
			"ToPort":     rule.ToPort,
			"IpProtocol": rule.IpProtocol,
		}
		ingressRules = append(ingressRules, ingressRule)
	}
	props["IngressRules"] = ingressRules

	return true, props, nil
}

func (v *ResourceValidator) checkInternetGateway(ctx context.Context, igwID string) (bool, map[string]interface{}, error) {
	input := &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []string{igwID},
	}

	result, err := v.awsClient.EC2.DescribeInternetGateways(ctx, input)
	if err != nil {
		return false, nil, nil
	}

	if len(result.InternetGateways) == 0 {
		return false, nil, nil
	}

	igw := result.InternetGateways[0]
	props := map[string]interface{}{
		"InternetGatewayId": *igw.InternetGatewayId,
	}

	if len(igw.Attachments) > 0 && igw.Attachments[0].VpcId != nil {
		props["AttachedVpcId"] = *igw.Attachments[0].VpcId
	}

	return true, props, nil
}

func (v *ResourceValidator) checkVPCEndpoint(ctx context.Context, endpointID string) (bool, map[string]interface{}, error) {
	input := &ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: []string{endpointID},
	}

	result, err := v.awsClient.EC2.DescribeVpcEndpoints(ctx, input)
	if err != nil {
		return false, nil, nil
	}

	if len(result.VpcEndpoints) == 0 {
		return false, nil, nil
	}

	endpoint := result.VpcEndpoints[0]
	props := map[string]interface{}{
		"VpcEndpointId": *endpoint.VpcEndpointId,
		"ServiceName":   *endpoint.ServiceName,
		"VpcId":         *endpoint.VpcId,
	}

	return true, props, nil
}

func (v *ResourceValidator) checkECRRepository(ctx context.Context, repoName string) (bool, map[string]interface{}, error) {
	input := &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repoName},
	}

	result, err := v.awsClient.ECR.DescribeRepositories(ctx, input)
	if err != nil {
		return false, nil, nil
	}

	if len(result.Repositories) == 0 {
		return false, nil, nil
	}

	repo := result.Repositories[0]
	props := map[string]interface{}{
		"RepositoryName": *repo.RepositoryName,
		"RepositoryUri":  *repo.RepositoryUri,
	}

	if repo.EncryptionConfiguration != nil {
		props["EncryptionType"] = string(repo.EncryptionConfiguration.EncryptionType)
	}

	return true, props, nil
}

func (v *ResourceValidator) checkECSCluster(ctx context.Context, clusterName string) (bool, map[string]interface{}, error) {
	input := &ecs.DescribeClustersInput{
		Clusters: []string{clusterName},
	}

	result, err := v.awsClient.ECS.DescribeClusters(ctx, input)
	if err != nil {
		return false, nil, nil
	}

	if len(result.Clusters) == 0 {
		return false, nil, nil
	}

	cluster := result.Clusters[0]
	props := map[string]interface{}{
		"ClusterName": *cluster.ClusterName,
		"Status":      *cluster.Status,
	}

	return true, props, nil
}

func (v *ResourceValidator) checkTaskDefinition(ctx context.Context, taskDefName string) (bool, map[string]interface{}, error) {
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: &taskDefName,
	}

	result, err := v.awsClient.ECS.DescribeTaskDefinition(ctx, input)
	if err != nil {
		return false, nil, nil
	}

	if result.TaskDefinition == nil {
		return false, nil, nil
	}

	taskDef := result.TaskDefinition
	props := map[string]interface{}{
		"Family":   *taskDef.Family,
		"Revision": taskDef.Revision,
		"Status":   string(taskDef.Status),
	}

	return true, props, nil
}

func (v *ResourceValidator) checkECSService(ctx context.Context, serviceName string) (bool, map[string]interface{}, error) {
	clusters, err := v.awsClient.ECS.ListClusters(ctx, &ecs.ListClustersInput{})
	if err != nil {
		return false, nil, err
	}

	for _, clusterArn := range clusters.ClusterArns {
		input := &ecs.DescribeServicesInput{
			Cluster:  &clusterArn,
			Services: []string{serviceName},
		}

		result, err := v.awsClient.ECS.DescribeServices(ctx, input)
		if err != nil {
			continue
		}

		if len(result.Services) > 0 && result.Services[0].Status != nil {
			service := result.Services[0]
			props := map[string]interface{}{
				"ServiceName": *service.ServiceName,
				"Status":      *service.Status,
			}

			props["DesiredCount"] = service.DesiredCount
			props["RunningCount"] = service.RunningCount

			return true, props, nil
		}
	}

	return false, nil, nil
}

func (v *ResourceValidator) checkLoadBalancer(ctx context.Context, albName string) (bool, map[string]interface{}, error) {
	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{
		Names: []string{albName},
	}

	result, err := v.awsClient.ELBv2.DescribeLoadBalancers(ctx, input)
	if err != nil {
		return false, nil, nil
	}

	if len(result.LoadBalancers) == 0 {
		return false, nil, nil
	}

	alb := result.LoadBalancers[0]
	props := map[string]interface{}{
		"LoadBalancerName": *alb.LoadBalancerName,
		"State":            string(alb.State.Code),
		"Type":             string(alb.Type),
	}

	return true, props, nil
}

func (v *ResourceValidator) checkTargetGroup(ctx context.Context, tgName string) (bool, map[string]interface{}, error) {
	input := &elasticloadbalancingv2.DescribeTargetGroupsInput{
		Names: []string{tgName},
	}

	result, err := v.awsClient.ELBv2.DescribeTargetGroups(ctx, input)
	if err != nil {
		return false, nil, nil
	}

	if len(result.TargetGroups) == 0 {
		return false, nil, nil
	}

	tg := result.TargetGroups[0]
	props := map[string]interface{}{
		"TargetGroupName": *tg.TargetGroupName,
		"Protocol":        string(tg.Protocol),
		"Port":            *tg.Port,
	}

	return true, props, nil
}

func (v *ResourceValidator) checkCloudControlResource(ctx context.Context, resourceType, resourceID string) (bool, map[string]interface{}, error) {
	resource, err := v.awsClient.GetResource(ctx, resourceType, resourceID)
	if err != nil {
		return false, nil, nil
	}

	return true, resource.Properties, nil
}
