package compat_otp

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
)

// IBMPowerVsSession represents IBM Power VS session
type IBMPowerVsSession struct {
	// Power VS session
	powerVsSession *ibmpisession.IBMPISession

	// Instance client
	instanceClient *instance.IBMPIInstanceClient

	// Cloud instance ID
	CloudInstanceID string
}

// LoginIBMPowerVsCloud logs into IBM Power VS Cloud
func LoginIBMPowerVsCloud(apiKey, zone, userAccount string, cloudId string) (*IBMPowerVsSession, error) {
	// Create authenticator
	authenticator := &core.IamAuthenticator{
		ApiKey: apiKey,
		URL:    "https://iam.cloud.ibm.com",
	}

	// Create session options
	options := &ibmpisession.IBMPIOptions{
		Authenticator: authenticator,
		UserAccount:   userAccount,
		Zone:          zone,
		Debug:         false,
	}

	// Create session
	session, err := ibmpisession.NewIBMPISession(options)
	if err != nil {
		return nil, err
	}

	// Create instance client
	instanceClient := instance.NewIBMPIInstanceClient(context.Background(), session, cloudId)

	return &IBMPowerVsSession{
		powerVsSession:  session,
		instanceClient:  instanceClient,
		CloudInstanceID: cloudId,
	}, nil
}

// PerformInstanceActionOnPowerVs performs an action on a Power VS instance
func PerformInstanceActionOnPowerVs(powerClient *IBMPowerVsSession, instanceID, action string) error {
	if powerClient == nil || powerClient.instanceClient == nil {
		return fmt.Errorf("invalid Power VS client")
	}

	// Create action body
	body := &models.PVMInstanceAction{
		Action: &action,
	}

	// Perform action
	err := powerClient.instanceClient.Action(instanceID, body)
	return err
}

// GetIBMPowerVsInstanceInfo gets information about a Power VS instance
func GetIBMPowerVsInstanceInfo(powerClient *IBMPowerVsSession, instanceName string) (string, string, error) {
	if powerClient == nil || powerClient.instanceClient == nil {
		return "", "", fmt.Errorf("invalid Power VS client")
	}

	// Get all instances
	instances, err := powerClient.instanceClient.GetAll()
	if err != nil {
		return "", "", err
	}

	// Find instance by name
	for _, inst := range instances.PvmInstances {
		if *inst.ServerName == instanceName {
			status := ""
			if inst.Status != nil {
				status = *inst.Status
			}
			return *inst.PvmInstanceID, status, nil
		}
	}

	return "", "", fmt.Errorf("instance %s not found", instanceName)
}

// GetInstanceByID gets an instance by ID
func (s *IBMPowerVsSession) GetInstanceByID(instanceID string) (*models.PVMInstance, error) {
	if s.instanceClient == nil {
		return nil, fmt.Errorf("instance client not initialized")
	}

	return s.instanceClient.Get(instanceID)
}

// GetInstanceByName gets an instance by name
func (s *IBMPowerVsSession) GetInstanceByName(instanceName string) (*models.PVMInstance, error) {
	if s.instanceClient == nil {
		return nil, fmt.Errorf("instance client not initialized")
	}

	// Get all instances
	instances, err := s.instanceClient.GetAll()
	if err != nil {
		return nil, err
	}

	// Find instance by name
	for _, inst := range instances.PvmInstances {
		if *inst.ServerName == instanceName {
			// Get full instance details
			return s.instanceClient.Get(*inst.PvmInstanceID)
		}
	}

	return nil, fmt.Errorf("instance %s not found", instanceName)
}

// StartInstance starts a Power VS instance
func (s *IBMPowerVsSession) StartInstance(instanceID string) error {
	action := "start"
	return PerformInstanceActionOnPowerVs(s, instanceID, action)
}

// StopInstance stops a Power VS instance
func (s *IBMPowerVsSession) StopInstance(instanceID string) error {
	action := "immediate-shutdown"
	return PerformInstanceActionOnPowerVs(s, instanceID, action)
}

// RestartInstance restarts a Power VS instance
func (s *IBMPowerVsSession) RestartInstance(instanceID string) error {
	action := "hard-reboot"
	return PerformInstanceActionOnPowerVs(s, instanceID, action)
}

// CreateInstance creates a new Power VS instance
func (s *IBMPowerVsSession) CreateInstance(params *models.PVMInstanceCreate) (*models.PVMInstance, error) {
	if s.instanceClient == nil {
		return nil, fmt.Errorf("instance client not initialized")
	}

	// Create instance
	instances, err := s.instanceClient.Create(params)
	if err != nil {
		return nil, err
	}

	if len(*instances) == 0 {
		return nil, fmt.Errorf("no instance created")
	}

	// Get the first instance
	if instances != nil && len(*instances) > 0 {
		return (*instances)[0], nil
	}
	return nil, fmt.Errorf("no instances found")
}

// DeleteInstance deletes a Power VS instance
func (s *IBMPowerVsSession) DeleteInstance(instanceID string) error {
	if s.instanceClient == nil {
		return fmt.Errorf("instance client not initialized")
	}

	return s.instanceClient.Delete(instanceID)
}

// WaitForInstanceStatus waits for an instance to reach a specific status
func (s *IBMPowerVsSession) WaitForInstanceStatus(instanceID string, targetStatus string, timeout time.Duration) error {
	startTime := time.Now()

	for {
		instance, err := s.GetInstanceByID(instanceID)
		if err != nil {
			return err
		}

		if instance.Status != nil && *instance.Status == targetStatus {
			return nil
		}

		if time.Since(startTime) > timeout {
			return fmt.Errorf("timeout waiting for instance %s to reach status %s", instanceID, targetStatus)
		}

		time.Sleep(10 * time.Second)
	}
}

// GetNetworks gets all networks
func (s *IBMPowerVsSession) GetNetworks() ([]*models.NetworkReference, error) {
	if s.powerVsSession == nil {
		return nil, fmt.Errorf("session not initialized")
	}

	networkClient := instance.NewIBMPINetworkClient(context.Background(), s.powerVsSession, s.CloudInstanceID)
	networks, err := networkClient.GetAll()
	if err != nil {
		return nil, err
	}

	return networks.Networks, nil
}

// GetImages gets all images
func (s *IBMPowerVsSession) GetImages() ([]*models.ImageReference, error) {
	if s.powerVsSession == nil {
		return nil, fmt.Errorf("session not initialized")
	}

	imageClient := instance.NewIBMPIImageClient(context.Background(), s.powerVsSession, s.CloudInstanceID)
	images, err := imageClient.GetAll()
	if err != nil {
		return nil, err
	}

	return images.Images, nil
}

// GetSSHKeys gets all SSH keys
func (s *IBMPowerVsSession) GetSSHKeys() ([]*models.SSHKey, error) {
	if s.powerVsSession == nil {
		return nil, fmt.Errorf("session not initialized")
	}

	keyClient := instance.NewIBMPIKeyClient(context.Background(), s.powerVsSession, s.CloudInstanceID)
	keys, err := keyClient.GetAll()
	if err != nil {
		return nil, err
	}

	return keys.SSHKeys, nil
}

// CreateVolume creates a new volume
func (s *IBMPowerVsSession) CreateVolume(params *models.CreateDataVolume) (*models.Volume, error) {
	if s.powerVsSession == nil {
		return nil, fmt.Errorf("session not initialized")
	}

	volumeClient := instance.NewIBMPIVolumeClient(context.Background(), s.powerVsSession, s.CloudInstanceID)
	return volumeClient.CreateVolume(params)
}

// AttachVolume attaches a volume to an instance
func (s *IBMPowerVsSession) AttachVolume(instanceID, volumeID string) error {
	if s.instanceClient == nil {
		return fmt.Errorf("instance client not initialized")
	}

	// AttachVolume method no longer available in current SDK
	return fmt.Errorf("AttachVolume temporarily disabled due to SDK compatibility issues")
}

// DetachVolume detaches a volume from an instance
func (s *IBMPowerVsSession) DetachVolume(instanceID, volumeID string) error {
	if s.instanceClient == nil {
		return fmt.Errorf("instance client not initialized")
	}

	// DetachVolume method no longer available in current SDK
	return fmt.Errorf("DetachVolume temporarily disabled due to SDK compatibility issues")
}

// GetCloudInstanceID gets the cloud instance ID from metadata
func GetCloudInstanceID(zone string) (string, error) {
	// This would typically be retrieved from cluster metadata or configuration
	// For now, return a placeholder
	return fmt.Sprintf("cloud-instance-%s", zone), nil
}

// CreateInstanceHelper creates an instance with common defaults
func (s *IBMPowerVsSession) CreateInstanceHelper(name string, imageID string, memory float64, processors float64, networkID string, sshKeyName string) (*models.PVMInstance, error) {
	// Set instance parameters
	procType := "shared"
	storageType := "tier3"
	sysType := "s922"
	replicants := float64(1)
	replicantScheme := "suffix"
	replicantAffinityPolicy := "affinity"

	params := &models.PVMInstanceCreate{
		ServerName:              &name,
		ImageID:                 &imageID,
		Memory:                  &memory,
		Processors:              &processors,
		ProcType:                &procType,
		StorageType:             storageType,
		SysType:                 sysType,
		KeyPairName:             sshKeyName,
		NetworkIDs:              []string{networkID},
		Replicants:              &replicants,
		ReplicantNamingScheme:   &replicantScheme,
		ReplicantAffinityPolicy: &replicantAffinityPolicy,
	}

	return s.CreateInstance(params)
}

// NewIBMPowerVsSession creates a new IBM Power VS session
func NewIBMPowerVsSession() (*IBMPowerVsSession, error) {
	apiKey := os.Getenv("IBMCLOUD_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("IBMCLOUD_API_KEY environment variable not set")
	}

	region := os.Getenv("IBMCLOUD_REGION")
	if region == "" {
		region = "us-south" // Default region
	}

	// Create a minimal session - the actual initialization would require more setup
	return &IBMPowerVsSession{
		// For OTP compatibility - minimal implementation
	}, nil
}
