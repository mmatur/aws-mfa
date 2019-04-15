package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// ListMFADevices list available MFA devices.
func ListMFADevices(awsConfig aws.Config) ([]string, error) {
	arn, err := getCurrentUserARN(awsConfig)
	if err != nil {
		return nil, err
	}

	splits := strings.Split(arn, "user/")
	if len(splits) < 2 {
		return nil, fmt.Errorf("unable to split arn %q", arn)
	}

	iAM := iam.New(awsConfig)
	req := iAM.ListMFADevicesRequest(&iam.ListMFADevicesInput{
		UserName: aws.String(splits[1]),
	})

	resp, err := req.Send()
	if err != nil {
		return nil, err
	}

	return displayMFADevices(resp), nil
}

func displayMFADevices(output *iam.ListMFADevicesOutput) []string {
	var devices []string
	if output == nil || len(output.MFADevices) == 0 {
		fmt.Println("* 0 devices.")
	} else {
		for _, d := range output.MFADevices {
			devices = append(devices, fmt.Sprintf("Name %s: %s", aws.StringValue(d.UserName), aws.StringValue(d.SerialNumber)))
		}
	}

	return devices
}
