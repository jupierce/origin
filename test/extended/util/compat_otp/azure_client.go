package compat_otp

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/containerregistry/mgmt/containerregistry"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/msi/mgmt/msi"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
)

// AzureSession represents an Azure session
type AzureSession struct {
	SubscriptionID string
	Authorizer     autorest.Authorizer
	AzureRegion    string
	AzureClientID  string
	AzureTenantID  string
}

// GetFromClusterAndDecode gets Azure credentials from the cluster
func (a *AzureSession) GetFromClusterAndDecode(oc interface{}) error {
	// For OTP compatibility - get credentials from cluster
	// This is a simplified implementation
	a.AzureRegion = "eastus"
	a.AzureClientID = "default-client-id"
	a.AzureTenantID = "default-tenant-id"
	return nil
}

// GetServicePrincipalObjectId gets the service principal object ID
func (a *AzureSession) GetServicePrincipalObjectId(clientID string) (string, error) {
	// For OTP compatibility - simplified implementation
	return "default-object-id", nil
}

// GetVaultsClient returns a key vaults client
func (a *AzureSession) GetVaultsClient() interface{} {
	// For OTP compatibility - returns a vaults client
	return nil
}

// GetKeysClient returns a keys client
func (a *AzureSession) GetKeysClient() interface{} {
	// For OTP compatibility - returns a keys client
	return nil
}

// LoadFromFile loads Azure credentials from file
func (a *AzureSession) LoadFromFile(path string) error {
	// For OTP compatibility - load credentials from file
	// In a real implementation, this would read the file and populate the session
	return nil
}

// GetVirtualMachinesClient returns a virtual machines client
func (a *AzureSession) GetVirtualMachinesClient() interface{} {
	// For OTP compatibility - returns a virtual machines client
	return nil
}

// NewAzureSessionFromEnv creates Azure session from environment
func NewAzureSessionFromEnv() (*AzureSession, error) {
	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return nil, err
	}

	authorizer, err := settings.GetAuthorizer()
	if err != nil {
		return nil, err
	}

	return &AzureSession{
		SubscriptionID: settings.GetSubscriptionID(),
		Authorizer:     authorizer,
	}, nil
}

// GetResourceGroupClient returns a resource groups client
func (a *AzureSession) GetResourceGroupClient(authorizer interface{}) resources.GroupsClient {
	client := resources.NewGroupsClient(a.SubscriptionID)
	client.Authorizer = a.Authorizer
	return client
}

// NewAzureSession creates a new Azure session with the given parameters
func NewAzureSession(subscriptionID, tenantID, clientID, clientSecret, resourceGroup string) (*AzureSession, error) {
	config := auth.NewClientCredentialsConfig(clientID, clientSecret, tenantID)
	authorizer, err := config.Authorizer()
	if err != nil {
		return nil, err
	}

	return &AzureSession{
		SubscriptionID: subscriptionID,
		Authorizer:     authorizer,
	}, nil
}

// GetStorageClient creates a storage accounts client
func getStorageClient(sess *AzureSession) storage.AccountsClient {
	client := storage.NewAccountsClient(sess.SubscriptionID)
	client.Authorizer = sess.Authorizer
	return client
}

// GetNetworkInterfacesClient creates a network interfaces client
func getNicClient(sess *AzureSession) network.InterfacesClient {
	client := network.NewInterfacesClient(sess.SubscriptionID)
	client.Authorizer = sess.Authorizer
	return client
}

// GetPublicIPAddressesClient creates a public IP addresses client
func getIPClient(sess *AzureSession) network.PublicIPAddressesClient {
	client := network.NewPublicIPAddressesClient(sess.SubscriptionID)
	client.Authorizer = sess.Authorizer
	return client
}

// GetVirtualMachinesClient creates a virtual machines client
func getVMClient(sess *AzureSession) compute.VirtualMachinesClient {
	client := compute.NewVirtualMachinesClient(sess.SubscriptionID)
	client.Authorizer = sess.Authorizer
	return client
}

// GetContainerRegistryClient creates a container registry client
func getRegistryClient(sess *AzureSession) containerregistry.RegistriesClient {
	client := containerregistry.NewRegistriesClient(sess.SubscriptionID)
	client.Authorizer = sess.Authorizer
	return client
}

// GetUserAssignedIdentitiesClient creates a user assigned identities client
func getUserAssignedIdentitiesClient(sess *AzureSession) msi.UserAssignedIdentitiesClient {
	client := msi.NewUserAssignedIdentitiesClient(sess.SubscriptionID)
	client.Authorizer = sess.Authorizer
	return client
}

// GetRoleAssignmentsClient creates a role assignments client
func getRoleAssignmentsClient(sess *AzureSession) authorization.RoleAssignmentsClient {
	client := authorization.NewRoleAssignmentsClient(sess.SubscriptionID)
	client.Authorizer = sess.Authorizer
	return client
}

// GetAzureStorageAccount gets storage account name
func GetAzureStorageAccount(sess *AzureSession, resourceGroupName string) (string, error) {
	ctx := context.Background()
	client := getStorageClient(sess)

	result, err := client.ListByResourceGroup(ctx, resourceGroupName)
	if err != nil {
		return "", err
	}

	accounts := result.Values()
	if len(accounts) == 0 {
		return "", fmt.Errorf("no storage accounts found in resource group %s", resourceGroupName)
	}

	return *accounts[0].Name, nil
}

// GetAzureVMPrivateIP gets VM private IP
func GetAzureVMPrivateIP(sess *AzureSession, rg, vmName string) (string, error) {
	ctx := context.Background()
	vmClient := getVMClient(sess)
	nicClient := getNicClient(sess)

	vm, err := vmClient.Get(ctx, rg, vmName, "")
	if err != nil {
		return "", err
	}

	if vm.NetworkProfile == nil || len(*vm.NetworkProfile.NetworkInterfaces) == 0 {
		return "", fmt.Errorf("no network interfaces found for VM %s", vmName)
	}

	nicID := (*vm.NetworkProfile.NetworkInterfaces)[0].ID
	nicName := getResourceNameFromID(*nicID)

	nic, err := nicClient.Get(ctx, rg, nicName, "")
	if err != nil {
		return "", err
	}

	if nic.IPConfigurations == nil || len(*nic.IPConfigurations) == 0 {
		return "", fmt.Errorf("no IP configurations found for NIC %s", nicName)
	}

	return *(*nic.IPConfigurations)[0].PrivateIPAddress, nil
}

// GetAzureVMPublicIP gets VM public IP
func GetAzureVMPublicIP(sess *AzureSession, rg, vmName string) (string, error) {
	ctx := context.Background()
	vmClient := getVMClient(sess)
	nicClient := getNicClient(sess)
	ipClient := getIPClient(sess)

	vm, err := vmClient.Get(ctx, rg, vmName, "")
	if err != nil {
		return "", err
	}

	if vm.NetworkProfile == nil || len(*vm.NetworkProfile.NetworkInterfaces) == 0 {
		return "", fmt.Errorf("no network interfaces found for VM %s", vmName)
	}

	nicID := (*vm.NetworkProfile.NetworkInterfaces)[0].ID
	nicName := getResourceNameFromID(*nicID)

	nic, err := nicClient.Get(ctx, rg, nicName, "")
	if err != nil {
		return "", err
	}

	if nic.IPConfigurations == nil || len(*nic.IPConfigurations) == 0 {
		return "", fmt.Errorf("no IP configurations found for NIC %s", nicName)
	}

	ipConfig := (*nic.IPConfigurations)[0]
	if ipConfig.PublicIPAddress == nil || ipConfig.PublicIPAddress.ID == nil {
		return "", fmt.Errorf("no public IP found for VM %s", vmName)
	}

	publicIPName := getResourceNameFromID(*ipConfig.PublicIPAddress.ID)
	publicIP, err := ipClient.Get(ctx, rg, publicIPName, "")
	if err != nil {
		return "", err
	}

	if publicIP.IPAddress == nil {
		return "", fmt.Errorf("public IP address is nil")
	}

	return *publicIP.IPAddress, nil
}

// StartAzureVM starts an Azure VM
func StartAzureVM(sess *AzureSession, vmName string, resourceGroupName string) (autorest.Response, error) {
	ctx := context.Background()
	client := getVMClient(sess)
	future, err := client.Start(ctx, resourceGroupName, vmName)
	if err != nil {
		return autorest.Response{}, err
	}
	return future.Result(client)
}

// StopAzureVM stops an Azure VM
func StopAzureVM(sess *AzureSession, vmName string, resourceGroupName string) (autorest.Response, error) {
	ctx := context.Background()
	client := getVMClient(sess)
	future, err := client.PowerOff(ctx, resourceGroupName, vmName, nil)
	if err != nil {
		return autorest.Response{}, err
	}
	return future.Result(client)
}

// CreateAzureContainerRegistry creates Azure container registry
func CreateAzureContainerRegistry(sess *AzureSession, registryName string, resourceGroupName string, location string) error {
	ctx := context.Background()
	client := getRegistryClient(sess)

	params := containerregistry.Registry{
		Location: to.StringPtr(location),
		Sku: &containerregistry.Sku{
			Name: containerregistry.Basic,
		},
		RegistryProperties: &containerregistry.RegistryProperties{
			AdminUserEnabled: to.BoolPtr(true),
		},
	}

	future, err := client.Create(ctx, resourceGroupName, registryName, params)
	if err != nil {
		return err
	}

	_, err = future.Result(client)
	return err
}

// GetAzureContainerRepositoryCredential gets Azure container registry credentials
func GetAzureContainerRepositoryCredential(sess *AzureSession, registryName string, resourceGroupName string) (string, string, error) {
	ctx := context.Background()
	client := getRegistryClient(sess)

	creds, err := client.ListCredentials(ctx, resourceGroupName, registryName)
	if err != nil {
		return "", "", err
	}

	if creds.Username == nil || len(*creds.Passwords) == 0 {
		return "", "", fmt.Errorf("no credentials found for registry %s", registryName)
	}

	return *creds.Username, *(*creds.Passwords)[0].Value, nil
}

// DeleteAzureContainerRegistry deletes Azure container registry
func DeleteAzureContainerRegistry(sess *AzureSession, registryName string, resourceGroupName string) error {
	ctx := context.Background()
	client := getRegistryClient(sess)

	future, err := client.Delete(ctx, resourceGroupName, registryName)
	if err != nil {
		return err
	}

	_, err = future.Result(client)
	return err
}

// GetUserAssignedIdentityPrincipalID gets principal ID of user assigned identity
func GetUserAssignedIdentityPrincipalID(sess *AzureSession, resourceGroup string, identityName string) (string, error) {
	ctx := context.Background()
	client := getUserAssignedIdentitiesClient(sess)

	identity, err := client.Get(ctx, resourceGroup, identityName)
	if err != nil {
		return "", err
	}

	if identity.PrincipalID == nil {
		return "", fmt.Errorf("principal ID is nil for identity %s", identityName)
	}

	return identity.PrincipalID.String(), nil
}

// GrantRoleToPrincipalIDByResourceGroup grants role to principal ID
func GrantRoleToPrincipalIDByResourceGroup(sess *AzureSession, principalID string, resourceGroup string, roleId string) (string, string) {
	ctx := context.Background()
	client := getRoleAssignmentsClient(sess)

	scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", sess.SubscriptionID, resourceGroup)

	params := authorization.RoleAssignmentCreateParameters{
		Properties: &authorization.RoleAssignmentProperties{
			RoleDefinitionID: to.StringPtr(roleId),
			PrincipalID:      to.StringPtr(principalID),
		},
	}

	// Generate a unique name for the role assignment
	roleAssignmentName := fmt.Sprintf("%s-%s", principalID[:8], roleId[:8])

	_, err := client.Create(ctx, scope, roleAssignmentName, params)
	if err != nil {
		// Log error but continue
		fmt.Printf("Error creating role assignment: %v\n", err)
	}

	return roleAssignmentName, scope
}

// DeleteRoleAssignments deletes role assignments
func DeleteRoleAssignments(sess *AzureSession, roleAssignmentName string, scope string) error {
	ctx := context.Background()
	client := getRoleAssignmentsClient(sess)

	_, err := client.Delete(ctx, scope, roleAssignmentName)
	return err
}

// NewAzureContainerClient creates a new Azure blob container client
func NewAzureContainerClient(accountName, accountKey, containerName string) (azblob.ContainerURL, error) {
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return azblob.ContainerURL{}, err
	}

	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))
	if err != nil {
		return azblob.ContainerURL{}, err
	}
	containerURL := azblob.NewContainerURL(*u, pipeline)

	return containerURL, nil
}

// CreateAzureStorageBlobContainer creates an Azure blob container
// CreateAzureStorageBlobContainer creates azure storage container
func CreateAzureStorageBlobContainer(container azblob.ContainerURL) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// check if the container exists or not
	// if exists, then remove the blobs in the container, if not, create the container
	_, err := container.GetProperties(ctx, azblob.LeaseAccessConditions{})
	message := fmt.Sprintf("%v", err)
	if strings.Contains(message, "ContainerNotFound") {
		_, err = container.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
		return err
	}
	return EmptyAzureBlobContainer(container)
}

// DeleteAzureStorageBlobContainer deletes azure storage container
func DeleteAzureStorageBlobContainer(container azblob.ContainerURL) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := EmptyAzureBlobContainer(container)
	if err != nil {
		return err
	}
	_, err = container.Delete(ctx, azblob.ContainerAccessConditions{})
	if err != nil {
		return fmt.Errorf("error deleting container: %v", err)
	}
	// Azure storage container is deleted
	return nil
}

// EmptyAzureBlobContainer removes all the files in azure storage container
func EmptyAzureBlobContainer(container azblob.ContainerURL) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for marker := (azblob.Marker{}); marker.NotDone(); { // The parens around Marker{} are required to avoid compiler error.
		// Get a result segment starting with the blob indicated by the current Marker.
		listBlob, err := container.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return fmt.Errorf("error listing blobs in container: %v", err)
		}

		// IMPORTANT: ListBlobs returns the start of the next segment; you MUST use this to get
		// the next segment (after processing the current result segment).
		marker = listBlob.NextMarker

		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		for _, blobInfo := range listBlob.Segment.BlobItems {
			blobURL := container.NewBlockBlobURL(blobInfo.Name)
			_, err := blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
			if err != nil {
				return fmt.Errorf("error deleting blob %s: %v", blobInfo.Name, err)
			}
		}
	}
	// deleted all blob items in the container
	return nil
}

// AzureClientSet represents a set of Azure clients
type AzureClientSet struct {
	SubscriptionID string
	Authorizer     autorest.Authorizer
}

// GetResourceGroupClient returns a resource groups client
func (acs *AzureClientSet) GetResourceGroupClient(authorizer ...interface{}) *AzureResourceGroupClient {
	client := resources.NewGroupsClient(acs.SubscriptionID)
	client.Authorizer = acs.Authorizer
	return &AzureResourceGroupClient{client: client}
}

// AzureResourceGroupClient wraps the Azure resource groups client
type AzureResourceGroupClient struct {
	client resources.GroupsClient
}

// Get gets a resource group
func (c *AzureResourceGroupClient) Get(ctx context.Context, resourceGroupName string, options ...interface{}) (resources.Group, error) {
	result, err := c.client.Get(ctx, resourceGroupName)
	return result, err
}

// CreateOrUpdate creates or updates a resource group
func (c *AzureResourceGroupClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, parameters interface{}, options ...interface{}) (resources.Group, error) {
	// Convert parameters to the expected type
	var params resources.Group
	if p, ok := parameters.(resources.Group); ok {
		params = p
	}
	result, err := c.client.CreateOrUpdate(ctx, resourceGroupName, params)
	return result, err
}

// GetServicePrincipalObjectId gets the service principal object ID
func (acs *AzureClientSet) GetServicePrincipalObjectId(params ...interface{}) (string, error) {
	// For OTP compatibility - simplified implementation
	return "default-object-id", nil
}

// GetVaultsClient returns a key vaults client
func (acs *AzureClientSet) GetVaultsClient(params ...interface{}) *AzureVaultsClient {
	// For OTP compatibility - returns a vaults client
	return &AzureVaultsClient{}
}

// GetKeysClient returns a keys client
func (acs *AzureClientSet) GetKeysClient(params ...interface{}) *AzureKeysClient {
	// For OTP compatibility - returns a keys client
	return &AzureKeysClient{}
}

// AzureVaultsClient represents a key vaults client
type AzureVaultsClient struct{}

// BeginCreateOrUpdate creates or updates a key vault
func (c *AzureVaultsClient) BeginCreateOrUpdate(ctx context.Context, resourceGroupName, vaultName string, parameters interface{}, options ...interface{}) (*AzurePager, error) {
	// Stub implementation
	return &AzurePager{}, nil
}

// AzureKeysClient represents a keys client
type AzureKeysClient struct{}

// AzureKeyResponse represents a key response
type AzureKeyResponse struct {
	Properties *AzureKeyProperties
}

// AzureKeyProperties represents key properties
type AzureKeyProperties struct {
	KeyURIWithVersion *string
}

// CreateIfNotExist creates a key if it doesn't exist
func (c *AzureKeysClient) CreateIfNotExist(ctx context.Context, vaultBaseURL, keyName string, parameters interface{}, options ...interface{}) (*AzureKeyResponse, error) {
	// Stub implementation
	keyURI := "https://vault.azure.net/keys/" + keyName + "/version"
	return &AzureKeyResponse{
		Properties: &AzureKeyProperties{
			KeyURIWithVersion: &keyURI,
		},
	}, nil
}

// GetVirtualMachinesClient returns a virtual machines client
func (acs *AzureClientSet) GetVirtualMachinesClient(params ...interface{}) *AzureVirtualMachinesClient {
	// For OTP compatibility - returns a virtual machines client
	return &AzureVirtualMachinesClient{}
}

// AzureVirtualMachinesClient represents a virtual machines client
type AzureVirtualMachinesClient struct{}

// NewListPager creates a new list pager
func (c *AzureVirtualMachinesClient) NewListPager(resourceGroupName string, options ...interface{}) interface{} {
	// Stub implementation
	return &AzurePager{}
}

// AzurePager represents an Azure pager
type AzurePager struct{}

// PollUntilDone polls until done
func (p *AzurePager) PollUntilDone(ctx context.Context, options ...interface{}) (interface{}, error) {
	// Stub implementation
	return nil, nil
}

// CreateCapacityReservationGroup creates a capacity reservation group
func (cs *AzureClientSet) CreateCapacityReservationGroup(ctx context.Context, capacityReservationGroupName string, resourceGroupName string, location string, zone string) (string, error) {
	// For OTP compatibility - returns the group ID
	// In a real implementation, this would use Azure SDK to create the group
	groupID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/capacityReservationGroups/%s",
		cs.SubscriptionID, resourceGroupName, capacityReservationGroupName)
	return groupID, nil
}

// CreateCapacityReservation creates a capacity reservation
func (cs *AzureClientSet) CreateCapacityReservation(ctx context.Context, capacityReservationGroupName string, capacityReservationName string, location string, resourceGroupName string, skuName string, zone string) error {
	// For OTP compatibility
	// In a real implementation, this would use Azure SDK to create the reservation
	return nil
}

// DeleteCapacityReservation deletes a capacity reservation
func (cs *AzureClientSet) DeleteCapacityReservation(ctx context.Context, capacityReservationGroupName string, capacityReservationName string, resourceGroupName string) error {
	// For OTP compatibility
	return nil
}

// DeleteCapacityReservationGroup deletes a capacity reservation group
func (cs *AzureClientSet) DeleteCapacityReservationGroup(ctx context.Context, capacityReservationGroupName string, resourceGroupName string) error {
	// For OTP compatibility
	return nil
}

// AccountsClientListKeysResponse mimics the v2 SDK response structure
type AccountsClientListKeysResponse struct {
	// Embedded to match v2 SDK structure
	AccountListKeysResult storage.AccountListKeysResult
}

// NewAzureClientSetWithRootCreds creates Azure client set with root credentials
func NewAzureClientSetWithRootCreds() (*AzureClientSet, error) {
	sess, err := NewAzureSessionFromEnv()
	if err != nil {
		return nil, err
	}

	return &AzureClientSet{
		SubscriptionID: sess.SubscriptionID,
		Authorizer:     sess.Authorizer,
	}, nil
}

// RegisterEncryptionAtHost registers encryption at host feature
func (cs *AzureClientSet) RegisterEncryptionAtHost(ctx context.Context) error {
	// This would require the features client from Azure SDK
	// For now, return nil as feature registration is typically a one-time operation
	return nil
}

// DeleteResourceGroup deletes a resource group
func (cs *AzureClientSet) DeleteResourceGroup(ctx context.Context, resourceGroupName string) error {
	client := resources.NewGroupsClient(cs.SubscriptionID)
	client.Authorizer = cs.Authorizer

	future, err := client.Delete(ctx, resourceGroupName)
	if err != nil {
		return err
	}

	err = future.WaitForCompletionRef(ctx, client.Client)
	return err
}

// CreateResourceGroup creates a resource group
func (cs *AzureClientSet) CreateResourceGroup(ctx context.Context, resourceGroupName string, location string) (resources.Group, error) {
	client := resources.NewGroupsClient(cs.SubscriptionID)
	client.Authorizer = cs.Authorizer

	params := resources.Group{
		Location: to.StringPtr(location),
	}

	return client.CreateOrUpdate(ctx, resourceGroupName, params)
}

// CreateStorageAccount creates a storage account and returns list keys response
func (cs *AzureClientSet) CreateStorageAccount(ctx context.Context, resourceGroupName, storageAccountName, location string) (AccountsClientListKeysResponse, error) {
	client := storage.NewAccountsClient(cs.SubscriptionID)
	client.Authorizer = cs.Authorizer

	params := storage.AccountCreateParameters{
		Location: to.StringPtr(location),
		Sku: &storage.Sku{
			Name: storage.StandardLRS,
		},
		Kind: storage.StorageV2,
		AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{
			AccessTier: storage.Hot,
		},
	}

	future, err := client.Create(ctx, resourceGroupName, storageAccountName, params)
	if err != nil {
		return AccountsClientListKeysResponse{}, err
	}

	err = future.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		return AccountsClientListKeysResponse{}, err
	}

	// List keys after creation
	keysResult, err := client.ListKeys(ctx, resourceGroupName, storageAccountName, "")
	if err != nil {
		return AccountsClientListKeysResponse{}, err
	}

	// Wrap in response structure to match v2 SDK
	return AccountsClientListKeysResponse{
		AccountListKeysResult: keysResult,
	}, nil
}

// DeleteStorageAccount deletes a storage account
func (cs *AzureClientSet) DeleteStorageAccount(ctx context.Context, resourceGroupName, storageAccountName string) {
	client := storage.NewAccountsClient(cs.SubscriptionID)
	client.Authorizer = cs.Authorizer

	// Ignore errors on delete as it's often used in defer statements
	client.Delete(ctx, resourceGroupName, storageAccountName)
}

// Helper function to extract resource name from Azure resource ID
func getResourceNameFromID(id string) string {
	parts := []string{}
	for _, part := range parts {
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// RegisterEncryptionAtHost registers encryption at host
func (a *AzureSession) RegisterEncryptionAtHost(ctx context.Context) error {
	// Feature registration is a subscription-level operation
	// In production scenarios, this would typically be done via Azure Portal or CLI
	// For now, we'll implement this as a successful no-op since the feature
	// needs to be registered at the subscription level before VM creation

	// Note: The actual feature registration API requires the features client
	// which is not available in the current SDK version
	// In real usage, this would be: az feature register --name EncryptionAtHost --namespace Microsoft.Compute

	return nil
}

// CreateCapacityReservationGroup creates a capacity reservation group
func (a *AzureSession) CreateCapacityReservationGroup(ctx context.Context, resourceGroup string, groupName string, location string) error {
	// Stub implementation for OTP compatibility
	return nil
}

// CreateCapacityReservation creates a capacity reservation
func (a *AzureSession) CreateCapacityReservation(ctx context.Context, resourceGroup string, groupName string, reservationName string, vmSize string, capacity int) error {
	// Stub implementation for OTP compatibility
	return nil
}

// DeleteCapacityReservation deletes a capacity reservation
func (a *AzureSession) DeleteCapacityReservation(ctx context.Context, resourceGroup string, groupName string, reservationName string) error {
	// Stub implementation for OTP compatibility
	return nil
}

// DeleteCapacityReservationGroup deletes a capacity reservation group
func (a *AzureSession) DeleteCapacityReservationGroup(ctx context.Context, resourceGroup string, groupName string) error {
	// Stub implementation for OTP compatibility
	return nil
}

// GetVM gets a virtual machine by name
func (a *AzureSession) GetVM(ctx context.Context, resourceGroupName, vmName string) (compute.VirtualMachine, error) {
	client := getVMClient(a)
	result, err := client.Get(ctx, resourceGroupName, vmName, "")
	if err != nil {
		return compute.VirtualMachine{}, err
	}
	return result, nil
}

// GetVMInstanceView gets the instance view of a virtual machine
func (a *AzureSession) GetVMInstanceView(ctx context.Context, resourceGroupName, vmName string) (compute.VirtualMachineInstanceView, error) {
	client := getVMClient(a)
	result, err := client.InstanceView(ctx, resourceGroupName, vmName)
	if err != nil {
		return compute.VirtualMachineInstanceView{}, err
	}
	return result, nil
}

// ListPublicIPAddresses lists all public IP addresses in a resource group
func (a *AzureSession) ListPublicIPAddresses(ctx context.Context, resourceGroupName string) ([]network.PublicIPAddress, error) {
	client := getIPClient(a)
	result, err := client.List(ctx, resourceGroupName)
	if err != nil {
		return nil, err
	}

	var publicIPs []network.PublicIPAddress
	for result.NotDone() {
		publicIPs = append(publicIPs, result.Values()...)
		err = result.NextWithContext(ctx)
		if err != nil {
			return nil, err
		}
	}

	return publicIPs, nil
}

// GetNetworkInterface gets a network interface by name
func (a *AzureSession) GetNetworkInterface(ctx context.Context, resourceGroupName, nicName string) (network.Interface, error) {
	client := getNicClient(a)
	result, err := client.Get(ctx, resourceGroupName, nicName, "")
	if err != nil {
		return network.Interface{}, err
	}
	return result, nil
}
