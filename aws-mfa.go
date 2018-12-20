package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/containous/flaeg"
	"github.com/mitchellh/go-homedir"
	"github.com/mmatur/aws-mfa/cmd"
	"github.com/mmatur/aws-mfa/meta"
	"github.com/mmatur/aws-mfa/types"
	"github.com/ogier/pflag"
	"gopkg.in/ini.v1"
)

const (
	longTermSuffix = "-long-term"
)

func main() {
	homeDirPath, err := homedir.Dir()
	if err != nil {
		log.Fatalln(err)
	}

	defaultCfg := &types.Config{
		CredentialFile: filepath.Join(homeDirPath, ".aws", "credentials"),
		Duration:       43200,
		Profile:        "default",
	}
	defaultPointerCfg := &types.Config{}

	rootCmd := &flaeg.Command{
		Name:                  "aws-mfa",
		Description:           "AWS - MFA",
		Config:                defaultCfg,
		DefaultPointersConfig: defaultPointerCfg,
		Run: func() error {
			return rootRun(defaultCfg)
		},
	}

	flag := flaeg.New(rootCmd, os.Args[1:])

	// version

	versionCmd := &flaeg.Command{
		Name:                  "version",
		Description:           "Display the version.",
		Config:                &types.NoOption{},
		DefaultPointersConfig: &types.NoOption{},
		Run: func() error {
			meta.DisplayVersion()
			return nil
		},
	}

	flag.AddCommand(versionCmd)

	// Run command

	if err = flag.Run(); err != nil && err != pflag.ErrHelp {
		log.Printf("Error: %v\n", err)
	}
}

func rootRun(config *types.Config) error {
	cfg, err := ini.Load(config.CredentialFile)
	if err != nil {
		return err
	}

	if err = validate(config, cfg); err != nil {
		return err
	}

	if credentialStillValid(config, cfg) {
		return nil
	}

	//awsConfig, err := external.LoadDefaultAWSConfig()
	awsConfig, err := external.LoadDefaultAWSConfig(
		external.WithSharedConfigProfile(config.Profile + longTermSuffix))

	if err != nil {
		return err
	}

	var devices []string
	if cfg.Section(config.Profile + longTermSuffix).HasKey("aws_mfa_device") {
		device := cfg.Section(config.Profile + longTermSuffix).Key("aws_mfa_device").String()
		if config.Debug {
			fmt.Printf("MFA device %q found in %q for profile %q\n", device, config.CredentialFile, config.Profile+longTermSuffix)
		}
		devices = append(devices, device)
	} else {
		devices, err = cmd.ListMFADevices(awsConfig)
		if err != nil {
			return err
		}
	}

	answer, err := cmd.PromptSurvey(devices)
	if err != nil {
		return err
	}

	profile, err := cmd.GetSessionToken(awsConfig, config.Duration, answer.Device, answer.Code)
	if err != nil {
		return err
	}

	if err = updateAWSCredentials(config, cfg, profile); err != nil {
		return err
	}

	fmt.Printf("Success! Credentials for profile %q valid until %s \n", config.Profile, profile.Expiration)
	return nil
}

func credentialStillValid(config *types.Config, cfg *ini.File) bool {
	if cfg.Section(config.Profile).HasKey("expiration") && !config.Force {
		expirationUnparsed := cfg.Section(config.Profile).Key("expiration").String()
		expiration, err := time.Parse("2006-01-02 15:04:05", expirationUnparsed)
		if err != nil {
			log.Fatalf("Unable to parse %s", expirationUnparsed)
		}

		secondsRemaining := expiration.Unix() - time.Now().Unix()
		if secondsRemaining > 0 {
			fmt.Printf("Credentials for profile %q still valid for %d seconds until %s\n", config.Profile, secondsRemaining, expirationUnparsed)
			return true
		}
	}

	return false
}

func updateAWSCredentials(config *types.Config, cfg *ini.File, profile *cmd.Profile) error {
	if err := cfg.Section(config.Profile).ReflectFrom(profile); err != nil {
		return err
	}

	return cfg.SaveTo(config.CredentialFile)
}

func validate(config *types.Config, cfg *ini.File) error {
	_, err := cfg.GetSection(config.Profile + longTermSuffix)
	if err != nil {
		return fmt.Errorf("profile %s does not have long-term suffix", config.Profile)
	}

	return nil
}
