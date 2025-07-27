package compat_otp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// Vmware represents a VMware vSphere client
type Vmware struct {
	UserName   string
	Password   string
	VCenters   string
	DataCenter string
	Cluster    string
	DefaultDS  string
	GovmomiURL string

	// Internal client
	client *govmomi.Client
}

// Login logs into vSphere and returns the client
func (vmware *Vmware) Login() (*Vmware, *govmomi.Client) {
	ctx := context.Background()

	u, err := url.Parse(vmware.GovmomiURL)
	if err != nil {
		// Return empty client for compatibility
		return vmware, nil
	}

	// Set credentials
	u.User = url.UserPassword(vmware.UserName, vmware.Password)

	// Create client
	c, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		// Return empty client for compatibility
		return vmware, nil
	}

	vmware.client = c
	return vmware, c
}

// GetVspheresInstance gets a vSphere VM instance
func (vmware *Vmware) GetVspheresInstance(c *govmomi.Client, vmInstance string) (string, error) {
	if c == nil && vmware.client != nil {
		c = vmware.client
	}
	if c == nil {
		return "", fmt.Errorf("no vSphere client available")
	}

	ctx := context.Background()
	finder := find.NewFinder(c.Client, true)

	if vmware.DataCenter != "" {
		dc, err := finder.Datacenter(ctx, vmware.DataCenter)
		if err != nil {
			return "", err
		}
		finder.SetDatacenter(dc)
	}

	vm, err := finder.VirtualMachine(ctx, vmInstance)
	if err != nil {
		return "", err
	}

	return vm.Reference().Value, nil
}

// GetVspheresInstanceState gets the power state of a vSphere VM
func (vmware *Vmware) GetVspheresInstanceState(c *govmomi.Client, vmInstance string) (string, error) {
	if c == nil && vmware.client != nil {
		c = vmware.client
	}
	if c == nil {
		return "", fmt.Errorf("no vSphere client available")
	}

	ctx := context.Background()
	finder := find.NewFinder(c.Client, true)

	if vmware.DataCenter != "" {
		dc, err := finder.Datacenter(ctx, vmware.DataCenter)
		if err != nil {
			return "", err
		}
		finder.SetDatacenter(dc)
	}

	vm, err := finder.VirtualMachine(ctx, vmInstance)
	if err != nil {
		return "", err
	}

	var o mo.VirtualMachine
	err = vm.Properties(ctx, vm.Reference(), []string{"runtime.powerState"}, &o)
	if err != nil {
		return "", err
	}

	return string(o.Runtime.PowerState), nil
}

// StopVsphereInstance stops (powers off) a vSphere VM
func (vmware *Vmware) StopVsphereInstance(c *govmomi.Client, vmInstance string) error {
	if c == nil && vmware.client != nil {
		c = vmware.client
	}
	if c == nil {
		return fmt.Errorf("no vSphere client available")
	}

	ctx := context.Background()
	finder := find.NewFinder(c.Client, true)

	if vmware.DataCenter != "" {
		dc, err := finder.Datacenter(ctx, vmware.DataCenter)
		if err != nil {
			return err
		}
		finder.SetDatacenter(dc)
	}

	vm, err := finder.VirtualMachine(ctx, vmInstance)
	if err != nil {
		return err
	}

	task, err := vm.PowerOff(ctx)
	if err != nil {
		return err
	}

	return task.Wait(ctx)
}

// StartVsphereInstance starts (powers on) a vSphere VM
func (vmware *Vmware) StartVsphereInstance(c *govmomi.Client, vmInstance string) error {
	if c == nil && vmware.client != nil {
		c = vmware.client
	}
	if c == nil {
		return fmt.Errorf("no vSphere client available")
	}

	ctx := context.Background()
	finder := find.NewFinder(c.Client, true)

	if vmware.DataCenter != "" {
		dc, err := finder.Datacenter(ctx, vmware.DataCenter)
		if err != nil {
			return err
		}
		finder.SetDatacenter(dc)
	}

	vm, err := finder.VirtualMachine(ctx, vmInstance)
	if err != nil {
		return err
	}

	task, err := vm.PowerOn(ctx)
	if err != nil {
		return err
	}

	return task.Wait(ctx)
}

// GetVsphereConnectionLogout logs out from vSphere
func (vmware *Vmware) GetVsphereConnectionLogout(c *govmomi.Client) error {
	if c == nil && vmware.client != nil {
		c = vmware.client
	}
	if c == nil {
		return nil
	}

	return c.Logout(context.Background())
}

// GetVMsByTag gets VMs by tag
func (vmware *Vmware) GetVMsByTag(c *govmomi.Client, tagName string) ([]*object.VirtualMachine, error) {
	if c == nil && vmware.client != nil {
		c = vmware.client
	}
	if c == nil {
		return nil, fmt.Errorf("no vSphere client available")
	}

	ctx := context.Background()
	finder := find.NewFinder(c.Client, true)

	if vmware.DataCenter != "" {
		dc, err := finder.Datacenter(ctx, vmware.DataCenter)
		if err != nil {
			return nil, err
		}
		finder.SetDatacenter(dc)
	}

	// Get all VMs
	vms, err := finder.VirtualMachineList(ctx, "*")
	if err != nil {
		return nil, err
	}

	// Filter by tag
	var taggedVMs []*object.VirtualMachine
	pc := property.DefaultCollector(c.Client)

	for _, vm := range vms {
		var o mo.VirtualMachine
		err := pc.RetrieveOne(ctx, vm.Reference(), []string{"tag"}, &o)
		if err != nil {
			continue
		}

		// Check if VM has the tag
		for _, tag := range o.Tag {
			if tag.Key == tagName {
				taggedVMs = append(taggedVMs, vm)
				break
			}
		}
	}

	return taggedVMs, nil
}

// GetDatastore gets a datastore by name
func (vmware *Vmware) GetDatastore(c *govmomi.Client, name string) (*object.Datastore, error) {
	if c == nil && vmware.client != nil {
		c = vmware.client
	}
	if c == nil {
		return nil, fmt.Errorf("no vSphere client available")
	}

	ctx := context.Background()
	finder := find.NewFinder(c.Client, true)

	if vmware.DataCenter != "" {
		dc, err := finder.Datacenter(ctx, vmware.DataCenter)
		if err != nil {
			return nil, err
		}
		finder.SetDatacenter(dc)
	}

	return finder.Datastore(ctx, name)
}

// CreateVM creates a new virtual machine
func (vmware *Vmware) CreateVM(c *govmomi.Client, name string, template string, cpus int32, memoryMB int64) (*object.VirtualMachine, error) {
	if c == nil && vmware.client != nil {
		c = vmware.client
	}
	if c == nil {
		return nil, fmt.Errorf("no vSphere client available")
	}

	ctx := context.Background()
	finder := find.NewFinder(c.Client, true)

	if vmware.DataCenter != "" {
		dc, err := finder.Datacenter(ctx, vmware.DataCenter)
		if err != nil {
			return nil, err
		}
		finder.SetDatacenter(dc)
	}

	// Find template
	tmpl, err := finder.VirtualMachine(ctx, template)
	if err != nil {
		return nil, err
	}

	// Find resource pool
	var pool *object.ResourcePool
	if vmware.Cluster != "" {
		cluster, err := finder.ClusterComputeResource(ctx, vmware.Cluster)
		if err != nil {
			return nil, err
		}
		pool, err = cluster.ResourcePool(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		pool, err = finder.DefaultResourcePool(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Find datastore
	var ds *object.Datastore
	if vmware.DefaultDS != "" {
		ds, err = finder.Datastore(ctx, vmware.DefaultDS)
		if err != nil {
			return nil, err
		}
	} else {
		ds, err = finder.DefaultDatastore(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Clone VM
	poolRef := pool.Reference()
	dsRef := ds.Reference()
	cloneSpec := types.VirtualMachineCloneSpec{
		Location: types.VirtualMachineRelocateSpec{
			Pool:      &poolRef,
			Datastore: &dsRef,
		},
		Config: &types.VirtualMachineConfigSpec{
			Name:     name,
			NumCPUs:  cpus,
			MemoryMB: memoryMB,
		},
	}

	// Find folder
	folder, err := finder.DefaultFolder(ctx)
	if err != nil {
		return nil, err
	}

	task, err := tmpl.Clone(ctx, folder, name, cloneSpec)
	if err != nil {
		return nil, err
	}

	info, err := task.WaitForResult(ctx, nil)
	if err != nil {
		return nil, err
	}

	return object.NewVirtualMachine(c.Client, info.Result.(types.ManagedObjectReference)), nil
}
