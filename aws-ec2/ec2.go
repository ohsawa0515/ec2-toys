package aws_ec2

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"regexp"
	"sort"
)

type ec2_instances []*ec2.Instance

func (instance ec2_instances) Len() int {
	return len(instance)
}

func (instance ec2_instances) Swap(i, j int) {
	instance[i], instance[j] = instance[j], instance[i]
}

func (instance ec2_instances) Less(i, j int) bool {
	return GetTagValue(instance[i], "Name") < GetTagValue(instance[j], "Name")
}

func ParseFilter(filters string) []*ec2.Filter {

	// filters e.g. "Name=tag:Foo,Values=Bar Name=instance-type,Values=m1.small"
	ec2_filters := make([]*ec2.Filter, 0)

	re_space := regexp.MustCompile(`\s+`)
	re_name := regexp.MustCompile(`Name=`)
	re_values := regexp.MustCompile(`,Values=`)
	for _, i := range re_space.Split(filters, -1) {
		for _, j := range re_name.Split(i, -1) {
			if len(j) != 0 {
				v := re_values.Split(j, -1)
				name := v[0]
				ec2_filters = append(ec2_filters, &ec2.Filter{
					Name: aws.String(name),
					Values: []*string{
						aws.String(v[1]),
					},
				})
			}
		}
	}
	return ec2_filters
}

func generateSession(region, profile string) (*session.Session, error) {

	sess_opt := session.Options{}
	if len(region) != 0 {
		sess_opt.Config = aws.Config{Region: aws.String(region)}
	}
	if len(profile) != 0 {
		sess_opt.Profile = profile
	}
	return session.NewSessionWithOptions(sess_opt)
}

func DescribeInstances(region, profile, filters string) (ec2_instances, error) {

	var instances ec2_instances
	sess, err := generateSession(region, profile)
	if err != nil {
		return nil, err
	}
	svc := ec2.New(sess)

	params := &ec2.DescribeInstancesInput{}
	if len(filters) != 0 {
		params = &ec2.DescribeInstancesInput{
			Filters: ParseFilter(filters),
		}
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return nil, err
	}
	if len(resp.Reservations) == 0 {
		return ec2_instances{}, nil
	}
	for _, res := range resp.Reservations {
		for _, instance := range res.Instances {
			instances = append(instances, instance)
		}
	}
	sort.Sort(instances)
	return instances, nil
}

func PrintInstances(instances ec2_instances) {
	for _, instance := range instances {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			GetTagValue(instance, "Name"),
			GetPrivateIpAddress(instance),
			GetPublicIpAddress(instance),
			*instance.InstanceId,
			*instance.InstanceType,
			*instance.Placement.AvailabilityZone,
			*instance.State.Name,
			GetPlatform(instance),
		)
	}
}

func GetTagValue(instance *ec2.Instance, tag_name string) string {
	for _, t := range instance.Tags {
		if *t.Key == tag_name {
			return *t.Value
		}
	}
	return ""
}

func GetPrivateIpAddress(instance *ec2.Instance) string {
	if instance.PrivateIpAddress != nil {
		return *instance.PrivateIpAddress
	}
	return ""
}

func GetPublicIpAddress(instance *ec2.Instance) string {
	if instance.PublicIpAddress != nil {
		return *instance.PublicIpAddress
	}
	return ""
}

func GetPlatform(instance *ec2.Instance) string {
	if instance.Platform != nil {
		return *instance.Platform
	}
	return "linux"
}
