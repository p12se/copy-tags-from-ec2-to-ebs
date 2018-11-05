package main

import (
	"encoding/json"
	"github.com/sadlil/gologger"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/caarlos0/env"
)

type Config struct {
	Region string `env:"AWS_REGION,required"`
}

type ResponseDescribeVolumes struct {
	Volumes []Volume `json:"Volumes"`
}

type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type Attachment struct {
	AttachTime          time.Time `json:"AttachTime"`
	DeleteOnTermination bool      `json:"DeleteOnTermination"`
	Device              string    `json:"Device"`
	InstanceID          string    `json:"InstanceId"`
	State               string    `json:"State"`
	VolumeID            string    `json:"VolumeId"`
}

type Volume struct {
	Attachments      []Attachment `json:"Attachments"`
	AvailabilityZone string       `json:"AvailabilityZone"`
	CreateTime       time.Time    `json:"CreateTime"`
	Encrypted        bool         `json:"Encrypted"`
	Iops             interface{}  `json:"Iops"`
	KmsKeyID         interface{}  `json:"KmsKeyId"`
	Size             int          `json:"Size"`
	SnapshotID       string       `json:"SnapshotId"`
	State            string       `json:"State"`
	Tags             []Tag        `json:"Tags"`
	VolumeID         string       `json:"VolumeId"`
	VolumeType       string       `json:"VolumeType"`
}

type Instance struct {
	InstanceID     string `json:"InstanceId"`
	RootDeviceName string `json:"RootDeviceName"`
	Tags           []Tag  `json:"Tags"`
}

type ResponseDescribeInstances struct {
	Reservations []struct {
		Instance []Instance `json:"Instances"`
		OwnerID  string     `json:"OwnerId"`
	} `json:"Reservations"`
}

func (v *Volume) tagExists() bool {
	if len(v.Tags) > 0 {
		return true
	}
	return false
}

func (a *Attachment) getInstanseId() string {
	return a.InstanceID
}

var logger = gologger.GetLogger(gologger.CONSOLE, gologger.SimpleLog)

func copyTag(i Instance, v Volume, c Config) {

	//TODO: optimize for once call
	cfg := &aws.Config{Region: aws.String(c.Region)}
	session, err := session.NewSession(cfg)

	if err != nil {
		//logger.Error("Session not start")
		os.Exit(2)
	}

	svc := ec2.New(session, cfg)
	var tags []*ec2.Tag

	for _, v := range i.Tags {
		tags = append(tags, &ec2.Tag{
			Key:   aws.String(v.Key),
			Value: aws.String(v.Value),
		})
	}

	input := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(v.VolumeID)},
		Tags: tags,
	}

	_, err = svc.CreateTags(input)

	if err != nil {
		logger.Error("Error copy tags")
		os.Exit(3)
	}

}

func main() {
	conf := Config{}
	err := env.Parse(&conf)

	if err != nil {
		//logger.Fatal(err.Error())
	}

	cfg := &aws.Config{Region: aws.String(conf.Region)}
	session, err := session.NewSession(cfg)

	if err != nil {
		logger.Error("Session not start")
		os.Exit(1)
	}
	svc := ec2.New(session, cfg)

	input := &ec2.DescribeVolumesInput{}

	describeVolumes, err := svc.DescribeVolumes(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				logger.Error(aerr.Error())
			}
		} else {
			logger.Error(err.Error())
		}
		return
	}

	describeVolumesJson, err := json.Marshal(describeVolumes)

	if err != nil {
		os.Exit(2)
	}

	responseDescribeVolumes := ResponseDescribeVolumes{}
	responseDescribeInstances := ResponseDescribeInstances{}

	json.Unmarshal(describeVolumesJson, &responseDescribeVolumes)

	for _, volume := range responseDescribeVolumes.Volumes {

		if !volume.tagExists() {
			for _, i := range volume.Attachments {
				input := &ec2.DescribeInstancesInput{
					InstanceIds: []*string{
						aws.String(i.getInstanseId()),
					},
				}

				describeInstances, _ := svc.DescribeInstances(input)
				describeInstancesJson, _ := json.MarshalIndent(describeInstances, "", "  ")

				logger.Info(string(describeInstancesJson))

				json.Unmarshal(describeInstancesJson, &responseDescribeInstances)

				for _, reservation := range responseDescribeInstances.Reservations {
					for _, instance := range reservation.Instance {
						copyTag(instance, volume, conf)
					}
				}
			}
		}
	}
}
