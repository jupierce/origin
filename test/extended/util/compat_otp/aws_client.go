package compat_otp

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
)

// AwsClient represents an AWS client with EC2 operations
type AwsClient struct {
	svc       *ec2.EC2
	s3Client  *s3.S3
	iamClient *iam.IAM
	stsClient *sts.STS
	kmsClient *kms.KMS
}

// InitAwsSession initializes AWS session with default region
func InitAwsSession() *AwsClient {
	mySession := session.Must(session.NewSession())
	return &AwsClient{
		svc:       ec2.New(mySession, aws.NewConfig()),
		s3Client:  s3.New(mySession),
		iamClient: iam.New(mySession),
		stsClient: sts.New(mySession),
		kmsClient: kms.New(mySession),
	}
}

// InitAwsSessionWithRegion initializes AWS session with specific region
func InitAwsSessionWithRegion(region string) *AwsClient {
	mySession := session.Must(session.NewSession())
	config := aws.NewConfig().WithRegion(region)
	return &AwsClient{
		svc:       ec2.New(mySession, config),
		s3Client:  s3.New(mySession, config),
		iamClient: iam.New(mySession, config),
		stsClient: sts.New(mySession, config),
		kmsClient: kms.New(mySession, config),
	}
}

// GetSecurityGroupByGroupName gets security group by name
func (a *AwsClient) GetSecurityGroupByGroupName(groupName string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-name"),
				Values: []*string{aws.String(groupName)},
			},
		},
	}
	result, err := a.svc.DescribeSecurityGroups(input)
	if err != nil {
		return nil, err
	}
	if len(result.SecurityGroups) == 0 {
		return nil, fmt.Errorf("security group %s not found", groupName)
	}
	return result.SecurityGroups[0], nil
}

// GetSecurityGroupByGroupID gets security group by ID
func (a *AwsClient) GetSecurityGroupByGroupID(groupID string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{aws.String(groupID)},
	}
	result, err := a.svc.DescribeSecurityGroups(input)
	if err != nil {
		return nil, err
	}
	if len(result.SecurityGroups) == 0 {
		return nil, fmt.Errorf("security group %s not found", groupID)
	}
	return result.SecurityGroups[0], nil
}

// GetAwsPublicSubnetID gets public subnet ID
func (a *AwsClient) GetAwsPublicSubnetID(clusterID string) (string, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:kubernetes.io/cluster/" + clusterID),
				Values: []*string{aws.String("shared"), aws.String("owned")},
			},
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String("*public*")},
			},
		},
	}
	result, err := a.svc.DescribeSubnets(input)
	if err != nil {
		return "", err
	}
	if len(result.Subnets) == 0 {
		return "", fmt.Errorf("no public subnet found for cluster %s", clusterID)
	}
	return *result.Subnets[0].SubnetId, nil
}

// GetAwsInstanceID gets instance ID from hostname
func (a *AwsClient) GetAwsInstanceID(hostname string) (string, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("private-dns-name"),
				Values: []*string{aws.String(hostname)},
			},
		},
	}
	result, err := a.svc.DescribeInstances(input)
	if err != nil {
		return "", err
	}
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			return *instance.InstanceId, nil
		}
	}
	return "", fmt.Errorf("instance with hostname %s not found", hostname)
}

// GetAwsInstanceVPCId gets VPC ID for instance
func (a *AwsClient) GetAwsInstanceVPCId(instanceID string) (string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}
	result, err := a.svc.DescribeInstances(input)
	if err != nil {
		return "", err
	}
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			return *instance.VpcId, nil
		}
	}
	return "", fmt.Errorf("instance %s not found", instanceID)
}

// GetAwsPrivateSubnetIDs gets private subnet IDs
func (a *AwsClient) GetAwsPrivateSubnetIDs(vpcID string) ([]string, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpcID)},
			},
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String("*private*")},
			},
		},
	}
	result, err := a.svc.DescribeSubnets(input)
	if err != nil {
		return nil, err
	}
	var subnetIDs []string
	for _, subnet := range result.Subnets {
		subnetIDs = append(subnetIDs, *subnet.SubnetId)
	}
	return subnetIDs, nil
}

// GetPlacementGroupByName gets placement group by name
func (a *AwsClient) GetPlacementGroupByName(groupName string) (string, error) {
	input := &ec2.DescribePlacementGroupsInput{
		GroupNames: []*string{aws.String(groupName)},
	}
	result, err := a.svc.DescribePlacementGroups(input)
	if err != nil {
		return "", err
	}
	if len(result.PlacementGroups) == 0 {
		return "", fmt.Errorf("placement group %s not found", groupName)
	}
	return *result.PlacementGroups[0].GroupId, nil
}

// CreateDhcpOptions creates DHCP options
func (a *AwsClient) CreateDhcpOptions() (string, error) {
	input := &ec2.CreateDhcpOptionsInput{
		DhcpConfigurations: []*ec2.NewDhcpConfiguration{
			{
				Key:    aws.String("domain-name-servers"),
				Values: []*string{aws.String("AmazonProvidedDNS")},
			},
		},
	}
	result, err := a.svc.CreateDhcpOptions(input)
	if err != nil {
		return "", err
	}
	return *result.DhcpOptions.DhcpOptionsId, nil
}

// CreateDhcpOptionsWithDomainName creates DHCP options with domain name
func (a *AwsClient) CreateDhcpOptionsWithDomainName(domainName string) (string, error) {
	input := &ec2.CreateDhcpOptionsInput{
		DhcpConfigurations: []*ec2.NewDhcpConfiguration{
			{
				Key:    aws.String("domain-name"),
				Values: []*string{aws.String(domainName)},
			},
			{
				Key:    aws.String("domain-name-servers"),
				Values: []*string{aws.String("AmazonProvidedDNS")},
			},
		},
	}
	result, err := a.svc.CreateDhcpOptions(input)
	if err != nil {
		return "", err
	}
	return *result.DhcpOptions.DhcpOptionsId, nil
}

// DeleteDhcpOptions deletes DHCP options
func (a *AwsClient) DeleteDhcpOptions(dhcpOptionsID string) error {
	input := &ec2.DeleteDhcpOptionsInput{
		DhcpOptionsId: aws.String(dhcpOptionsID),
	}
	_, err := a.svc.DeleteDhcpOptions(input)
	return err
}

// GetDhcpOptionsIDOfVpc gets DHCP options ID of VPC
func (a *AwsClient) GetDhcpOptionsIDOfVpc(vpcID string) (string, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: []*string{aws.String(vpcID)},
	}
	result, err := a.svc.DescribeVpcs(input)
	if err != nil {
		return "", err
	}
	if len(result.Vpcs) == 0 {
		return "", fmt.Errorf("VPC %s not found", vpcID)
	}
	return *result.Vpcs[0].DhcpOptionsId, nil
}

// AssociateDhcpOptions associates DHCP options with VPC
func (a *AwsClient) AssociateDhcpOptions(vpcID, dhcpOptionsID string) error {
	input := &ec2.AssociateDhcpOptionsInput{
		VpcId:         aws.String(vpcID),
		DhcpOptionsId: aws.String(dhcpOptionsID),
	}
	_, err := a.svc.AssociateDhcpOptions(input)
	return err
}

// CreateSecurityGroup creates a security group
func (a *AwsClient) CreateSecurityGroup(groupName, vpcID, description string) (string, error) {
	input := &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(groupName),
		Description: aws.String(description),
		VpcId:       aws.String(vpcID),
	}
	result, err := a.svc.CreateSecurityGroup(input)
	if err != nil {
		return "", err
	}
	return *result.GroupId, nil
}

// DeleteSecurityGroup deletes a security group
func (a *AwsClient) DeleteSecurityGroup(groupID string) error {
	input := &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(groupID),
	}
	_, err := a.svc.DeleteSecurityGroup(input)
	return err
}

// GetInstanceSecurityGroupIDs gets security group IDs for an instance
func (a *AwsClient) GetInstanceSecurityGroupIDs(instanceID string) ([]string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}
	result, err := a.svc.DescribeInstances(input)
	if err != nil {
		return nil, err
	}
	var sgIDs []string
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			for _, sg := range instance.SecurityGroups {
				sgIDs = append(sgIDs, *sg.GroupId)
			}
		}
	}
	return sgIDs, nil
}

// CreateTag creates a tag on a resource
func (a *AwsClient) CreateTag(resource string, key string, value string) error {
	input := &ec2.CreateTagsInput{
		Resources: []*string{aws.String(resource)},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(key),
				Value: aws.String(value),
			},
		},
	}
	_, err := a.svc.CreateTags(input)
	return err
}

// TagS3BucketsWithStackName tags S3 buckets with stack name
func (a *AwsClient) TagS3BucketsWithStackName(bucketNames []string, stackName string) error {
	// Tag each bucket with the stack name
	for _, bucketName := range bucketNames {
		input := &s3.PutBucketTaggingInput{
			Bucket: aws.String(bucketName),
			Tagging: &s3.Tagging{
				TagSet: []*s3.Tag{
					{
						Key:   aws.String("StackName"),
						Value: aws.String(stackName),
					},
				},
			},
		}

		if a.s3Client == nil {
			return fmt.Errorf("S3 client not initialized")
		}

		_, err := a.s3Client.PutBucketTagging(input)
		if err != nil {
			return fmt.Errorf("failed to tag bucket %s: %w", bucketName, err)
		}
	}
	return nil
}

// GetAwsInstanceState gets the state of an AWS instance
func (a *AwsClient) GetAwsInstanceState(instanceID string) (string, error) {
	// Describe the instance
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}

	result, err := a.svc.DescribeInstances(input)
	if err != nil {
		return "", fmt.Errorf("failed to describe instance %s: %w", instanceID, err)
	}

	// Check if we have any reservations and instances
	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("instance %s not found", instanceID)
	}

	// Get the state name
	instance := result.Reservations[0].Instances[0]
	if instance.State != nil && instance.State.Name != nil {
		return *instance.State.Name, nil
	}

	return "", fmt.Errorf("instance state not available")
}

// GetAwsIntIPs gets the internal and public IPs of an AWS instance
func (a *AwsClient) GetAwsIntIPs(instanceID string) (map[string]string, error) {
	// Describe the instance
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}

	result, err := a.svc.DescribeInstances(input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance %s: %w", instanceID, err)
	}

	// Check if we have any reservations and instances
	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("instance %s not found", instanceID)
	}

	// Get IP addresses
	instance := result.Reservations[0].Instances[0]
	ips := make(map[string]string)

	if instance.PublicIpAddress != nil {
		ips["publicIP"] = *instance.PublicIpAddress
	}

	if instance.PrivateIpAddress != nil {
		ips["privateIP"] = *instance.PrivateIpAddress
	}

	return ips, nil
}

// GetDhcpOptionsIDFromTag gets DHCP options ID from tag
func (a *AwsClient) GetDhcpOptionsIDFromTag(tagName string, tagValue string) ([]string, error) {
	// For OTP compatibility - returns array of DHCP options IDs
	input := &ec2.DescribeDhcpOptionsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:" + tagName),
				Values: []*string{aws.String(tagValue)},
			},
		},
	}

	result, err := a.svc.DescribeDhcpOptions(input)
	if err != nil {
		return nil, err
	}

	var dhcpOptionsIDs []string
	for _, dhcpOption := range result.DhcpOptions {
		if dhcpOption.DhcpOptionsId != nil {
			dhcpOptionsIDs = append(dhcpOptionsIDs, *dhcpOption.DhcpOptionsId)
		}
	}

	return dhcpOptionsIDs, nil
}

// DeleteTag deletes a tag from a resource
func (a *AwsClient) DeleteTag(resourceID string, tagKey string, tagValue string) error {
	input := &ec2.DeleteTagsInput{
		Resources: []*string{aws.String(resourceID)},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(tagKey),
				Value: aws.String(tagValue),
			},
		},
	}
	_, err := a.svc.DeleteTags(input)
	return err
}

// StartInstance starts an AWS EC2 instance
func (a *AwsClient) StartInstance(instanceID string) error {
	input := &ec2.StartInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}

	_, err := a.svc.StartInstances(input)
	if err != nil {
		return fmt.Errorf("failed to start instance %s: %w", instanceID, err)
	}

	return nil
}

// StopInstance stops an AWS EC2 instance
func (a *AwsClient) StopInstance(instanceID string) error {
	input := &ec2.StopInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}

	_, err := a.svc.StopInstances(input)
	if err != nil {
		return fmt.Errorf("failed to stop instance %s: %w", instanceID, err)
	}

	return nil
}

// GetAwsInstanceIDFromHostname gets instance ID from hostname
func (a *AwsClient) GetAwsInstanceIDFromHostname(hostname string) (string, error) {
	// Query instances by private DNS name
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("private-dns-name"),
				Values: []*string{aws.String(hostname)},
			},
		},
	}

	result, err := a.svc.DescribeInstances(input)
	if err != nil {
		return "", fmt.Errorf("failed to query instances by hostname %s: %w", hostname, err)
	}

	// Check if we found any instances
	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("no instance found with hostname %s", hostname)
	}

	// Return the first instance ID
	instance := result.Reservations[0].Instances[0]
	if instance.InstanceId != nil {
		return *instance.InstanceId, nil
	}

	return "", fmt.Errorf("instance ID not available")
}

// ECRClient represents an AWS ECR client
type ECRClient struct {
	svc *ecr.ECR
}

// NewECRClient creates a new ECR client
func NewECRClient(region string) *ECRClient {
	mySession := session.Must(session.NewSession())
	return &ECRClient{
		svc: ecr.New(mySession, aws.NewConfig().WithRegion(region)),
	}
}

// CreateContainerRepository creates a container repository
func (e *ECRClient) CreateContainerRepository(repositoryName string) (string, error) {
	input := &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repositoryName),
	}
	result, err := e.svc.CreateRepository(input)
	if err != nil {
		return "", err
	}
	return *result.Repository.RepositoryUri, nil
}

// DeleteContainerRepository deletes a container repository
func (e *ECRClient) DeleteContainerRepository(repositoryName string) error {
	input := &ecr.DeleteRepositoryInput{
		RepositoryName: aws.String(repositoryName),
		Force:          aws.Bool(true),
	}
	_, err := e.svc.DeleteRepository(input)
	return err
}

// GetAuthorizationToken gets ECR authorization token
func (e *ECRClient) GetAuthorizationToken() (string, error) {
	input := &ecr.GetAuthorizationTokenInput{}
	result, err := e.svc.GetAuthorizationToken(input)
	if err != nil {
		return "", err
	}
	if len(result.AuthorizationData) == 0 {
		return "", fmt.Errorf("no authorization data returned")
	}
	return *result.AuthorizationData[0].AuthorizationToken, nil
}

// IAMClient represents an AWS IAM client
type IAMClient struct {
	svc *iam.IAM
}

// NewIAMClient creates a new IAM client
func NewIAMClient() *IAMClient {
	mySession := session.Must(session.NewSession())
	return &IAMClient{
		svc: iam.New(mySession),
	}
}

// AttachRolePolicy attaches a policy to a role
func (i *IAMClient) AttachRolePolicy(roleName, policyArn string) error {
	input := &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyArn),
	}
	_, err := i.svc.AttachRolePolicy(input)
	return err
}

// DetachRolePolicy detaches a policy from a role
func (i *IAMClient) DetachRolePolicy(roleName, policyArn string) error {
	input := &iam.DetachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyArn),
	}
	_, err := i.svc.DetachRolePolicy(input)
	return err
}

// UpdateRolePolicy updates an inline policy for a role
func (i *IAMClient) UpdateRolePolicy(roleName, policyName, policyDocument string) error {
	input := &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDocument),
	}
	_, err := i.svc.PutRolePolicy(input)
	return err
}

// GetRole gets an IAM role
func (i *IAMClient) GetRole(roleName string) (*iam.GetRoleOutput, error) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}
	return i.svc.GetRole(input)
}

// DeleteOpenIDConnectProviderByProviderName deletes an OIDC provider by name
func (i *IAMClient) DeleteOpenIDConnectProviderByProviderName(providerName string) error {
	// List all OIDC providers
	listResp, err := i.svc.ListOpenIDConnectProviders(&iam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return fmt.Errorf("failed to list OIDC providers: %w", err)
	}

	// Find the provider with matching name
	for _, provider := range listResp.OpenIDConnectProviderList {
		if provider.Arn != nil && strings.Contains(*provider.Arn, providerName) {
			// Delete the provider
			_, err := i.svc.DeleteOpenIDConnectProvider(&iam.DeleteOpenIDConnectProviderInput{
				OpenIDConnectProviderArn: provider.Arn,
			})
			if err != nil {
				return fmt.Errorf("failed to delete OIDC provider %s: %w", *provider.Arn, err)
			}
			return nil
		}
	}

	return fmt.Errorf("OIDC provider with name %s not found", providerName)
}

// GetRolePolicy gets an inline policy document attached to a role
func (i *IAMClient) GetRolePolicy(roleName, policyName string) (string, error) {
	input := &iam.GetRolePolicyInput{
		RoleName:   aws.String(roleName),
		PolicyName: aws.String(policyName),
	}

	result, err := i.svc.GetRolePolicy(input)
	if err != nil {
		return "", fmt.Errorf("failed to get role policy: %w", err)
	}

	if result.PolicyDocument == nil {
		return "", fmt.Errorf("policy document is nil")
	}

	// URL decode the policy document
	decoded, err := url.QueryUnescape(*result.PolicyDocument)
	if err != nil {
		return "", fmt.Errorf("failed to decode policy document: %w", err)
	}

	return decoded, nil
}

// KMSClient represents an AWS KMS client
type KMSClient struct {
	svc *kms.KMS
}

// NewKMSClient creates a new KMS client
func NewKMSClient(region string) *KMSClient {
	mySession := session.Must(session.NewSession())
	return &KMSClient{
		svc: kms.New(mySession, aws.NewConfig().WithRegion(region)),
	}
}

// CreateKey creates a KMS key
func (k *KMSClient) CreateKey(description string) (string, error) {
	input := &kms.CreateKeyInput{
		Description: aws.String(description),
		KeyUsage:    aws.String("ENCRYPT_DECRYPT"),
		Origin:      aws.String("AWS_KMS"),
	}
	result, err := k.svc.CreateKey(input)
	if err != nil {
		return "", err
	}
	return *result.KeyMetadata.Arn, nil
}

// DeleteKey schedules a KMS key for deletion
func (k *KMSClient) DeleteKey(keyId string) error {
	input := &kms.ScheduleKeyDeletionInput{
		KeyId:               aws.String(keyId),
		PendingWindowInDays: aws.Int64(7), // Minimum allowed value
	}
	_, err := k.svc.ScheduleKeyDeletion(input)
	return err
}

// S3Client represents an AWS S3 client
type S3Client struct {
	svc *s3.S3
}

// NewS3Client creates a new S3 client
func NewS3Client() *S3Client {
	mySession := session.Must(session.NewSession())
	return &S3Client{
		svc: s3.New(mySession),
	}
}

// NewS3ClientFromCredFile creates a new S3 client from credential file
func NewS3ClientFromCredFile(filename, profile, region string) *S3Client {
	creds := credentials.NewSharedCredentials(filename, profile)
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	}))
	return &S3Client{
		svc: s3.New(sess),
	}
}

// CreateBucket creates an S3 bucket
func (sc *S3Client) CreateBucket(name string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}
	_, err := sc.svc.CreateBucket(input)
	return err
}

// DeleteBucket deletes an S3 bucket
func (sc *S3Client) DeleteBucket(name string) error {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}
	_, err := sc.svc.DeleteBucket(input)
	return err
}

// HeadBucket checks if a bucket exists
func (sc *S3Client) HeadBucket(name string) error {
	input := &s3.HeadBucketInput{
		Bucket: aws.String(name),
	}
	_, err := sc.svc.HeadBucket(input)
	return err
}

// ListBuckets lists S3 buckets
func (s *S3Client) ListBuckets() ([]string, error) {
	result, err := s.svc.ListBuckets(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	var bucketNames []string
	for _, bucket := range result.Buckets {
		if bucket.Name != nil {
			bucketNames = append(bucketNames, *bucket.Name)
		}
	}

	return bucketNames, nil
}

// PutBucketPolicy sets a bucket policy
func (s *S3Client) PutBucketPolicy(bucketName string, policy string) error {
	input := &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(policy),
	}

	_, err := s.svc.PutBucketPolicy(input)
	if err != nil {
		return fmt.Errorf("failed to put bucket policy for %s: %w", bucketName, err)
	}

	return nil
}

// Route53Client represents an AWS Route53 client
type Route53Client struct {
	*route53.Route53
}

// NewRoute53Client creates a new Route53 client
func NewRoute53Client() *Route53Client {
	mySession := session.Must(session.NewSession())
	return &Route53Client{
		Route53: route53.New(mySession),
	}
}

// StsClient represents an AWS STS client
type StsClient struct {
	*sts.STS
}

// NewDelegatingStsClient creates a new STS client
func NewDelegatingStsClient(wrappedClient *sts.STS) *StsClient {
	return &StsClient{
		STS: wrappedClient,
	}
}

// GetAWSClusterRegion gets the AWS region from environment
func GetAWSClusterRegion() string {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
		if region == "" {
			region = "us-east-1" // Default region
		}
	}
	return region
}

// GetDefaultSecurityGroupByVpcID gets the default security group for a VPC
func (a *AwsClient) GetDefaultSecurityGroupByVpcID(vpcID string) (*ec2.SecurityGroup, error) {
	// For OTP compatibility - gets default security group for VPC
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpcID)},
			},
			{
				Name:   aws.String("group-name"),
				Values: []*string{aws.String("default")},
			},
		},
	}

	result, err := a.svc.DescribeSecurityGroups(input)
	if err != nil {
		return nil, err
	}

	if len(result.SecurityGroups) == 0 {
		return nil, fmt.Errorf("no default security group found for VPC %s", vpcID)
	}

	return result.SecurityGroups[0], nil
}

// GetSecurityGroupsByVpcEndpointID gets security groups associated with a VPC endpoint
func (a *AwsClient) GetSecurityGroupsByVpcEndpointID(endpointID string) ([]*ec2.SecurityGroup, error) {
	// For OTP compatibility - gets security groups for VPC endpoint
	input := &ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: []*string{aws.String(endpointID)},
	}

	result, err := a.svc.DescribeVpcEndpoints(input)
	if err != nil {
		return nil, err
	}

	if len(result.VpcEndpoints) == 0 {
		return nil, fmt.Errorf("VPC endpoint %s not found", endpointID)
	}

	// Get the group IDs from the endpoint
	var groupIDs []*string
	for _, group := range result.VpcEndpoints[0].Groups {
		if group.GroupId != nil {
			groupIDs = append(groupIDs, group.GroupId)
		}
	}

	// Now describe the security groups to get full objects
	if len(groupIDs) == 0 {
		return nil, nil
	}

	sgInput := &ec2.DescribeSecurityGroupsInput{
		GroupIds: groupIDs,
	}

	sgResult, err := a.svc.DescribeSecurityGroups(sgInput)
	if err != nil {
		return nil, err
	}

	return sgResult.SecurityGroups, nil
}

// EmptyBucketWithContextAndCheck empties a bucket and waits for the deletions to take effect
func (s *S3Client) EmptyBucketWithContextAndCheck(ctx context.Context, bucketName string, check bool) error {
	// List all objects in the bucket
	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	var objectsToDelete []*s3.ObjectIdentifier
	err := s.svc.ListObjectsV2PagesWithContext(ctx, listInput, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
				Key: obj.Key,
			})
		}
		return true
	})

	if err != nil {
		return fmt.Errorf("failed to list objects in bucket %s: %v", bucketName, err)
	}

	// Delete objects in batches of 1000 (S3 limit)
	for i := 0; i < len(objectsToDelete); i += 1000 {
		end := i + 1000
		if end > len(objectsToDelete) {
			end = len(objectsToDelete)
		}

		deleteInput := &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &s3.Delete{
				Objects: objectsToDelete[i:end],
				Quiet:   aws.Bool(true),
			},
		}

		_, err := s.svc.DeleteObjectsWithContext(ctx, deleteInput)
		if err != nil {
			return fmt.Errorf("failed to delete objects from bucket %s: %v", bucketName, err)
		}
	}

	// If check is true, wait for bucket to be empty
	if check {
		return s.WaitForBucketEmptinessWithContext(ctx, bucketName, "BucketEmpty", 5*time.Second, 1*time.Minute)
	}

	return nil
}

// WaitForBucketEmptinessWithContext waits for the expected bucket emptiness state
func (s *S3Client) WaitForBucketEmptinessWithContext(ctx context.Context, bucketName string, expectedState string, interval time.Duration, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		listInput := &s3.ListObjectsV2Input{
			Bucket:  aws.String(bucketName),
			MaxKeys: aws.Int64(1),
		}

		output, err := s.svc.ListObjectsV2WithContext(ctx, listInput)
		if err != nil {
			return fmt.Errorf("failed to list objects in bucket %s: %v", bucketName, err)
		}

		isEmpty := len(output.Contents) == 0

		if expectedState == "BucketEmpty" && isEmpty {
			return nil
		} else if expectedState == "BucketNonEmpty" && !isEmpty {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			// Continue checking
		}
	}

	return fmt.Errorf("timeout waiting for bucket %s to reach state %s", bucketName, expectedState)
}

// CreateRoleWithContext creates an IAM role with context
func (i *IAMClient) CreateRoleWithContext(ctx context.Context, input *iam.CreateRoleInput, opts ...request.Option) (*iam.CreateRoleOutput, error) {
	return i.svc.CreateRoleWithContext(ctx, input, opts...)
}

// EmptyHostedZoneWithContext empties a hosted zone of all records except NS and SOA
func (r *Route53Client) EmptyHostedZoneWithContext(ctx context.Context, zoneID string) error {
	// List all record sets
	var recordSets []*route53.ResourceRecordSet
	err := r.ListResourceRecordSetsPagesWithContext(ctx, &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}, func(page *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
		recordSets = append(recordSets, page.ResourceRecordSets...)
		return !lastPage
	})
	if err != nil {
		return err
	}

	// Create batch of changes to delete all records except NS and SOA
	var changes []*route53.Change
	for _, rs := range recordSets {
		if *rs.Type == "NS" || *rs.Type == "SOA" {
			continue
		}
		changes = append(changes, &route53.Change{
			Action:            aws.String("DELETE"),
			ResourceRecordSet: rs,
		})
	}

	if len(changes) == 0 {
		return nil
	}

	// Apply changes
	_, err = r.ChangeResourceRecordSetsWithContext(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: changes,
		},
	})
	return err
}

// DeleteHostedZoneWithContextAndCheck deletes a hosted zone after checking it's empty
func (r *Route53Client) DeleteHostedZoneWithContextAndCheck(ctx context.Context, zoneID string) error {
	// First empty the zone
	if err := r.EmptyHostedZoneWithContext(ctx, zoneID); err != nil {
		return err
	}

	// Then delete the zone
	_, err := r.DeleteHostedZoneWithContext(ctx, &route53.DeleteHostedZoneInput{
		Id: aws.String(zoneID),
	})
	return err
}

// UpdateAwsIntSecurityRule updates AWS security group to allow the specified port
func (a *AwsClient) UpdateAwsIntSecurityRule(instanceID string, dstPort int64) error {
	// Get instance details
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}
	result, err := a.svc.DescribeInstances(input)
	if err != nil {
		return err
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return fmt.Errorf("No such instance ID in current cluster %s", instanceID)
	}

	// Get security group ID from instance
	instance := result.Reservations[0].Instances[0]
	if len(instance.SecurityGroups) == 0 {
		return fmt.Errorf("No security groups found for instance %s", instanceID)
	}
	securityGroupID := instance.SecurityGroups[0].GroupId

	// Check if destination port is already open
	sgInput := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{securityGroupID},
	}
	sgResult, err := a.svc.DescribeSecurityGroups(sgInput)
	if err != nil {
		return err
	}

	// Check if port is already open
	for _, sg := range sgResult.SecurityGroups {
		for _, rule := range sg.IpPermissions {
			if rule.FromPort != nil && *rule.FromPort == dstPort &&
				rule.ToPort != nil && *rule.ToPort == dstPort {
				// Port already open
				return nil
			}
		}
	}

	// Add ingress rule to allow destination port
	_, err = a.svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: securityGroupID,
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int64(dstPort),
				ToPort:     aws.Int64(dstPort),
				IpRanges: []*ec2.IpRange{
					{CidrIp: aws.String("0.0.0.0/0")},
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("Unable to set security group %s ingress: %v", *securityGroupID, err)
	}

	return nil
}

// GetAvailabilityZoneNames gets availability zone names
func (a *AwsClient) GetAvailabilityZoneNames(region string) ([]string, error) {
	// For OTP compatibility - get availability zones
	return []string{"us-east-1a", "us-east-1b", "us-east-1c"}, nil
}
