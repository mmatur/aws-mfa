package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/mmatur/aws-mfa/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"gopkg.in/ini.v1"
)

var Version = "dev"
var ShortCommit = ""
var Date = ""

type rootConfig struct {
	credentialFile string
	duration       int64
	profile        string
	force          bool
	debug          bool
}

const (
	// CredentialsFile is the default path for the AWS credentials.
	credentialsFile = "/.aws/credentials"
	longTermSuffix  = "-long-term"
)

func main() {
	log.SetFlags(log.Lshortfile)

	var rootCfg rootConfig
	var cfg *ini.File

	rootCmd := &cobra.Command{
		Use:     "aws-mfa",
		Short:   "AWS - MFA",
		Long:    "AWS - MFA",
		Version: Version,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("AWS-MFA: %s - %s - %s\n", Version, Date, ShortCommit)

			var err error
			cfg, err = ini.Load(rootCfg.credentialFile)
			if err != nil {
				return err
			}

			if err = validate(rootCfg, cfg); err != nil {
				return err
			}

			if credentialStillValid(rootCfg, cfg) {
				return nil
			}

			err = os.Setenv(external.AWSProfileEnvVar, rootCfg.profile+longTermSuffix)
			if err != nil {
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := rootRun(rootCfg, cfg); err != nil {
				log.Fatal(err)
			}
		},
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	rootCmd.Flags().StringVar(&rootCfg.credentialFile, "credential-file", path.Join(usr.HomeDir, credentialsFile), "Credential file.")
	rootCmd.Flags().Int64Var(&rootCfg.duration, "duration", 43200, "Duration in seconds for credentials to remain valid.")
	rootCmd.Flags().StringVar(&rootCfg.profile, "profile", "default", "AWS profile to use.")
	rootCmd.Flags().BoolVar(&rootCfg.force, "force", false, "Force credentials renew.")
	rootCmd.Flags().BoolVar(&rootCfg.debug, "debug", false, "Enable debug mode.")

	docCmd := &cobra.Command{
		Use:    "doc",
		Short:  "Generate documentation",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doc.GenMarkdownTree(rootCmd, "./docs")
		},
	}

	rootCmd.AddCommand(docCmd)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version",
		Run: func(_ *cobra.Command, _ []string) {
			displayVersion(rootCmd.Name())
		},
	}

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func displayVersion(name string) {
	fmt.Printf(name+`:
 version     : %s
 commit      : %s
 build date  : %s
 go version  : %s
 go compiler : %s
 platform    : %s/%s
`, Version, ShortCommit, Date, runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH)
}

func rootRun(rootCfg rootConfig, cfg *ini.File) error {
	awsConfig, err := external.LoadDefaultAWSConfig(external.DefaultSharedCredentialsFilename())
	if err != nil {
		return err
	}

	var devices []string

	if cfg.Section(rootCfg.profile + longTermSuffix).HasKey("aws_mfa_device") {
		device := cfg.Section(rootCfg.profile + longTermSuffix).Key("aws_mfa_device").String()

		if rootCfg.debug {
			fmt.Printf("MFA device %q found in %q for profile %q\n", device, rootCfg.credentialFile, rootCfg.profile+longTermSuffix)
		}

		devices = append(devices, device)
	} else {
		devices, err = internal.ListMFADevices(awsConfig)
		if err != nil {
			return err
		}
	}

	answer, err := internal.PromptSurvey(devices)
	if err != nil {
		return err
	}

	p, err := internal.GetSessionToken(awsConfig, rootCfg.duration, answer.Device, answer.Code)
	if err != nil {
		return err
	}

	if err = updateAWSCredentials(rootCfg, cfg, p); err != nil {
		return err
	}

	fmt.Printf("Success! Credentials for profile %q valid until %s \n", rootCfg.profile, p.Expiration)

	return nil
}

func credentialStillValid(rootCfg rootConfig, cfg *ini.File) bool {
	if cfg.Section(rootCfg.profile).HasKey("expiration") && !rootCfg.force {
		expirationUnparsed := cfg.Section(rootCfg.profile).Key("expiration").String()

		expiration, err := time.Parse("2006-01-02 15:04:05", expirationUnparsed)
		if err != nil {
			log.Fatalf("Unable to parse %s", expirationUnparsed)
		}

		secondsRemaining := expiration.Unix() - time.Now().Unix()
		if secondsRemaining > 0 {
			fmt.Printf("Credentials for profile %q still valid for %d seconds until %s\n", rootCfg.profile, secondsRemaining, expirationUnparsed)
			return true
		}
	}

	return false
}

func updateAWSCredentials(rootCfg rootConfig, cfg *ini.File, p *internal.Profile) error {
	if err := cfg.Section(rootCfg.profile).ReflectFrom(p); err != nil {
		return err
	}

	return cfg.SaveTo(rootCfg.credentialFile)
}

func validate(rootCfg rootConfig, cfg *ini.File) error {
	if _, err := cfg.GetSection(rootCfg.profile + longTermSuffix); err != nil {
		return fmt.Errorf("profile %s does not have long-term suffix", rootCfg.profile)
	}

	return nil
}
