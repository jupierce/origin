package compat_otp

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"
)

// getRandomString generates a random string of specified length
func getRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// NutanixClient represents a Nutanix Prism client
type NutanixClient struct {
	Provider      string
	Username      string
	Password      string
	Endpoint      string
	Port          string
	SkipTLSVerify bool

	// For OTP compatibility - these are used by the REST API methods
	nutanixToken string
	nutanixHost  string

	// Internal client
	client *v3.Client
}

// NewNutanixClient creates a new Nutanix client
func NewNutanixClient(provider, username, password, endpoint, port string, skipTLSVerify bool) (*NutanixClient, error) {
	client := &NutanixClient{
		Provider:      provider,
		Username:      username,
		Password:      password,
		Endpoint:      endpoint,
		Port:          port,
		SkipTLSVerify: skipTLSVerify,
	}

	// Initialize the Prism client
	if err := client.initClient(); err != nil {
		return nil, err
	}

	return client, nil
}

// NewNutanixClientForRestAPI creates a new Nutanix client for REST API operations (OTP compatibility)
func NewNutanixClientForRestAPI(nutanixToken, nutanixHost string) (*NutanixClient, error) {
	client := &NutanixClient{
		nutanixToken: nutanixToken,
		nutanixHost:  nutanixHost,
	}
	return client, nil
}

// initClient initializes the Prism API client
func (n *NutanixClient) initClient() error {
	// Build endpoint URL
	endpoint := fmt.Sprintf("https://%s:%s", n.Endpoint, n.Port)

	// Create client configuration
	credentials := prismgoclient.Credentials{
		URL:      endpoint,
		Endpoint: endpoint,
		Username: n.Username,
		Password: n.Password,
		Insecure: n.SkipTLSVerify,
	}

	// Create the client
	client, err := v3.NewV3Client(credentials)
	if err != nil {
		return fmt.Errorf("failed to create Nutanix client: %w", err)
	}

	n.client = client
	return nil
}

// GetVM gets a virtual machine by UUID
func (n *NutanixClient) GetVM(ctx context.Context, vmUUID string) (*v3.VMIntentResponse, error) {
	if n.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	vm, err := n.client.V3.GetVM(ctx, vmUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	return vm, nil
}

// ListVMs lists all virtual machines
func (n *NutanixClient) ListVMs(ctx context.Context) (*v3.VMListIntentResponse, error) {
	if n.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Create list metadata
	length := int64(100)
	offset := int64(0)
	listMetadata := &v3.DSMetadata{
		Length: &length, // Get up to 100 VMs
		Offset: &offset,
	}

	vms, err := n.client.V3.ListVM(ctx, listMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	return vms, nil
}

// CreateVM creates a new virtual machine
func (n *NutanixClient) CreateVM(ctx context.Context, request *v3.VMIntentInput) (*v3.VMIntentResponse, error) {
	if n.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	vm, err := n.client.V3.CreateVM(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	return vm, nil
}

// DeleteVM deletes a virtual machine
func (n *NutanixClient) DeleteVM(ctx context.Context, vmUUID string) error {
	if n.client == nil {
		return fmt.Errorf("client not initialized")
	}

	_, err := n.client.V3.DeleteVM(ctx, vmUUID)
	if err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	return nil
}

// UpdateVM updates a virtual machine
func (n *NutanixClient) UpdateVM(ctx context.Context, vmUUID string, request *v3.VMIntentInput) (*v3.VMIntentResponse, error) {
	if n.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	vm, err := n.client.V3.UpdateVM(ctx, vmUUID, request)
	if err != nil {
		return nil, fmt.Errorf("failed to update VM: %w", err)
	}

	return vm, nil
}

// PowerOnVM powers on a virtual machine
func (n *NutanixClient) PowerOnVM(ctx context.Context, vmUUID string) error {
	if n.client == nil {
		return fmt.Errorf("client not initialized")
	}

	// Get current VM state
	vm, err := n.GetVM(ctx, vmUUID)
	if err != nil {
		return err
	}

	// Update power state
	request := &v3.VMIntentInput{
		Spec:     vm.Spec,
		Metadata: vm.Metadata,
	}
	powerStateOn := "ON"
	request.Spec.Resources.PowerState = &powerStateOn

	_, err = n.UpdateVM(ctx, vmUUID, request)
	if err != nil {
		return fmt.Errorf("failed to power on VM: %w", err)
	}

	return nil
}

// PowerOffVM powers off a virtual machine
func (n *NutanixClient) PowerOffVM(ctx context.Context, vmUUID string) error {
	if n.client == nil {
		return fmt.Errorf("client not initialized")
	}

	// Get current VM state
	vm, err := n.GetVM(ctx, vmUUID)
	if err != nil {
		return err
	}

	// Update power state
	request := &v3.VMIntentInput{
		Spec:     vm.Spec,
		Metadata: vm.Metadata,
	}
	powerStateOff := "OFF"
	request.Spec.Resources.PowerState = &powerStateOff

	_, err = n.UpdateVM(ctx, vmUUID, request)
	if err != nil {
		return fmt.Errorf("failed to power off VM: %w", err)
	}

	return nil
}

// GetCluster gets cluster information
func (n *NutanixClient) GetCluster(ctx context.Context, clusterUUID string) (*v3.ClusterIntentResponse, error) {
	if n.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	cluster, err := n.client.V3.GetCluster(ctx, clusterUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	return cluster, nil
}

// ListClusters lists all clusters
func (n *NutanixClient) ListClusters(ctx context.Context) (*v3.ClusterListIntentResponse, error) {
	if n.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Create list metadata
	length := int64(100)
	offset := int64(0)
	listMetadata := &v3.DSMetadata{
		Length: &length,
		Offset: &offset,
	}

	clusters, err := n.client.V3.ListCluster(ctx, listMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	return clusters, nil
}

// GetImage gets an image by UUID
func (n *NutanixClient) GetImage(ctx context.Context, imageUUID string) (*v3.ImageIntentResponse, error) {
	if n.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	image, err := n.client.V3.GetImage(ctx, imageUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return image, nil
}

// ListImages lists all images
func (n *NutanixClient) ListImages(ctx context.Context) (*v3.ImageListIntentResponse, error) {
	if n.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Create list metadata
	length := int64(100)
	offset := int64(0)
	listMetadata := &v3.DSMetadata{
		Length: &length,
		Offset: &offset,
	}

	images, err := n.client.V3.ListImage(ctx, listMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	return images, nil
}

// GetSubnet gets a subnet by UUID
func (n *NutanixClient) GetSubnet(ctx context.Context, subnetUUID string) (*v3.SubnetIntentResponse, error) {
	if n.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	subnet, err := n.client.V3.GetSubnet(ctx, subnetUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subnet: %w", err)
	}

	return subnet, nil
}

// ListSubnets lists all subnets
func (n *NutanixClient) ListSubnets(ctx context.Context) (*v3.SubnetListIntentResponse, error) {
	if n.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Create list metadata
	length := int64(100)
	offset := int64(0)
	listMetadata := &v3.DSMetadata{
		Length: &length,
		Offset: &offset,
	}

	subnets, err := n.client.V3.ListSubnet(ctx, listMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to list subnets: %w", err)
	}

	return subnets, nil
}

// WaitForVMPowerState waits for a VM to reach the desired power state
func (n *NutanixClient) WaitForVMPowerState(ctx context.Context, vmUUID string, desiredState string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		vm, err := n.GetVM(ctx, vmUUID)
		if err != nil {
			return err
		}

		if vm.Status != nil && vm.Status.Resources != nil &&
			vm.Status.Resources.PowerState != nil &&
			*vm.Status.Resources.PowerState == desiredState {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			// Continue polling
		}
	}

	return fmt.Errorf("timeout waiting for VM power state to become %s", desiredState)
}

// Close closes the client connection
func (n *NutanixClient) Close() error {
	// Prism Go client doesn't require explicit closing
	n.client = nil
	return nil
}

// GetNutanixVMUUID returns a VM UUID for a given VM name
func (n *NutanixClient) GetNutanixVMUUID(vmName string) (string, error) {
	cmdCurl := `curl -s -X POST --header "Content-Type: application/json" \
	--header "Accept: application/json" \
	--header "Authorization: Basic %v" \
	"https://%v/api/nutanix/v3/vms/list" \
	-d '{ "kind": "vm","filter": "","length": 60,"offset": 0}' |
	jq -r '.entities[] | select(.spec.name == "'"%v"'") | .metadata.uuid'
	`
	formattedCmd := fmt.Sprintf(cmdCurl, n.nutanixToken, n.nutanixHost, vmName)
	uuid, cmdErr := exec.Command("bash", "-c", formattedCmd).Output()
	if cmdErr != nil || string(uuid) == "" {
		return "", cmdErr
	}
	return strings.TrimRight(string(uuid), "\n"), nil
}

// GetNutanixVMState returns the power state of a VM
func (n *NutanixClient) GetNutanixVMState(vmUUID string) (string, error) {
	cmdCurl := `curl -s --header "Content-Type: application/json"\
	--header "Authorization: Basic %v" \
	"https://%v/api/nutanix/v3/vms/%v" \
	| jq -r '.spec.resources.power_state'
	`
	formattedCmd := fmt.Sprintf(cmdCurl, n.nutanixToken, n.nutanixHost, vmUUID)
	state, cmdErr := exec.Command("bash", "-c", formattedCmd).Output()
	if cmdErr != nil || string(state) == "" {
		return "", cmdErr
	}
	return strings.TrimRight(string(state), "\n"), nil
}

// ChangeNutanixVMState changes the power state of a VM
func (n *NutanixClient) ChangeNutanixVMState(vmUUID string, targetState string) error {
	cmdCurl := `curl -s --header "Content-Type: application/json" \
	--header "Accept: application/json" \
	--header "Authorization: Basic %v" \
	"https://%v/api/nutanix/v3/vms/%v"  \
	| jq 'del(.status) | .spec.resources.power_state |= "%v"' > %v
	`
	currentTime := time.Now()
	dateTimeString := currentTime.Format("20060102")
	randStr := getRandomString(5) // Default length like OTP
	filePath := "/tmp/" + randStr + dateTimeString + ".json"
	formattedCmd := fmt.Sprintf(cmdCurl, n.nutanixToken, n.nutanixHost, vmUUID, targetState, filePath)
	_, cmdErr := exec.Command("bash", "-c", formattedCmd).Output()
	defer func() {
		if err := os.RemoveAll(filePath); err != nil {
			fmt.Printf("Error removing file %v: %v\n", filePath, err.Error())
		}
	}()
	if cmdErr != nil {
		return cmdErr
	}

	// Submit the payload to change the VM state
	updateAPI := `curl -s -X 'PUT' --header "Content-Type: application/json" --header "Accept: application/json" --header "Authorization: Basic %v" "https://%v/api/nutanix/v3/vms/%v" -d @%v`
	formattedUpdateCmd := fmt.Sprintf(updateAPI, n.nutanixToken, n.nutanixHost, vmUUID, filePath)
	_, cmdErr = exec.Command("bash", "-c", formattedUpdateCmd).Output()
	if cmdErr != nil {
		return cmdErr
	}
	return nil
}
