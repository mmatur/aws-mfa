package types

import "strings"

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
