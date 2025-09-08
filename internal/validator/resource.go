package validator

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sbcntr2-test-tool/internal/aws"
	"sbcntr2-test-tool/internal/config"
	"strconv"
	"strings"

	awsutil "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

func (v *ResourceValidator) CheckResourceExists(ctx context.Context, resourceType, resourceName string) (bool, map[string]interface{}, error) {
	switch resourceType {
	case "AWS::EC2::VPC":
		return v.checkVPC(ctx, resourceName)
	case "AWS::EC2::Subnet":
		return v.checkSubnet(ctx, resourceName)
	case "AWS::EC2::SecurityGroup":
		return v.checkSecurityGroup(ctx, resourceName)
	case "AWS::EC2::InternetGateway":
		return v.checkInternetGateway(ctx, resourceName)
	case "AWS::EC2::VPCEndpoint":
		return v.checkVPCEndpoint(ctx, resourceName)
	case "AWS::ECR::Repository":
		return v.checkECRRepository(ctx, resourceName)
	case "AWS::ECS::Cluster":
		return v.checkECSCluster(ctx, resourceName)
	case "AWS::ECS::TaskDefinition":
		return v.checkTaskDefinition(ctx, resourceName)
	case "AWS::ECS::Service":
		return v.checkECSService(ctx, resourceName)
	case "AWS::ElasticLoadBalancingV2::LoadBalancer":
		return v.checkLoadBalancer(ctx, resourceName)
	case "AWS::ElasticLoadBalancingV2::TargetGroup":
		return v.checkTargetGroup(ctx, resourceName)
	default:
		return v.checkCloudControlResource(ctx, resourceType, resourceName)
	}
}

func (v *ResourceValidator) ValidateRule(actualProps map[string]interface{}, rule config.ValidationRule) error {
	// ネストされたプロパティや配列アクセスに対応
	actualValue, exists := v.getNestedProperty(actualProps, rule.Property)
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

// getNestedProperty はネストされたプロパティや配列要素にアクセスする
// 例: "IngressRules[0].FromPort" -> IngressRules配列の0番目のFromPortプロパティ
func (v *ResourceValidator) getNestedProperty(props map[string]interface{}, path string) (interface{}, bool) {
	if path == "" {
		return nil, false
	}

	// パスを分割（例: "IngressRules[0].FromPort" -> ["IngressRules[0]", "FromPort"]）
	parts := strings.Split(path, ".")
	var current interface{} = props

	for _, part := range parts {
		// 配列インデックスを確認
		if strings.Contains(part, "[") {
			// 配列名とインデックスを分離
			arrayMatch := regexp.MustCompile(`^([^\[]+)\[(\d+)\]$`).FindStringSubmatch(part)
			if len(arrayMatch) != 3 {
				return nil, false
			}

			arrayName := arrayMatch[1]
			index, err := strconv.Atoi(arrayMatch[2])
			if err != nil {
				return nil, false
			}

			// 現在のオブジェクトから配列を取得
			switch obj := current.(type) {
			case map[string]interface{}:
				arr, ok := obj[arrayName]
				if !ok {
					return nil, false
				}

				// 配列要素にアクセス
				switch a := arr.(type) {
				case []interface{}:
					if index < 0 || index >= len(a) {
						return nil, false
					}
					current = a[index]
				case []map[string]interface{}:
					if index < 0 || index >= len(a) {
						return nil, false
					}
					current = a[index]
				case []string:
					if index < 0 || index >= len(a) {
						return nil, false
					}
					current = a[index]
				default:
					return nil, false
				}
			default:
				return nil, false
			}
		} else {
			// 通常のプロパティアクセス
			switch obj := current.(type) {
			case map[string]interface{}:
				val, ok := obj[part]
				if !ok {
					return nil, false
				}
				current = val
			default:
				return nil, false
			}
		}
	}

	return current, true
}

func (v *ResourceValidator) validateProperty(actual interface{}, rule config.ValidationRule) error {
	switch rule.Operator {
	case "eq":
		// 数値の場合は型変換を試みる
		if actualNum, ok := toFloat64(actual); ok {
			if expectedNum, ok := toFloat64(rule.Expected); ok {
				if actualNum != expectedNum {
					return fmt.Errorf("%s: expected %v, got %v", rule.ErrorMessage, rule.Expected, actual)
				}
				return nil
			}
		}

		// それ以外の場合は通常の比較
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
	case "ge":
		if !v.compareNumbers(actual, rule.Expected, ">=") {
			return fmt.Errorf("%s: %v should be greater than or equal to %v", rule.ErrorMessage, actual, rule.Expected)
		}
	case "le":
		if !v.compareNumbers(actual, rule.Expected, "<=") {
			return fmt.Errorf("%s: %v should be less than or equal to %v", rule.ErrorMessage, actual, rule.Expected)
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
	case []string:
		count = len(val)
	default:
		return fmt.Errorf("cannot count non-collection type")
	}

	expected, ok := rule.Expected.(int)
	if !ok {
		return fmt.Errorf("expected count must be an integer")
	}

	// operatorが指定されていない場合はデフォルトで"eq"を使用
	operator := rule.Operator
	if operator == "" {
		operator = "eq"
	}

	switch operator {
	case "eq":
		if count != expected {
			return fmt.Errorf("%s: expected count %d, got %d", rule.ErrorMessage, expected, count)
		}
	case "ne":
		if count == expected {
			return fmt.Errorf("%s: count should not be %d", rule.ErrorMessage, expected)
		}
	case "gt":
		if count <= expected {
			return fmt.Errorf("%s: count %d should be greater than %d", rule.ErrorMessage, count, expected)
		}
	case "lt":
		if count >= expected {
			return fmt.Errorf("%s: count %d should be less than %d", rule.ErrorMessage, count, expected)
		}
	case "ge":
		if count < expected {
			return fmt.Errorf("%s: count %d should be greater than or equal to %d", rule.ErrorMessage, count, expected)
		}
	case "le":
		if count > expected {
			return fmt.Errorf("%s: count %d should be less than or equal to %d", rule.ErrorMessage, count, expected)
		}
	default:
		return fmt.Errorf("unsupported operator for count: %s", operator)
	}

	return nil
}

func (v *ResourceValidator) validateCustom(props map[string]interface{}, rule config.ValidationRule) error {
	// カスタムバリデーションは現在使用しない
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
	case int16:
		return float64(v), true
	case int8:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint8:
		return float64(v), true
	default:
		return 0, false
	}
}

func (v *ResourceValidator) checkVPC(ctx context.Context, vpcName string) (bool, map[string]interface{}, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: []ec2types.Filter{
			{
				Name:   awsutil.String("tag:Name"),
				Values: []string{vpcName},
			},
		},
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

func (v *ResourceValidator) checkSubnet(ctx context.Context, subnetName string) (bool, map[string]interface{}, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{
				Name:   awsutil.String("tag:Name"),
				Values: []string{subnetName},
			},
		},
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

func (v *ResourceValidator) checkSecurityGroup(ctx context.Context, sgName string) (bool, map[string]interface{}, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: []ec2types.Filter{
			{
				Name:   awsutil.String("tag:Name"),
				Values: []string{sgName},
			},
		},
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
			"IpProtocol": rule.IpProtocol,
		}

		// ポートをデリファレンスして格納
		if rule.FromPort != nil {
			ingressRule["FromPort"] = *rule.FromPort
		}
		if rule.ToPort != nil {
			ingressRule["ToPort"] = *rule.ToPort
		}

		// CIDRブロックを追加
		if len(rule.IpRanges) > 0 {
			var cidrs []string
			for _, ipRange := range rule.IpRanges {
				if ipRange.CidrIp != nil {
					cidrs = append(cidrs, *ipRange.CidrIp)
				}
			}
			ingressRule["CidrBlocks"] = cidrs
		}

		// ソースセキュリティグループを追加
		if len(rule.UserIdGroupPairs) > 0 {
			var sourceGroups []string
			var sourceGroupNames []string

			for _, group := range rule.UserIdGroupPairs {
				if group.GroupId != nil {
					sourceGroups = append(sourceGroups, *group.GroupId)

					// セキュリティグループIDからNameタグを取得
					sgName, err := v.getSecurityGroupName(ctx, *group.GroupId)
					if err == nil && sgName != "" {
						sourceGroupNames = append(sourceGroupNames, sgName)
					}
				}
			}
			ingressRule["SourceSecurityGroups"] = sourceGroups
			if len(sourceGroupNames) > 0 {
				ingressRule["SourceSecurityGroupNames"] = sourceGroupNames
			}
		}

		ingressRules = append(ingressRules, ingressRule)
	}

	props["IngressRules"] = ingressRules

	return true, props, nil
}

// getSecurityGroupName はセキュリティグループIDからNameタグを取得する
func (v *ResourceValidator) getSecurityGroupName(ctx context.Context, sgID string) (string, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{sgID},
	}

	result, err := v.awsClient.EC2.DescribeSecurityGroups(ctx, input)
	if err != nil {
		return "", err
	}

	if len(result.SecurityGroups) == 0 {
		return "", fmt.Errorf("security group not found: %s", sgID)
	}

	sg := result.SecurityGroups[0]
	for _, tag := range sg.Tags {
		if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
			return *tag.Value, nil
		}
	}

	return "", nil
}

func (v *ResourceValidator) checkInternetGateway(ctx context.Context, igwName string) (bool, map[string]interface{}, error) {
	input := &ec2.DescribeInternetGatewaysInput{
		Filters: []ec2types.Filter{
			{
				Name:   awsutil.String("tag:Name"),
				Values: []string{igwName},
			},
		},
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

func (v *ResourceValidator) checkVPCEndpoint(ctx context.Context, endpointName string) (bool, map[string]interface{}, error) {
	input := &ec2.DescribeVpcEndpointsInput{
		Filters: []ec2types.Filter{
			{
				Name:   awsutil.String("tag:Name"),
				Values: []string{endpointName},
			},
		},
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

	// ImageTagMutabilityを追加
	props["ImageTagMutability"] = string(repo.ImageTagMutability)

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

func (v *ResourceValidator) checkCloudControlResource(ctx context.Context, resourceType, resourceName string) (bool, map[string]interface{}, error) {
	resource, err := v.awsClient.GetResource(ctx, resourceType, resourceName)
	if err != nil {
		return false, nil, nil
	}

	return true, resource.Properties, nil
}
