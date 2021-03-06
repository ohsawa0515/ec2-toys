package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// EC2Client -
type EC2Client interface {
	ListInstances(string) (Ec2Instances, error)
}

// ec2Client -
type ec2Client struct {
	client ec2iface.EC2API
}

// NewEC2Client is construct of ec2 object.
func NewEC2Client(svc ec2iface.EC2API) EC2Client {
	return &ec2Client{
		client: svc,
	}
}

// Ec2Instances is list of EC2 instance.
type Ec2Instances []*ec2.Instance

func (instance Ec2Instances) Len() int {
	return len(instance)
}

func (instance Ec2Instances) Swap(i, j int) {
	instance[i], instance[j] = instance[j], instance[i]
}

func (instance Ec2Instances) Less(i, j int) bool {
	return GetTagValue(instance[i], "Name") < GetTagValue(instance[j], "Name")
}

// ParseFilter parse filter option.
func ParseFilter(filters string) []*ec2.Filter {

	// filters e.g. "Name=tag:Foo,Values=Bar Name=instance-type,Values=m1.small"
	var ec2Filter []*ec2.Filter

	re := regexp.MustCompile(`Name=(.+),Values=(.+)`)
	for _, i := range strings.Fields(filters) {
		matches := re.FindAllStringSubmatch(i, -1)
		ec2Filter = append(ec2Filter, &ec2.Filter{
			Name: aws.String(matches[0][1]),
			Values: []*string{
				aws.String(matches[0][2]),
			},
		})
	}
	return ec2Filter
}

// GenerateSession generate session.
func GenerateSession(region, profile string) (*session.Session, error) {

	sessOpt := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}
	if len(region) != 0 {
		sessOpt.Config = aws.Config{Region: aws.String(region)}
	}
	if len(profile) != 0 {
		sessOpt.Profile = profile
	}
	return session.NewSessionWithOptions(sessOpt)
}

// ListInstances lists one or more of your instances.
func (svc *ec2Client) ListInstances(filters string) (Ec2Instances, error) {

	var instances Ec2Instances
	params := &ec2.DescribeInstancesInput{}
	if len(filters) != 0 {
		params = &ec2.DescribeInstancesInput{
			Filters: ParseFilter(filters),
		}
	}
	resp, err := svc.client.DescribeInstances(params)
	if err != nil {
		return nil, err
	}
	if len(resp.Reservations) == 0 {
		return Ec2Instances{}, nil
	}
	for _, res := range resp.Reservations {
		for _, instance := range res.Instances {
			instances = append(instances, instance)
		}
	}
	sort.Sort(instances)
	return instances, nil
}

// PrintInstances is output stdout.
func PrintInstances(instances Ec2Instances) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 2, 8, 2, ' ', 0)
	for _, instance := range instances {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			GetTagValue(instance, "Name"),
			GetPrivateIPAddress(instance),
			GetPublicIPAddress(instance),
			*instance.InstanceId,
			*instance.InstanceType,
			*instance.Placement.AvailabilityZone,
			*instance.State.Name,
			GetPlatform(instance),
		)
	}
	w.Flush()
}

// GetTagValue returns values of EC2 tag.
func GetTagValue(instance *ec2.Instance, tagName string) string {
	for _, t := range instance.Tags {
		if *t.Key == tagName {
			return *t.Value
		}
	}
	return ""
}

// GetPrivateIPAddress returns value of EC2 private ip address, if there is a value.
func GetPrivateIPAddress(instance *ec2.Instance) string {
	if instance.PrivateIpAddress != nil {
		return *instance.PrivateIpAddress
	}
	return ""
}

// GetPublicIPAddress returns value of EC2 public ip address, if there is a value.
func GetPublicIPAddress(instance *ec2.Instance) string {
	if instance.PublicIpAddress != nil {
		return *instance.PublicIpAddress
	}
	return ""
}

// GetPlatform returns platform name(linux or windows).
func GetPlatform(instance *ec2.Instance) string {
	if instance.Platform != nil {
		return *instance.Platform
	}
	return "linux"
}
