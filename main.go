package main

import (
    "strconv"

    "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        _, err := createInfrastructure(ctx)
        return err
    })
}

type Infrastructure struct {
    firewall *hcloud.Firewall
    server   *hcloud.Server
}

func createInfrastructure(ctx *pulumi.Context) (*Infrastructure, error) {
    firewall, err := hcloud.NewFirewall(ctx, "web-server", &hcloud.FirewallArgs{
        Name: pulumi.String("http-ingress"),
        Rules: hcloud.FirewallRuleArray{
            &hcloud.FirewallRuleArgs{
                Direction: pulumi.String("in"),
                Protocol:  pulumi.String("tcp"),
                Port:      pulumi.String("80"),
                SourceIps: pulumi.StringArray{
                    pulumi.String("0.0.0.0/0"),
                },
            },
            &hcloud.FirewallRuleArgs{
                Direction: pulumi.String("in"),
                Protocol:  pulumi.String("tcp"),
                Port:      pulumi.String("22"),
                SourceIps: pulumi.StringArray{
                    pulumi.String("0.0.0.0/0"),
                },
            },
        },
    })
    if err != nil {
        return nil, err
    }
    server, err := hcloud.NewServer(ctx, "web-server", &hcloud.ServerArgs{
        Image:       pulumi.String("debian-9"),
        ServerType:  pulumi.String("cx11"),
        Location:    pulumi.String("nbg1"),
        FirewallIds: pulumi.IntArray{IDtoInt(firewall.CustomResourceState)},
        UserData: pulumi.String(
`#!/bin/bash -v
apt-get update
apt-get install -y nginx
`),
    })
    if err != nil {
        return nil, err
    }
    master, err := hcloud.NewFloatingIp(ctx, "web-server", &hcloud.FloatingIpArgs{
        Type:         pulumi.String("ipv4"),
        HomeLocation: pulumi.String("nbg1"),
    })
    if err != nil {
        return nil, err
    }
    _, err = hcloud.NewFloatingIpAssignment(ctx, "web-server", &hcloud.FloatingIpAssignmentArgs{
        FloatingIpId: IDtoInt(master.CustomResourceState),
        ServerId:     IDtoInt(server.CustomResourceState),
    })
    if err != nil {
        return nil, err
    }
    ctx.Export("publicIp", server.Ipv4Address)
    return &Infrastructure{
        server:   server,
        firewall: firewall,
    }, nil
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
