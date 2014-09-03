package main

import (
    "testing"
    "os"
    "launchpad.net/goamz/aws"
    "vip/main"
)

func TestRegion(t *testing.T) {
    regionName =  "us-west-2"
    os.Setenv("AWS_REGION", regionName)
    region = getRegion()
    if region.Name == regionName {
        t.log('Region test passed')
    } else {
        t.Error("Region test failed: expected region %s recieved region %s", regionName, region.Name)
    }
}

func TestRegionDefault(t *testing T) {
    regionName = "blarg"
    os.Setenv("AWS_REGION", regionName)
    region = getRegion()
    if region.Name = "us-east-1"{
        t.log("Default Region test passed")
    } else {
        t.Error("Default Region test failed: default region us-east-1 not returned")
    }
}
