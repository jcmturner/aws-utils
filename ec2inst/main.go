package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"os"
)

var githash = "No version available"
var buildstamp = "Not set"

const (
	instanceID       = "id"
	region           = "region"
	accountID        = "account"
	availabilityZone = "az"
	arch             = "arch"
	billingProducts  = "billingproducts"
	productCodes     = "productcodes"
	imageID          = "imgid"
	instanceType     = "type"
	kernelID         = "kernelid"
	pendingTime      = "pending"
	privateIP        = "pvtip"
	tags             = "tags"
	vpcID            = "vpcid"
	subnetID         = "subnetid"
	state            = "state"
	securityGroups   = "sg"
	publicIP         = "publicip"
	publicDNS        = "publicdns"
	ebsVols          = "ebs"
)

func main() {
	sess := session.Must(session.NewSession())
	ec2md := ec2metadata.New(sess)
	if !ec2md.Available() {
		fmt.Fprintln(os.Stderr, "Instance metadata not available. Instance may not be running on AWS.")
		os.Exit(1)
	}
	idoc, err := ec2md.GetInstanceIdentityDocument()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting instance identity document: %v\n", err)
		os.Exit(1)
	}
	sess = session.Must(session.NewSession((&aws.Config{
		Region: aws.String(idoc.Region),
	}).WithCredentialsChainVerboseErrors(true)))

	flag.Parse()

	var l []string
	switch flag.Arg(0) {
	case accountID:
		l = append(l, idoc.AccountID)
	case region:
		l = append(l, idoc.Region)
	case availabilityZone:
		l = append(l, idoc.AvailabilityZone)
	case instanceID:
		l = append(l, idoc.InstanceID)
	case arch:
		l = append(l, idoc.Architecture)
	case billingProducts:
		l = append(l, idoc.BillingProducts...)
	case productCodes:
		l = append(l, idoc.DevpayProductCodes...)
	case imageID:
		l = append(l, idoc.ImageID)
	case instanceType:
		l = append(l, idoc.InstanceType)
	case kernelID:
		l = append(l, idoc.KernelID)
	case pendingTime:
		l = append(l, idoc.PendingTime.String())
	case privateIP:
		l = append(l, idoc.PrivateIP)
	case tags:
		d, err := describeInstance(sess, idoc.InstanceID)
		checkErr(err)
		l = append(l, getTags(d)...)
	case vpcID:
		d, err := describeInstance(sess, idoc.InstanceID)
		checkErr(err)
		l = append(l, *d.Reservations[0].Instances[0].VpcId)
	case subnetID:
		d, err := describeInstance(sess, idoc.InstanceID)
		checkErr(err)
		l = append(l, *d.Reservations[0].Instances[0].SubnetId)
	case state:
		d, err := describeInstance(sess, idoc.InstanceID)
		checkErr(err)
		l = append(l, *d.Reservations[0].Instances[0].State.Name)
	case securityGroups:
		d, err := describeInstance(sess, idoc.InstanceID)
		checkErr(err)
		l = append(l, getSGs(d)...)
	case publicDNS:
		d, err := describeInstance(sess, idoc.InstanceID)
		checkErr(err)
		l = append(l, *d.Reservations[0].Instances[0].PublicDnsName)
	case publicIP:
		d, err := describeInstance(sess, idoc.InstanceID)
		checkErr(err)
		l = append(l, *d.Reservations[0].Instances[0].PublicIpAddress)
	case ebsVols:
		d, err := describeInstance(sess, idoc.InstanceID)
		checkErr(err)
		l = append(l, getEBS(d)...)
	default:
		usage()
	}
	for i := range l {
		if l[i] != "" {
			fmt.Println(l[i])
		}
	}
	os.Exit(0)
}

func usage() {
	fmt.Printf(`
	Usage: %s [information type]
		%s			Show this instance's ID.
		%s			Show the account ID in which this instance resides.
		%s			Show the AWS region in which this instance resides.
		%s			Show the availability zone in which this instance resides.
		%s			Show the instance type.
		%s			Show the instance's image ID.
		%s			Show the instance architecture.
		%s		Show the instance's kernel ID.
		%s			Show the instance's pending time.
		%s		Show the instance's product codes.
		%s		Show the instance's billing products.
		%s			Show the instance's state.
		%s 			Show the instance's tags (<key>:<value>)
		%s			Show the instance's EBS volumes (<device name>:<volume id>:<status>:<attach time>:<delete on termination>)
		%s			Show the instance's private IP
		%s		Show the instance's public IP
		%s		Show the instance's public DNS name
		%s			Show the instance's security groups (<group id>:<group name>)
		%s			Show the instance's VPC ID
		%s		Show the instance's subnet ID

	References:
		- http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html
	Version:
		Build hash: %v
		Build timestamp: %v
`,
		os.Args[0],
		instanceID,
		accountID,
		region, availabilityZone,
		instanceType, imageID, arch, kernelID,
		pendingTime, productCodes, billingProducts,
		state,
		tags, ebsVols,
		privateIP, publicIP, publicDNS,
		securityGroups, vpcID, subnetID,
		githash, buildstamp,
	)
	os.Exit(0)
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Ensure the EC2 instance has an IAM role attached that can perform ec2:DescribeInstances.\nError detail: %v\n", err)
		os.Exit(1)
	}
}

func describeInstance(sess *session.Session, instID string) (*ec2.DescribeInstancesOutput, error) {
	var input ec2.DescribeInstancesInput
	input.SetInstanceIds([]*string{&instID})
	ec2Svc := ec2.New(sess)
	return ec2Svc.DescribeInstances(&input)
}

func getTags(output *ec2.DescribeInstancesOutput) (l []string) {
	for _, t := range output.Reservations[0].Instances[0].Tags {
		l = append(l, *t.Key+":"+*t.Value)
	}
	return
}

func getSGs(output *ec2.DescribeInstancesOutput) (l []string) {
	for _, sg := range output.Reservations[0].Instances[0].SecurityGroups {
		l = append(l, *sg.GroupId+":"+*sg.GroupName)
	}
	return
}

func getEBS(output *ec2.DescribeInstancesOutput) (l []string) {
	for _, vol := range output.Reservations[0].Instances[0].BlockDeviceMappings {
		s := fmt.Sprintf("%s:%s:%s:%v:%t", *vol.DeviceName, *vol.Ebs.VolumeId, *vol.Ebs.Status, vol.Ebs.AttachTime, *vol.Ebs.DeleteOnTermination)
		l = append(l, s)
	}
	return
}
