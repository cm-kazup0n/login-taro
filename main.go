package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"go.uber.org/zap"
)

type Host struct {
	Name           string
	KeyName        string
	PublicDnsName  string
	PrivateDnsName string
}

func FindNameTag(tags []*ec2.Tag) (string, error) {
	for _, tag := range tags {
		if *tag.Key == "Name" {
			return *tag.Value, nil
		}
	}
	return "", errors.New("No Name tag found.")
}

func SSHCommand(host Host, bastions map[string]Host, sockOpt string) (string, bool) {
	bastion, ok := bastions[host.KeyName]
	if sockOpt != "" {
		fmt.Sprintf("-o ProxyCommand=\\'nc -x %s %%h %%p\\'")
	}
	if host == bastion {
		//bastions
		if sockOpt != "" {
			return fmt.Sprintf("%s | ssh -o ProxyCommand='nc -x %s %%%%h %%%%p ' -i ~/.ssh/%s.pem ec2-user@%s", host.Name, sockOpt, host.KeyName, bastion.PublicDnsName), true
		} else {
			return fmt.Sprintf("%s | ssh -i ~/.ssh/%s.pem ec2-user@%s", host.Name, host.KeyName, bastion.PublicDnsName), true
		}
	} else if ok {
		//instances in private subnet
		if sockOpt != "" {
			return fmt.Sprintf("%s | ssh -o ProxyCommand='ssh -i ~/.ssh/%s.pem -W %%h:%%p -o ProxyCommand=\\'nc -x %s %%%%h %%%%p\\' ec2-user@%s' -i ~/.ssh/%s.pem ec2-user@%s", host.Name, host.KeyName, sockOpt, bastion.PublicDnsName, host.KeyName, host.PrivateDnsName), true
		} else {
			return fmt.Sprintf("%s | ssh -o ProxyCommand='ssh -i ~/.ssh/%s.pem -W %%h:%%p ec2-user@%s' -i ~/.ssh/%s.pem ec2-user@%s", host.Name, host.KeyName, bastion.PublicDnsName, host.KeyName, host.PrivateDnsName), true
		}
	} else {
		return "", false
	}
}

var (
	regionOpt = flag.String("region", "", "region")
	sockOpt   = flag.String("sock", "", "sock (e.g. proxy.hoge.com:1080)")
	logger, _ = zap.NewDevelopment()
)

func main() {
	defer logger.Sync()

	//parse flags
	flag.Parse()

	config := &aws.Config{Region: aws.String(*regionOpt)}

	ec2client := ec2.New(session.Must(session.NewSession(config)))

	output, err := ec2client.DescribeInstances(&ec2.DescribeInstancesInput{})

	if err != nil {
		logger.DPanic("Failed to get instances", zap.Error(err), zap.String("region", *regionOpt))
	}

	//{keyname -> bastion}
	bastions := map[string]Host{}
	hosts := []Host{}

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			name, err := FindNameTag(instance.Tags)
			if err != nil {
				logger.Warn("Name of instance not found.", zap.String("instace_id", *instance.InstanceId))
				name = *instance.InstanceId
			}
			host := Host{
				Name:           name,
				KeyName:        *instance.KeyName,
				PrivateDnsName: *instance.PrivateDnsName,
				PublicDnsName:  *instance.PublicDnsName}
			if strings.Contains(host.Name, "bastion") {
				bastions[host.KeyName] = host
			}
			hosts = append(hosts, host)
		}
	}

	for _, host := range hosts {
		command, ok := SSHCommand(host, bastions, *sockOpt)
		if ok {
			fmt.Println(command)
		}
	}

}
