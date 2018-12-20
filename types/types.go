package types

import "strings"

const (
	// CredentialsFile is the default path for the AWS credentials.
	CredentialsFile = "/.aws/credentials"
)

// Config holds configuration.
type Config struct {
	CredentialFile string `description:"Credential file. (default: ~/.aws/credentials)"`
	Duration       int64  `description:"Duration in seconds for credentials to remain valid (default: 43200)"`
	Profile        string `description:"AWS profile to use. (default: default)"`
	Force          bool   `description:"Force credentials renew"`
	Debug          bool   `description:"Enable debug"`
}

// NoOption empty struct.
type NoOption struct{}

// SurveyAnswer holds survey answer.
type SurveyAnswer struct {
	Device string `survey:"mfa-device"`
	Code   string `survey:"mfa-code"`
}

// CleanAnswers cleans answers.
func (s *SurveyAnswer) CleanAnswers() {
	if strings.Contains(s.Device, ": ") {
		s.Device = strings.Split(s.Device, ": ")[1]
	}
}
