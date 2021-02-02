package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Ec2Svc struct {
	ec2 *ec2.EC2
}

func (s *Ec2Svc) FetchInstances() ([]*ec2.Instance, error) {
	var instances []*ec2.Instance

	params := &ec2.DescribeInstancesInput{}

	err := s.ec2.DescribeInstancesPages(
		params,
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {

			for _, rev := range page.Reservations {
				instances = append(instances, rev.Instances...)
			}

			return !lastPage
		},
	)

	if err != nil {
		return nil, err
	}

	return instances, nil
}

func NewEc2Svc() *Ec2Svc {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.ApNortheast1RegionID),
	}))

	svc := ec2.New(sess)

	return &Ec2Svc{
		ec2: svc,
	}
}
