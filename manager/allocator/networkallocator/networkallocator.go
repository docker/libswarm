package networkallocator

import (
	"fmt"
	"net"

	"github.com/docker/libnetwork/driverapi"
	"github.com/docker/libnetwork/drivers/overlay/ovmanager"
	"github.com/docker/libnetwork/drvregistry"
	"github.com/docker/libnetwork/ipamapi"
	"github.com/docker/swarm-v2/api"
	"github.com/docker/swarm-v2/log"
	"golang.org/x/net/context"
)

const (
	defaultDriver = "overlay"
)

var (
	defaultDriverInitFunc = ovmanager.Init
)

// NetworkAllocator acts as the controller for all network related operations
// like managing network and IPAM drivers and also creating and
// deleting networks and the associated resources.
type NetworkAllocator struct {
	// The driver register which manages all internal and external
	// IPAM and network drivers.
	drvRegistry *drvregistry.DrvRegistry

	// Local network state used by NetworkAllocator to do network management.
	networks map[string]*network

	// Allocator state to indicate if allocation has been
	// successfully completed for this service.
	services map[string]struct{}
}

// Local in-memory state related to netwok that need to be tracked by NetworkAllocator
type network struct {
	id string

	// pools is used to save the internal poolIDs needed when
	// releasing the pool.
	pools map[string]string

	// endpoints is a map of endpoint IP to the poolID from which it
	// was allocated.
	endpoints map[string]string
}

// New returns a new NetworkAllocator handle
func New() (*NetworkAllocator, error) {
	na := &NetworkAllocator{
		networks: make(map[string]*network),
		services: make(map[string]struct{}),
	}

	// There are no driver configurations and notification
	// functions as of now.
	reg, err := drvregistry.New(nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	// Add the manager component of overlay driver to the registry.
	if err := reg.AddDriver(defaultDriver, defaultDriverInitFunc, nil); err != nil {
		return nil, err
	}

	na.drvRegistry = reg
	return na, nil
}

// Allocate allocates all the necessary resources both general
// and driver-specific which may be specified in the NetworkSpec
func (na *NetworkAllocator) Allocate(n *api.Network) error {
	if _, ok := na.networks[n.ID]; ok {
		return fmt.Errorf("network %s already allocated", n.ID)
	}

	pools, err := na.allocatePools(n)
	if err != nil {
		return fmt.Errorf("failed allocating pools and gateway IP for network %s: %v", n.ID, err)
	}

	if err := na.allocateDriverState(n); err != nil {
		return fmt.Errorf("failed while allocating driver state for network %s: %v", n.ID, err)
	}

	na.networks[n.ID] = &network{
		id:        n.ID,
		pools:     pools,
		endpoints: make(map[string]string),
	}

	return nil
}

func (na *NetworkAllocator) getNetwork(id string) *network {
	return na.networks[id]
}

// Deallocate frees all the general and driver specific resources
// whichs were assigned to the passed network.
func (na *NetworkAllocator) Deallocate(n *api.Network) error {
	localNet := na.getNetwork(n.ID)
	if localNet == nil {
		return fmt.Errorf("could not get networker state for network %s", n.ID)
	}

	if err := na.freeDriverState(n); err != nil {
		return fmt.Errorf("failed to free driver state for network %s: %v", n.ID, err)
	}

	delete(na.networks, n.ID)
	return na.freePools(n, localNet.pools)
}

// ServiceAllocate allocates all the network resources such as IPs.
func (na *NetworkAllocator) ServiceAllocate(s *api.Service) error {
	na.services[s.ID] = struct{}{}
	return nil
}

// ServiceDeallocate de-allocates all the network resources such as IPs.
func (na *NetworkAllocator) ServiceDeallocate(s *api.Service) error {
	delete(na.services, s.ID)

	return nil
}

// IsAllocated returns if the passed network has been allocated or not.
func (na *NetworkAllocator) IsAllocated(n *api.Network) bool {
	_, ok := na.networks[n.ID]
	return ok
}

// IsTaskAllocated returns if the passed task has it's network resources allocated or not.
func (na *NetworkAllocator) IsTaskAllocated(t *api.Task) bool {
	if t.GetContainer() == nil {
		return false
	}

	// If Networks is empty there is no way this Task is allocated.
	if len(t.GetContainer().Networks) == 0 {
		return false
	}

	// To determine whether the task has it's resources allocated,
	// we just need to look at one network(in case of
	// multi-network attachment).  This is because we make sure we
	// allocate for every network or we allocate for none.

	// If the network is not allocated, the task cannot be allocated.
	localNet, ok := na.networks[t.GetContainer().Networks[0].Network.ID]
	if !ok {
		return false
	}

	// Addresses empty. Task is not allocated.
	if len(t.GetContainer().Networks[0].Addresses) == 0 {
		return false
	}

	// The allocated IP address not found in local endpoint state. Not allocated.
	if _, ok := localNet.endpoints[t.GetContainer().Networks[0].Addresses[0]]; !ok {
		return false
	}

	return true
}

// IsServiceAllocated returns if the passed service has it's network resources allocated or not.
func (na *NetworkAllocator) IsServiceAllocated(s *api.Service) bool {
	// If there is no service endpoint configuration there is no
	// possibility of allocation.
	if s.Spec.Endpoint == nil {
		return false
	}

	if _, ok := na.services[s.ID]; !ok {
		return false
	}

	return true
}

// AllocateTask allocates all the endpoint resources for all the
// networks that a task is attached to.
func (na *NetworkAllocator) AllocateTask(t *api.Task) error {
	for i, nAttach := range t.GetContainer().Networks {
		ipam, _, err := na.resolveIPAM(nAttach.Network)
		if err != nil {
			return fmt.Errorf("failed to resolve IPAM while allocating : %v", err)
		}

		localNet := na.getNetwork(nAttach.Network.ID)
		if localNet == nil {
			return fmt.Errorf("could not find networker state")
		}

		if err := na.allocateNetworkIPs(nAttach, ipam, localNet); err != nil {
			if err := na.releaseEndpoints(t.GetContainer().Networks[:i]); err != nil {
				log.G(context.TODO()).Errorf("Failed to release IP addresses while rolling back allocation for task %s network %s: %v", t.ID, nAttach.Network.ID, err)
			}
			return fmt.Errorf("failed to allocate network IP for task %s network %s: %v", t.ID, nAttach.Network.ID, err)
		}
	}

	return nil
}

// DeallocateTask releases all the endpoint resources for all the
// networks that a task is attached to.
func (na *NetworkAllocator) DeallocateTask(t *api.Task) error {
	return na.releaseEndpoints(t.GetContainer().Networks)
}

func (na *NetworkAllocator) releaseEndpoints(networks []*api.Container_NetworkAttachment) error {
	for _, nAttach := range networks {
		ipam, _, err := na.resolveIPAM(nAttach.Network)
		if err != nil {
			return fmt.Errorf("failed to resolve IPAM while allocating : %v", err)
		}

		localNet := na.getNetwork(nAttach.Network.ID)
		if localNet == nil {
			return fmt.Errorf("could not find networker state")
		}

		// Do not fail and bail out if we fail to release IP
		// address here. Keep going and try releasing as many
		// addresses as possible.
		for _, addr := range nAttach.Addresses {
			// Retrieve the poolID and immediately nuke
			// out the mapping.
			poolID := localNet.endpoints[addr]
			delete(localNet.endpoints, addr)

			ip, _, err := net.ParseCIDR(addr)
			if err != nil {
				log.G(context.TODO()).Errorf("Could not parse IP address %s while releasing", addr)
				continue
			}

			if err := ipam.ReleaseAddress(poolID, ip); err != nil {
				log.G(context.TODO()).Errorf("IPAM failure while releasing IP address %s: %v", addr, err)
			}
		}

		// Clear out the address list when we are done with
		// this network.
		nAttach.Addresses = nil
	}

	return nil
}

// allocate the endpoint IP addresses for a single network attachment of the task.
func (na *NetworkAllocator) allocateNetworkIPs(nAttach *api.Container_NetworkAttachment, ipam ipamapi.Ipam, localNet *network) error {
	var ip *net.IPNet

	addresses := nAttach.Addresses
	if addresses == nil {
		addresses = []string{""}
	}

	for i, rawAddr := range addresses {
		var addr net.IP
		if rawAddr != "" {
			var err error
			addr, _, err = net.ParseCIDR(rawAddr)
			if err != nil {
				return err
			}
		}

		for _, poolID := range localNet.pools {
			var err error

			ip, _, err = ipam.RequestAddress(poolID, addr, nil)
			if err != nil && err != ipamapi.ErrNoAvailableIPs && err != ipamapi.ErrIPOutOfRange {
				return fmt.Errorf("could not allocate IP from IPAM: %v", err)
			}

			// If we got an address then we are done.
			if err == nil {
				ipStr := ip.String()
				localNet.endpoints[ipStr] = poolID
				addresses[i] = ipStr
				nAttach.Addresses = addresses
				return nil
			}
		}
	}

	return fmt.Errorf("could not find an available IP")
}

func (na *NetworkAllocator) freeDriverState(n *api.Network) error {
	d, _, err := na.resolveDriver(n)
	if err != nil {
		return err
	}

	return d.NetworkFree(n.ID)
}

func (na *NetworkAllocator) allocateDriverState(n *api.Network) error {
	d, dName, err := na.resolveDriver(n)
	if err != nil {
		return err
	}

	// Construct IPAM data for driver consumption.
	ipv4Data := make([]driverapi.IPAMData, 0, len(n.IPAM.Configs))
	for _, ic := range n.IPAM.Configs {
		if ic.Family == api.IPAMConfig_IPV6 {
			continue
		}

		_, subnet, err := net.ParseCIDR(ic.Subnet)
		if err != nil {
			return fmt.Errorf("error parsing subnet %s while allocating driver state: %v", ic.Subnet, err)
		}

		gwIP := net.ParseIP(ic.Gateway)
		gwNet := &net.IPNet{
			IP:   gwIP,
			Mask: subnet.Mask,
		}

		data := driverapi.IPAMData{
			Pool:    subnet,
			Gateway: gwNet,
		}

		ipv4Data = append(ipv4Data, data)
	}

	ds, err := d.NetworkAllocate(n.ID, nil, ipv4Data, nil)
	if err != nil {
		return err
	}

	// Update network object with the obtained driver state.
	n.DriverState = &api.Driver{
		Name:    dName,
		Options: ds,
	}

	return nil
}

// Resolve network driver
func (na *NetworkAllocator) resolveDriver(n *api.Network) (driverapi.Driver, string, error) {
	dName := defaultDriver
	if n.Spec.DriverConfiguration != nil && n.Spec.DriverConfiguration.Name != "" {
		dName = n.Spec.DriverConfiguration.Name
	}

	d, _ := na.drvRegistry.Driver(dName)
	if d == nil {
		return nil, "", fmt.Errorf("could not resolve network driver %s", dName)
	}

	return d, dName, nil
}

// Resolve the IPAM driver
func (na *NetworkAllocator) resolveIPAM(n *api.Network) (ipamapi.Ipam, string, error) {
	dName := ipamapi.DefaultIPAM
	if n.Spec.IPAM != nil && n.Spec.IPAM.Driver != nil && n.Spec.IPAM.Driver.Name != "" {
		dName = n.Spec.IPAM.Driver.Name
	}

	ipam, _ := na.drvRegistry.IPAM(dName)
	if ipam == nil {
		return nil, "", fmt.Errorf("could not resolve IPAM driver %s", dName)
	}

	return ipam, dName, nil
}

func (na *NetworkAllocator) freePools(n *api.Network, pools map[string]string) error {
	ipam, _, err := na.resolveIPAM(n)
	if err != nil {
		return fmt.Errorf("failed to resolve IPAM while freeing pools for network %s: %v", n.ID, err)
	}

	releasePools(ipam, n.IPAM.Configs, pools)
	return nil
}

func releasePools(ipam ipamapi.Ipam, icList []*api.IPAMConfig, pools map[string]string) {
	for _, ic := range icList {
		if err := ipam.ReleaseAddress(pools[ic.Subnet], net.ParseIP(ic.Gateway)); err != nil {
			log.G(context.TODO()).Errorf("Failed to release address %s: %v", ic.Subnet, err)
		}
	}

	for k, p := range pools {
		if err := ipam.ReleasePool(p); err != nil {
			log.G(context.TODO()).Errorf("Failed to release pool %s: %v", k, err)
		}
	}
}

func (na *NetworkAllocator) allocatePools(n *api.Network) (map[string]string, error) {
	ipam, dName, err := na.resolveIPAM(n)
	if err != nil {
		return nil, err
	}

	// We don't support user defined address spaces yet so just
	// retrive default address space names for the driver.
	_, asName, err := na.drvRegistry.IPAMDefaultAddressSpaces(dName)
	if err != nil {
		return nil, err
	}

	pools := make(map[string]string)

	if n.Spec.IPAM == nil {
		n.Spec.IPAM = &api.IPAMOptions{}
	}

	ipamConfigs := n.Spec.IPAM.Configs
	if ipamConfigs == nil {
		if n.IPAM != nil {
			ipamConfigs = n.IPAM.Configs
		}

		if ipamConfigs == nil {
			ipamConfigs = append(ipamConfigs, &api.IPAMConfig{Family: api.IPAMConfig_IPV4})
		}
	}

	// Update the runtime IPAM configurations with initial state
	n.IPAM = &api.IPAMOptions{
		Driver:  &api.Driver{Name: dName},
		Configs: ipamConfigs,
	}

	for i, ic := range ipamConfigs {
		poolID, poolIP, _, err := ipam.RequestPool(asName, ic.Subnet, ic.Range, nil, false)
		if err != nil {
			// Rollback by releasing all the resources allocated so far.
			releasePools(ipam, ipamConfigs[:i], pools)
			return nil, err
		}
		pools[poolIP.String()] = poolID

		gwIP, _, err := ipam.RequestAddress(poolID, net.ParseIP(ic.Gateway), nil)
		if err != nil {
			// Rollback by releasing all the resources allocated so far.
			releasePools(ipam, ipamConfigs[:i], pools)
			return nil, err
		}

		if ic.Subnet == "" {
			ic.Subnet = poolIP.String()
			ic.Gateway = gwIP.IP.String()
		}
	}

	return pools, nil
}
