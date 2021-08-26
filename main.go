package main

import (
    "strconv"

    "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        node1, err := hcloud.NewServer(ctx, "node1", &hcloud.ServerArgs{
            Image:      pulumi.String("debian-9"),
            ServerType: pulumi.String("cx11"),
            Datacenter: pulumi.String("fsn1-dc8"),
        })
        if err != nil {
            return err
        }
        master, err := hcloud.NewFloatingIp(ctx, "master", &hcloud.FloatingIpArgs{
            Type:         pulumi.String("ipv4"),
            HomeLocation: pulumi.String("nbg1"),
        })
        if err != nil {
            return err
        }
        _, err = hcloud.NewFloatingIpAssignment(ctx, "main", &hcloud.FloatingIpAssignmentArgs{
            FloatingIpId: IDtoInt(master.CustomResourceState),
            ServerId:     IDtoInt(node1.CustomResourceState),
        })
        if err != nil {
            return err
        }
        return nil
    })
}

func IDtoInt(crs pulumi.CustomResourceState) pulumi.IntOutput {
    return crs.ID().ApplyT(func(id pulumi.ID) int {
        number, err := strconv.Atoi(string(id))
        if err != nil {
            panic(err)
        }
        return number
    }).(pulumi.IntOutput)
}
