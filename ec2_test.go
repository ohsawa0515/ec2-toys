package main

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type mockEC2Iface struct {
	ec2iface.EC2API
}

type expectTag struct {
	name  string
	value string
}

func (svc *mockEC2Iface) DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {

	return &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					{
						InstanceId: aws.String("i-12345678"),
					},
					{
						InstanceId: aws.String("i-abcdefgh"),
					},
				},
			},
		},
	}, nil
}

func TestListInstances(t *testing.T) {
	mockSvc := &mockEC2Iface{}
	mockEC2Client := NewEC2Client(mockSvc)
	instances, err := mockEC2Client.ListInstances("")
	if err != nil {
		t.Errorf("Expected no error, but got %v.", err)
	}
	for _, instance := range instances {
		if len(*instance.InstanceId) == 0 {
			t.Errorf("Expected any instance info, but got empty.")
		}
	}
}

func TestParseFilter(t *testing.T) {
	var expectEC2Filter []*ec2.Filter
	filter := "Name=tag:Foo,Values=Bar Name=instance-type,Values=m1.small"
	expectEC2Filter = append(expectEC2Filter, &ec2.Filter{
		Name: aws.String("tag:Foo"),
		Values: []*string{
			aws.String("Bar"),
		},
	})
	expectEC2Filter = append(expectEC2Filter, &ec2.Filter{
		Name: aws.String("instance-type"),
		Values: []*string{
			aws.String("m1.small"),
		},
	})
	ec2Filter := ParseFilter(filter)
	if !reflect.DeepEqual(ec2Filter, expectEC2Filter) {
		t.Errorf("expect: %s\nactual: %s", expectEC2Filter, ec2Filter)
	}
}

func TestGetTagValue(t *testing.T) {
	var expectTags []expectTag
	expectTags = append(expectTags, expectTag{
		name:  "Name",
		value: "server001",
	})
	expectTags = append(expectTags, expectTag{
		name:  "Env",
		value: "Production",
	})

	instance := &ec2.Instance{
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("server001"),
			},
			{
				Key:   aws.String("Env"),
				Value: aws.String("Production"),
			},
		},
	}
	for _, expectTag := range expectTags {
		actualTagValue := GetTagValue(instance, expectTag.name)
		if len(actualTagValue) == 0 {
			t.Errorf("Expected tag value for name: %s, but got empty.", expectTag.name)
		}
		if expectTag.value != actualTagValue {
			t.Errorf("Expected tag name: %s and value: %s, but got tag value: %v.", expectTag.name, expectTag.value, actualTagValue)
		}
	}
	tagValue := GetTagValue(instance, "Deploy")
	if tagValue != "" {
		t.Errorf("Expected no value of tag no exist name of tag, but go %v.", tagValue)
	}
}

func TestEC2Info(t *testing.T) {
	var instance *ec2.Instance

	instance = &ec2.Instance{
		PrivateIpAddress: aws.String("192.0.2.11"),
		PublicIpAddress:  aws.String("203.0.113.11"),
		Platform:         nil,
	}
	if actual := GetPrivateIPAddress(instance); "192.0.2.11" != actual {
		t.Errorf("Expected private ip address: 192.0.2.11, but got %v", actual)
	}
	if actual := GetPublicIPAddress(instance); "203.0.113.11" != actual {
		t.Errorf("Expected public ip address: 203.0.113.11, but got %v", actual)
	}
	if actual := GetPlatform(instance); "linux" != actual {
		t.Errorf("Expected platform: linux, but got %v", actual)
	}

	instance = &ec2.Instance{
		PrivateIpAddress: nil,
		PublicIpAddress:  nil,
		Platform:         aws.String("windows"),
	}
	if actual := GetPrivateIPAddress(instance); len(actual) != 0 {
		t.Errorf("Expected no value of private ip address, but got %v", actual)
	}
	if actual := GetPublicIPAddress(instance); len(actual) != 0 {
		t.Errorf("Expected no value of public ip address, but got %v", actual)
	}
	if actual := GetPlatform(instance); "windows" != actual {
		t.Errorf("Expected platform: windows, but got %v", actual)
	}
}
