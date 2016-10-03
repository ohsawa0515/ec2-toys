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

	// filters e.g. "Name=tag:Vuls-Scan,Values=True Name=instance-type,Values=m1.small,m1.medium"
	ec2_filters := make([]*ec2.Filter, 0)

	re_space := regexp.MustCompile(`\s+`)
	re_name := regexp.MustCompile(`Name=`)
	re_values := regexp.MustCompile(`,Values=`)
	for _, i := range re_space.Split(filters, -1) {
		for _, j := range re_name.Split(i, -1) {
			if len(j) != 0 {
				v := re_values.Split(j, -1)
				name := v[0]
				/* // fix bug
				values := make([]*string, 0)
				for _, k := range v[1:] {
					fmt.Println(k)
					values = append(values, aws.String(k))
				}
				*/
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

func generateSession() (*session.Session, error) {
	return session.NewSessionWithOptions(session.Options{})
}

func DescribeInstances(filters string) (ec2_instances, error) {

	var instances ec2_instances
	sess, err := generateSession()
	if err != nil {
		return nil, err
	}
	svc := ec2.New(sess)
	params := &ec2.DescribeInstancesInput{
		Filters: ParseFilter(filters),
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
		fmt.Println(
			GetTagValue(instance, "Name"),
			*instance.PrivateIpAddress,
			*instance.InstanceId,
			*instance.InstanceType,
			*instance.Placement.AvailabilityZone,
			*instance.State.Name,
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
