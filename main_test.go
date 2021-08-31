package main

import (
	"fmt"
	"hash/fnv"
	"sync"
	"testing"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

type mocks int

// FNV32a hashes using fnv32a algorithm
func FNV32aStr(text string) string {
    return fmt.Sprint(FNV32aInt(text))
}

func FNV32aInt(text string) int {
    algorithm := fnv.New32a()
    algorithm.Write([]byte(text))
    return int(algorithm.Sum32())
}

func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
    return FNV32aStr(args.Name), args.Inputs, nil
}

func (mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
    return args.Args, nil
}

type testFnc = func(t *testing.T, wg *sync.WaitGroup, infra *Infrastructure)

func Setup(t *testing.T, test testFnc) {
    err := pulumi.RunErr(func(ctx *pulumi.Context) error {
        var wg sync.WaitGroup

        infra, err := createInfrastructure(ctx)
        assert.NoError(t, err)
        test(t, &wg, infra)

        wg.Wait()
        return nil
    }, pulumi.WithMocks("project", "stack", mocks(0)))
    assert.NoError(t, err)
}

func TestFirewall(t *testing.T) {
    Setup(t, func(t *testing.T, wg *sync.WaitGroup, infra *Infrastructure) {
        wg.Add(2)

        pulumi.All(infra.firewall.Name).ApplyT(func(all []interface{}) error {
            name := all[0].(string)
            assert.Equal(t, name, "http-ingress", "Expect a http-ingress firewall rule")
            wg.Done()
            return nil
        })

        // Test if the service has tags and a name tag.
        pulumi.All(infra.firewall.Rules).ApplyT(func(all []interface{}) error {
            firewallRules := all[0].([]hcloud.FirewallRule)

            assert.Len(t, firewallRules, 2, "Expect one firewall rule")
            //expect http rule
            assert.Equal(t, "in", firewallRules[0].Direction, "Expect ingress rule")
            assert.Equal(t, "80", *firewallRules[0].Port, "Expect http ingress rule")
            assert.Containsf(t, firewallRules[0].SourceIps, "0.0.0.0/0", "Expect http ingress rule")
            //expect ssh rule
            assert.Equal(t, "in", firewallRules[1].Direction, "Expect ingress rule")
            assert.Equal(t, "22", *firewallRules[1].Port, "Expect ssh ingress rule")
            assert.Containsf(t, firewallRules[1].SourceIps, "0.0.0.0/0", "Expect http ingress rule")

            wg.Done()
            return nil
        })
    })
}

func TestServer(t *testing.T) {
    Setup(t, func(t *testing.T, wg *sync.WaitGroup, infra *Infrastructure) {
        wg.Add(1)

        pulumi.All(infra.server.UserData ,infra.server.FirewallIds ,infra.server.Image).ApplyT(func(all []interface{}) error {
            userdata := all[0].(*string)
            firewallIds := all[1].([]int)
            image := all[2].(string)
            assert.Equal(t, *userdata, "#!/bin/bash -v\napt-get update\napt-get install -y nginx\n", "Expect nginx installation in userdata script")
            assert.Containsf(t, firewallIds, FNV32aInt("web-server"), "Expect firewall web-server to be set on server")
            assert.Equal(t, image, "debian-9", "Expect debain-9 image")
            wg.Done()
            return nil
        })
    })
}
