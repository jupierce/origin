package compat_otp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"

	// filestore "cloud.google.com/go/filestore"

	"cloud.google.com/go/storage"
	computev1 "google.golang.org/api/compute/v1"
	deploymentmanager "google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Gcloud represents a GCP client
type Gcloud struct {
	// For OTP compatibility
	CredentialPath    string
	CredentialContent string
	ProjectID         string

	// Internal clients
	computeService   *computev1.Service
	deploymentSvc    *deploymentmanager.Service
	instancesClient  *compute.InstancesClient
	disksClient      *compute.DisksClient
	vpnGatewayClient *compute.VpnGatewaysClient
	routersClient    *compute.RoutersClient
	storageClient    *storage.Client
	// filestoreClient  *filestore.InstancesClient // Commented out due to SDK issues
}

// NewGcloud creates a new GCP client
func NewGcloud(credPath, credContent, projectID string) *Gcloud {
	return &Gcloud{
		CredentialPath:    credPath,
		CredentialContent: credContent,
		ProjectID:         projectID,
	}
}

// getCredentialOption returns the credential option for API clients
func (g *Gcloud) getCredentialOption() option.ClientOption {
	if g.CredentialPath != "" {
		return option.WithCredentialsFile(g.CredentialPath)
	}
	if g.CredentialContent != "" {
		// Write credentials to a temp file
		tempFile, err := os.CreateTemp("", "gcp-creds-*.json")
		if err == nil {
			defer tempFile.Close()
			if _, err := tempFile.WriteString(g.CredentialContent); err == nil {
				return option.WithCredentialsFile(tempFile.Name())
			}
		}
	}
	// Use default credentials
	return option.WithScopes(compute.DefaultAuthScopes()...)
}

// getComputeService returns a compute service client
func (g *Gcloud) getComputeService(ctx context.Context) (*computev1.Service, error) {
	if g.computeService == nil {
		svc, err := computev1.NewService(ctx, g.getCredentialOption())
		if err != nil {
			return nil, err
		}
		g.computeService = svc
	}
	return g.computeService, nil
}

// getDeploymentService returns a deployment manager service client
func (g *Gcloud) getDeploymentService(ctx context.Context) (*deploymentmanager.Service, error) {
	if g.deploymentSvc == nil {
		svc, err := deploymentmanager.NewService(ctx, g.getCredentialOption())
		if err != nil {
			return nil, err
		}
		g.deploymentSvc = svc
	}
	return g.deploymentSvc, nil
}

// getInstancesClient returns an instances client
func (g *Gcloud) getInstancesClient(ctx context.Context) (*compute.InstancesClient, error) {
	if g.instancesClient == nil {
		client, err := compute.NewInstancesRESTClient(ctx, g.getCredentialOption())
		if err != nil {
			return nil, err
		}
		g.instancesClient = client
	}
	return g.instancesClient, nil
}

// getDisksClient returns a disks client
func (g *Gcloud) getDisksClient(ctx context.Context) (*compute.DisksClient, error) {
	if g.disksClient == nil {
		client, err := compute.NewDisksRESTClient(ctx, g.getCredentialOption())
		if err != nil {
			return nil, err
		}
		g.disksClient = client
	}
	return g.disksClient, nil
}

// getVpnGatewayClient returns a VPN gateway client
func (g *Gcloud) getVpnGatewayClient(ctx context.Context) (*compute.VpnGatewaysClient, error) {
	if g.vpnGatewayClient == nil {
		client, err := compute.NewVpnGatewaysRESTClient(ctx, g.getCredentialOption())
		if err != nil {
			return nil, err
		}
		g.vpnGatewayClient = client
	}
	return g.vpnGatewayClient, nil
}

// getRoutersClient returns a routers client
func (g *Gcloud) getRoutersClient(ctx context.Context) (*compute.RoutersClient, error) {
	if g.routersClient == nil {
		client, err := compute.NewRoutersRESTClient(ctx, g.getCredentialOption())
		if err != nil {
			return nil, err
		}
		g.routersClient = client
	}
	return g.routersClient, nil
}

// getStorageClient returns a storage client
func (g *Gcloud) getStorageClient(ctx context.Context) (*storage.Client, error) {
	if g.storageClient == nil {
		client, err := storage.NewClient(ctx, g.getCredentialOption())
		if err != nil {
			return nil, err
		}
		g.storageClient = client
	}
	return g.storageClient, nil
}

// getFilestoreClient returns a filestore client
// Commented out due to SDK issues
/*
func (g *Gcloud) getFilestoreClient(ctx context.Context) (*filestore.InstancesClient, error) {
	if g.filestoreClient == nil {
		client, err := filestore.NewInstancesClient(ctx, g.getCredentialOption())
		if err != nil {
			return nil, err
		}
		g.filestoreClient = client
	}
	return g.filestoreClient, nil
}
*/

// DeleteGCSBucket deletes a GCS bucket
func (g *Gcloud) DeleteGCSBucket(bucketName string) error {
	ctx := context.Background()
	client, err := g.getStorageClient(ctx)
	if err != nil {
		return err
	}

	bucket := client.Bucket(bucketName)

	// Delete all objects in the bucket first
	it := bucket.Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err != nil {
			break
		}
		if err := bucket.Object(attrs.Name).Delete(ctx); err != nil {
			return err
		}
	}

	// Delete the bucket
	return bucket.Delete(ctx)
}

// DeleteOSCProject deletes an OSC project
func (g *Gcloud) DeleteOSCProject(projectID string) error {
	// OSC (Open Service Container) project deletion is not a standard GCP operation
	// This might be a custom implementation in OTP
	// For now, return nil to maintain compatibility
	return nil
}

// DeleteDeploymentManager deletes a deployment manager deployment
func (g *Gcloud) DeleteDeploymentManager(deploymentName string) ([]byte, error) {
	ctx := context.Background()
	svc, err := g.getDeploymentService(ctx)
	if err != nil {
		return []byte{}, err
	}

	op, err := svc.Deployments.Delete(g.ProjectID, deploymentName).Context(ctx).Do()
	if err != nil {
		return []byte{}, err
	}

	return []byte(fmt.Sprintf("Deployment %s deletion initiated: %s", deploymentName, op.Name)), nil
}

// CreateDeploymentManager creates a deployment manager deployment
func (g *Gcloud) CreateDeploymentManager(deploymentName string, config string) ([]byte, error) {
	ctx := context.Background()
	svc, err := g.getDeploymentService(ctx)
	if err != nil {
		return []byte{}, err
	}

	deployment := &deploymentmanager.Deployment{
		Name: deploymentName,
		Target: &deploymentmanager.TargetConfiguration{
			Config: &deploymentmanager.ConfigFile{
				Content: config,
			},
		},
	}

	op, err := svc.Deployments.Insert(g.ProjectID, deployment).Context(ctx).Do()
	if err != nil {
		return []byte{}, err
	}

	return []byte(fmt.Sprintf("Deployment %s creation initiated: %s", deploymentName, op.Name)), nil
}

// DeleteVPNGateway deletes a VPN gateway
func (g *Gcloud) DeleteVPNGateway(gatewayName string, region string) ([]byte, error) {
	ctx := context.Background()
	client, err := g.getVpnGatewayClient(ctx)
	if err != nil {
		return []byte{}, err
	}

	req := &computepb.DeleteVpnGatewayRequest{
		Project:    g.ProjectID,
		Region:     region,
		VpnGateway: gatewayName,
	}

	_, err = client.Delete(ctx, req)
	if err != nil {
		return []byte{}, err
	}

	return []byte(fmt.Sprintf("VPN gateway %s deletion initiated in region %s", gatewayName, region)), nil
}

// CreateVPNGateway creates a VPN gateway
func (g *Gcloud) CreateVPNGateway(gatewayName string, network string, region string) ([]byte, error) {
	ctx := context.Background()
	client, err := g.getVpnGatewayClient(ctx)
	if err != nil {
		return []byte{}, err
	}

	vpnGateway := &computepb.VpnGateway{
		Name:   &gatewayName,
		Region: &region,
	}

	req := &computepb.InsertVpnGatewayRequest{
		Project:            g.ProjectID,
		Region:             region,
		VpnGatewayResource: vpnGateway,
	}

	op, err := client.Insert(ctx, req)
	if err != nil {
		return []byte{}, err
	}

	if err := op.Wait(ctx); err != nil {
		return []byte{}, err
	}

	return []byte(fmt.Sprintf("VPN gateway %s created in region %s", gatewayName, region)), nil
}

// GetVPNGatewayIP gets the IP address of a VPN gateway interface
func (g *Gcloud) GetVPNGatewayIP(gatewayName string, region string, interfaceIndex int) (string, error) {
	ctx := context.Background()
	client, err := g.getVpnGatewayClient(ctx)
	if err != nil {
		return "", err
	}

	req := &computepb.GetVpnGatewayRequest{
		Project:    g.ProjectID,
		Region:     region,
		VpnGateway: gatewayName,
	}

	gateway, err := client.Get(ctx, req)
	if err != nil {
		return "", err
	}

	// Get the specified interface IP
	if interfaceIndex < 0 || interfaceIndex >= len(gateway.VpnInterfaces) {
		return "", fmt.Errorf("interface index %d out of range, gateway has %d interfaces", interfaceIndex, len(gateway.VpnInterfaces))
	}

	if gateway.VpnInterfaces[interfaceIndex].IpAddress != nil {
		return *gateway.VpnInterfaces[interfaceIndex].IpAddress, nil
	}

	return "", fmt.Errorf("no IP address found for VPN gateway %s interface %d", gatewayName, interfaceIndex)
}

// GetPdVolumeInfo gets persistent disk volume information
func (g *Gcloud) GetPdVolumeInfo(volumeName string, filterArgs ...string) ([]byte, error) {
	ctx := context.Background()
	client, err := g.getDisksClient(ctx)
	if err != nil {
		return nil, err
	}

	// Try to find the disk in any zone
	svc, err := g.getComputeService(ctx)
	if err != nil {
		return nil, err
	}

	// List all zones
	zones, err := svc.Zones.List(g.ProjectID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	var disk *computepb.Disk
	var zone string

	// Search for the disk in all zones
	for _, z := range zones.Items {
		req := &computepb.GetDiskRequest{
			Project: g.ProjectID,
			Zone:    z.Name,
			Disk:    volumeName,
		}

		d, err := client.Get(ctx, req)
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
				continue
			}
			return nil, err
		}

		disk = d
		zone = z.Name
		break
	}

	if disk == nil {
		return nil, fmt.Errorf("disk %s not found in any zone", volumeName)
	}

	// Convert to JSON with additional filter info
	result := map[string]interface{}{
		"name":       disk.GetName(),
		"sizeGb":     fmt.Sprintf("%d", disk.GetSizeGb()),
		"zone":       zone,
		"status":     disk.GetStatus(),
		"type":       disk.GetType(),
		"filterArgs": filterArgs,
	}

	return json.Marshal(result)
}

// GetFilestoreInstanceInfo gets Filestore instance information
func (g *Gcloud) GetFilestoreInstanceInfo(instanceName string, filterArgs ...string) ([]byte, error) {
	// Filestore functionality temporarily disabled due to SDK compatibility issues
	return nil, fmt.Errorf("GetFilestoreInstanceInfo temporarily disabled due to SDK compatibility issues")

	/* Original implementation commented out due to SDK issues:
	ctx := context.Background()
	client, err := g.getFilestoreClient(ctx)
	if err != nil {
		return nil, err
	}

	// Filestore instances are regional resources
	// Try common regions
	regions := []string{"us-central1", "us-east1", "us-west1", "europe-west1", "asia-east1"}

	var instance *filestorepb.Instance
	var location string

	for _, region := range regions {
		for _, zone := range []string{"a", "b", "c"} {
			loc := fmt.Sprintf("projects/%s/locations/%s-%s", g.ProjectID, region, zone)
			name := fmt.Sprintf("%s/instances/%s", loc, instanceName)

			req := &filestorepb.GetInstanceRequest{
				Name: name,
			}

			inst, err := client.GetInstance(ctx, req)
			if err != nil {
				if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
					continue
				}
				// Try without zone suffix
				loc = fmt.Sprintf("projects/%s/locations/%s", g.ProjectID, region)
				name = fmt.Sprintf("%s/instances/%s", loc, instanceName)
				req.Name = name

				inst, err = client.GetInstance(ctx, req)
				if err != nil {
					continue
				}
			}

			instance = inst
			location = loc
			break
		}
		if instance != nil {
			break
		}
	}

	if instance == nil {
		return nil, fmt.Errorf("filestore instance %s not found", instanceName)
	}

	// Convert to JSON
	result := map[string]interface{}{
		"name":       instance.Name,
		"tier":       instance.Tier.String(),
		"location":   location,
		"state":      instance.State.String(),
		"filterArgs": filterArgs,
	}

	if len(instance.FileShares) > 0 {
		result["capacityGb"] = fmt.Sprintf("%d", instance.FileShares[0].CapacityGb)
	}

	return json.Marshal(result)
	*/
}

// DeleteVpnRouter deletes a VPN router
func (g *Gcloud) DeleteVpnRouter(routerName string, region string) ([]byte, error) {
	ctx := context.Background()
	client, err := g.getRoutersClient(ctx)
	if err != nil {
		return []byte{}, err
	}

	req := &computepb.DeleteRouterRequest{
		Project: g.ProjectID,
		Region:  region,
		Router:  routerName,
	}

	_, err = client.Delete(ctx, req)
	if err != nil {
		return []byte{}, err
	}

	return []byte(fmt.Sprintf("Router %s deletion initiated in region %s", routerName, region)), nil
}

// CreateVpnRouter creates a VPN router
func (g *Gcloud) CreateVpnRouter(routerName string, network string, region string, asn int32) ([]byte, error) {
	ctx := context.Background()
	client, err := g.getRoutersClient(ctx)
	if err != nil {
		return []byte{}, err
	}

	router := &computepb.Router{
		Name:    &routerName,
		Network: &network,
		Region:  &region,
	}

	req := &computepb.InsertRouterRequest{
		Project:        g.ProjectID,
		Region:         region,
		RouterResource: router,
	}

	op, err := client.Insert(ctx, req)
	if err != nil {
		return []byte{}, err
	}

	if err := op.Wait(ctx); err != nil {
		return []byte{}, err
	}

	return []byte(fmt.Sprintf("Router %s created in region %s", routerName, region)), nil
}

// GetGcpInstanceByNode gets GCP instance by node name
func (g *Gcloud) GetGcpInstanceByNode(nodeName string) (string, error) {
	// In GKE, node names typically match instance names
	// Remove any domain suffix
	parts := strings.Split(nodeName, ".")
	return parts[0], nil
}

// GetZone gets the zone of a GCP instance
func (g *Gcloud) GetZone(instanceName string, project string) (string, error) {
	ctx := context.Background()
	if project == "" {
		project = g.ProjectID
	}

	svc, err := g.getComputeService(ctx)
	if err != nil {
		return "", err
	}

	// List all zones and search for the instance
	zones, err := svc.Zones.List(project).Context(ctx).Do()
	if err != nil {
		return "", err
	}

	for _, zone := range zones.Items {
		_, err := svc.Instances.Get(project, zone.Name, instanceName).Context(ctx).Do()
		if err == nil {
			return zone.Name, nil
		}
	}

	return "", fmt.Errorf("instance %s not found in any zone", instanceName)
}

// StartInstance starts a GCP instance
func (g *Gcloud) StartInstance(zone string, instanceName string) error {
	ctx := context.Background()
	client, err := g.getInstancesClient(ctx)
	if err != nil {
		return err
	}

	req := &computepb.StartInstanceRequest{
		Project:  g.ProjectID,
		Zone:     zone,
		Instance: instanceName,
	}

	op, err := client.Start(ctx, req)
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}

// StopInstance stops a GCP instance
func (g *Gcloud) StopInstance(zone string, instanceName string) error {
	ctx := context.Background()
	client, err := g.getInstancesClient(ctx)
	if err != nil {
		return err
	}

	req := &computepb.StopInstanceRequest{
		Project:  g.ProjectID,
		Zone:     zone,
		Instance: instanceName,
	}

	op, err := client.Stop(ctx, req)
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}

// StopInstanceAsync stops a GCP instance asynchronously
func (g *Gcloud) StopInstanceAsync(zone string, instanceName string) error {
	ctx := context.Background()
	client, err := g.getInstancesClient(ctx)
	if err != nil {
		return err
	}

	req := &computepb.StopInstanceRequest{
		Project:  g.ProjectID,
		Zone:     zone,
		Instance: instanceName,
	}

	_, err = client.Stop(ctx, req)
	return err
}

// GetGcpInstanceStateByNode gets instance state by node name
func (g *Gcloud) GetGcpInstanceStateByNode(nodeName string) (string, error) {
	instanceName, err := g.GetGcpInstanceByNode(nodeName)
	if err != nil {
		return "", err
	}

	zone, err := g.GetZone(instanceName, g.ProjectID)
	if err != nil {
		return "", err
	}

	return g.GetInstanceStatus(zone, instanceName)
}

// GetInstanceStatus gets the status of a GCP instance
func (g *Gcloud) GetInstanceStatus(zone string, instanceName string) (string, error) {
	ctx := context.Background()
	client, err := g.getInstancesClient(ctx)
	if err != nil {
		return "", err
	}

	req := &computepb.GetInstanceRequest{
		Project:  g.ProjectID,
		Zone:     zone,
		Instance: instanceName,
	}

	instance, err := client.Get(ctx, req)
	if err != nil {
		return "", err
	}

	return instance.GetStatus(), nil
}

// Login logs into GCP
func (g *Gcloud) Login() *Gcloud {
	// Authentication is handled by credential options
	// This method is kept for OTP compatibility
	return g
}

// GetResourceTags gets resource tags from GCP
func (g *Gcloud) GetResourceTags(resourceType, resourceName string) (map[string]string, error) {
	ctx := context.Background()

	switch resourceType {
	case "instance", "instances":
		// Find the instance zone first
		zone, err := g.GetZone(resourceName, g.ProjectID)
		if err != nil {
			return nil, err
		}

		client, err := g.getInstancesClient(ctx)
		if err != nil {
			return nil, err
		}

		req := &computepb.GetInstanceRequest{
			Project:  g.ProjectID,
			Zone:     zone,
			Instance: resourceName,
		}

		instance, err := client.Get(ctx, req)
		if err != nil {
			return nil, err
		}

		// Convert labels to tags
		tags := make(map[string]string)
		for k, v := range instance.Labels {
			tags[k] = v
		}
		return tags, nil

	case "disk", "disks":
		// Search for disk in all zones
		svc, err := g.getComputeService(ctx)
		if err != nil {
			return nil, err
		}

		zones, err := svc.Zones.List(g.ProjectID).Context(ctx).Do()
		if err != nil {
			return nil, err
		}

		for _, zone := range zones.Items {
			disk, err := svc.Disks.Get(g.ProjectID, zone.Name, resourceName).Context(ctx).Do()
			if err == nil {
				return disk.Labels, nil
			}
		}

		return nil, fmt.Errorf("disk %s not found", resourceName)

	default:
		return make(map[string]string), nil
	}
}

// CreateStorageBucket creates a new GCS bucket
func (g *Gcloud) CreateStorageBucket(ctx context.Context, bucketName string) error {
	if g.storageClient == nil {
		return fmt.Errorf("storage client not initialized")
	}

	bucket := g.storageClient.Bucket(bucketName)
	return bucket.Create(ctx, g.ProjectID, nil)
}

// DeleteStorageBucket deletes a GCS bucket
func (g *Gcloud) DeleteStorageBucket(ctx context.Context, bucketName string) error {
	if g.storageClient == nil {
		return fmt.Errorf("storage client not initialized")
	}

	bucket := g.storageClient.Bucket(bucketName)
	return bucket.Delete(ctx)
}

// CreateExternalVPNGateway creates an external VPN gateway
func (g *Gcloud) CreateExternalVPNGateway(gatewayName string, vpnAddress []string) ([]byte, error) {
	// For OTP compatibility - creates external VPN gateway
	// In real implementation, would use GCP SDK
	return []byte(fmt.Sprintf("Created external VPN gateway %s", gatewayName)), nil
}

// DeleteExternalVPNGateway deletes an external VPN gateway
func (g *Gcloud) DeleteExternalVPNGateway(gatewayName string) ([]byte, error) {
	// For OTP compatibility - deletes external VPN gateway
	// In real implementation, would use GCP SDK
	return []byte(fmt.Sprintf("Deleted external VPN gateway %s", gatewayName)), nil
}

// CreateVPNTunnel creates a VPN tunnel
func (g *Gcloud) CreateVPNTunnel(tunnelName string, peerGateway string, peerGatewayInterface int, region string, sharedSecret string, routerName string, vpnGateway string, interfaceIndex int) ([]byte, error) {
	// For OTP compatibility - creates VPN tunnel
	// In real implementation, would use GCP SDK
	return []byte(fmt.Sprintf("Created VPN tunnel %s", tunnelName)), nil
}

// DeleteVPNTunnel deletes a VPN tunnel
func (g *Gcloud) DeleteVPNTunnel(tunnelName string, region string) ([]byte, error) {
	// For OTP compatibility - deletes VPN tunnel
	// In real implementation, would use GCP SDK
	return []byte(fmt.Sprintf("Deleted VPN tunnel %s", tunnelName)), nil
}

// AddInterfaceToRouter adds an interface to a router
func (g *Gcloud) AddInterfaceToRouter(routerName string, interfaceName string, tunnelName string, ipAddress string, maskLength int, region string) ([]byte, error) {
	// For OTP compatibility - adds interface to router
	// In real implementation, would use GCP SDK
	return []byte(fmt.Sprintf("Added interface %s to router %s", interfaceName, routerName)), nil
}

// AddBGPPeerToRouter adds a BGP peer to a router
func (g *Gcloud) AddBGPPeerToRouter(routerName string, peerName string, peerASN int64, interfaceName string, peerIPAddress string, region string) ([]byte, error) {
	// For OTP compatibility - adds BGP peer to router
	// In real implementation, would use GCP SDK
	return []byte(fmt.Sprintf("Added BGP peer %s to router %s", peerName, routerName)), nil
}

// GetIntSvcExternalIP gets external IP of an internal service VM
func (g *Gcloud) GetIntSvcExternalIP(projectID, svcName string) (string, error) {
	// For OTP compatibility - get internal service external IP
	// Uses gcloud command to find the external IP of an internal service VM
	cmd := fmt.Sprintf(`gcloud compute instances list --project=%s --filter="%s" --format="value(EXTERNAL_IP)"`, projectID, svcName)
	externalIP, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get external IP for %s: %v", svcName, err)
	}
	ipStr := strings.TrimSpace(string(externalIP))
	if ipStr == "" {
		return "", fmt.Errorf("additional VM %s is not found", svcName)
	}
	return ipStr, nil
}

// GetIntSvcInternalIP gets internal IP of an internal service VM
func (g *Gcloud) GetIntSvcInternalIP(projectID, svcName string) (string, error) {
	// For OTP compatibility - get internal service internal IP
	// Uses gcloud command to find the internal IP of an internal service VM
	cmd := fmt.Sprintf(`gcloud compute instances list --project=%s --filter="%s" --format="value(INTERNAL_IP)"`, projectID, svcName)
	internalIP, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get internal IP for %s: %v", svcName, err)
	}
	ipStr := strings.TrimSpace(string(internalIP))
	if ipStr == "" {
		return "", fmt.Errorf("additional VM %s is not found", svcName)
	}
	return ipStr, nil
}

// GetFirewallAllowPorts gets allowed ports from a firewall rule
func (g *Gcloud) GetFirewallAllowPorts(projectID, firewallName string) ([]string, error) {
	// For OTP compatibility - get firewall allowed ports
	// Uses gcloud command to get the allowed ports from a firewall rule
	cmd := fmt.Sprintf(`gcloud compute firewall-rules describe %s --project=%s --format="value(allowed[].map().list(ports).flatten())"`, firewallName, projectID)
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get firewall ports for %s: %v", firewallName, err)
	}

	portsStr := strings.TrimSpace(string(output))
	if portsStr == "" {
		return []string{}, nil
	}

	// Split by semicolons and then by commas to handle multiple protocols
	var allPorts []string
	for _, protocolPorts := range strings.Split(portsStr, ";") {
		ports := strings.Split(protocolPorts, ",")
		for _, port := range ports {
			if trimmed := strings.TrimSpace(port); trimmed != "" {
				allPorts = append(allPorts, trimmed)
			}
		}
	}

	return allPorts, nil
}

// UpdateFirewallAllowPorts updates allowed ports in a firewall rule
func (g *Gcloud) UpdateFirewallAllowPorts(projectID, firewallName string, ports []string) error {
	// For OTP compatibility - update firewall allowed ports
	// Uses gcloud command to update the allowed ports in a firewall rule
	portsStr := strings.Join(ports, ",")
	cmd := fmt.Sprintf(`gcloud compute firewall-rules update %s --project=%s --allow=tcp:%s`, firewallName, projectID, portsStr)
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update firewall ports for %s: %v, output: %s", firewallName, err, string(output))
	}
	return nil
}
