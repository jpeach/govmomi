package cluster

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	vim "github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vsan"
	"github.com/vmware/govmomi/vsan/methods"
	"github.com/vmware/govmomi/vsan/types"
)

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}

type stretch struct {
	*flags.DatacenterFlag
}

func init() {
	cli.Register("cluster.stretch", &stretch{})
}

func (cmd *stretch) Usage() string {
	return "CLUSTER"
}

func (cmd *stretch) Description() string {
	return `Configure witness host to form a vSAN Stretched Cluster
Examples:
  govc cluster.stretch <vsan_cluster_name> <esxi_vsan_witness_node> <primary_fault_domain>`
}

func (cmd *stretch) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.DatacenterFlag, ctx = flags.NewDatacenterFlag(ctx)
	cmd.DatacenterFlag.Register(ctx, f)
}

func (cmd *stretch) Run(ctx context.Context, f *flag.FlagSet) error {
	var clusterPath string

	switch f.NArg() {
	case 1:
		clusterPath = f.Arg(0)
	default:
		return flag.ErrHelp
	}

	client, err := cmd.Client()
	if err != nil {
		return err
	}

	finder, err := cmd.Finder()
	if err != nil {
		return err
	}

	v, err := vsan.NewClient(context.TODO(), client)
	check(err)

	clusterResource, err := finder.ClusterComputeResource(ctx, clusterPath)
	if err != nil {
		return err
	}

	primaryHost, err := finder.HostSystem(ctx, "/datacenter/host/cluster/10.187.108.20")
	if err != nil {
		return err
	}

	secondaryHost, err := finder.HostSystem(ctx, "/datacenter/host/cluster/10.187.109.118")
	if err != nil {
		return err
	}

	fmt.Printf("cluster ref -> %#v\n", clusterResource)

	req := types.VSANVcConvertToStretchedCluster{
		This:    vsan.VsanVcStretchedClusterSystem,
		Cluster: clusterResource.Reference(),
		FaultDomainConfig: types.VimClusterVSANStretchedClusterFaultDomainConfig{
			DynamicData:   vim.DynamicData{},
			FirstFdName:   "Primary",
			FirstFdHosts:  []vim.ManagedObjectReference{primaryHost.Reference()},
			SecondFdName:  "Secondary",
			SecondFdHosts: []vim.ManagedObjectReference{secondaryHost.Reference()},
		},
		WitnessHost: primaryHost.Reference(),
		PreferredFd: "Primary",
		DiskMapping: nil,
	}

	resp, err := methods.VSANVcConvertToStretchedCluster(ctx, v, &req)
	check(err)

	fmt.Printf("%#v", resp)
	/*
			dc, err := finder.Datacenter(ctx, os.Getenv("GOVC_DATACENTER"))
			check(err)
			finder.SetDatacenter(dc)

			clusterComputeResource, err := finder.ClusterComputeResourceList(ctx, "*")
			check(err)

			for _, cluster := range clusterComputeResource {
				_, err = v.VsanClusterGetConfig(context.TODO(), cluster.Reference())
				check(err)
			}

			hosts, err := finder.HostSystemList(context.TODO(), "/datacenter/host/cluster")
			check(err)

			req := vsantypes.VSANIsWitnessVirtualAppliance{
				This: vsan.VsanVcStretchedClusterSystem,
				Hosts: func() []types.ManagedObjectReference {
					var r []types.ManagedObjectReference
					for _, h := range hosts {
						ref := h.Reference()
						fmt.Printf("host '%s/%s'\n", ref.Type, ref.Value)
						r = append(r, h.Reference())
					}
					return r
				}(),
			}

			res, err := methods.VSANIsWitnessVirtualAppliance(context.TODO(), v, &req)
		check(err)

		for _, r := range res.Returnval {
			fmt.Printf("%s -> %t\n", r.HostKey, r.IsVirtualApp)
		}
	*/

	return nil
}
