package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Profile to insert into the template.
type Profile struct {
	AssumedRole        string `ini:"assumed_role"`
	AssumedRoleARN     string `ini:"assumed_role_arn,omitempty"`
	AWSAccessKeyID     string `ini:"aws_access_key_id"`
	AWSSecretAccessKey string `ini:"aws_secret_access_key"`
	AWSSessionToken    string `ini:"aws_session_token"`
	AWSSecurityToken   string `ini:"aws_security_token"`
	Expiration         string `ini:"expiration"`
}

// GetSessionToken fetch a new token through AWS Security Token Service
func GetSessionToken(awsConfig aws.Config, duration int64, device string, code string) (*Profile, error) {
	if len(device) == 0 && len(code) == 0 {
		return nil, errors.New("device and code are required")
	}

	service := sts.New(awsConfig)
	req := service.GetSessionTokenRequest(&sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(duration),
		SerialNumber:    aws.String(device),
		TokenCode:       aws.String(code),
	})

	resp, err := req.Send(context.Background())
	if err != nil {
		return nil, err
	}

	expirationDate := time.Now().UTC().Add(time.Duration(duration) * time.Second)

	if resp == nil || resp.Credentials == nil {
		return nil, errors.New("unable to read credentials")
	}

	return &Profile{
		AssumedRole:        "False",
		AssumedRoleARN:     "",
		AWSAccessKeyID:     aws.StringValue(resp.Credentials.AccessKeyId),
		AWSSecretAccessKey: aws.StringValue(resp.Credentials.SecretAccessKey),
		AWSSessionToken:    aws.StringValue(resp.Credentials.SessionToken),
		AWSSecurityToken:   aws.StringValue(resp.Credentials.SessionToken),
		Expiration:         expirationDate.Format("2006-01-02 15:04:05"),
	}, nil
}

func getCurrentUserARN(awsConfig aws.Config) (string, error) {
	service := sts.New(awsConfig)
	req := service.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})

	resp, err := req.Send(context.Background())
	if err != nil {
		return "", err
	}

	return aws.StringValue(resp.Arn), nil
}
