package internal

import (
	"errors"

	"github.com/mmatur/aws-mfa/internal/types"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

// PromptSurvey prompts survey to user
func PromptSurvey(devices []string) (*types.SurveyAnswer, error) {
	if len(devices) == 0 {
		return nil, errors.New("no devices")
	}

	var qs []*survey.Question

	answer := &types.SurveyAnswer{}

	if len(devices) > 1 {
		qs = append(qs, &survey.Question{
			Name: "mfa-device",
			Prompt: &survey.Select{
				Message: "Choose your mfa device",
				Options: devices,
			},
		})
	} else {
		answer.Device = devices[0]
	}

	qs = append(qs, &survey.Question{
		Name:      "mfa-code",
		Prompt:    &survey.Input{Message: "What is your MFA code?"},
		Validate:  survey.Required,
		Transform: survey.Title,
	})

	err := survey.Ask(qs, answer)
	if err != nil {
		return nil, err
	}

	answer.CleanAnswers()

	return answer, nil
}
