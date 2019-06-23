package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/mmatur/aws-mfa/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ini "gopkg.in/ini.v1"
)

const (
	// CredentialsFile is the default path for the AWS credentials.
	credentialsFile = "/.aws/credentials"
	longTermSuffix  = "-long-term"
)

var (
	credentialFile string
	duration       int64
	profile        string
	force          bool
	debug          bool
)

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	rootCmd.Flags().StringVar(&credentialFile, "credential-file", path.Join(usr.HomeDir, credentialsFile), "Credential file.")
	rootCmd.Flags().Int64Var(&duration, "duration", 43200, "Duration in seconds for credentials to remain valid.")
	rootCmd.Flags().StringVar(&profile, "profile", "default", "AWS profile to use.")
	rootCmd.Flags().BoolVar(&force, "force", false, "Force credentials renew.")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode.")
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "aws-mfa",
	Short:   "AWS - MFA",
	Long:    "AWS - MFA",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		if err := rootRun(); err != nil {
			log.Fatal(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func rootRun() error {
	cfg, err := ini.Load(credentialFile)
	if err != nil {
		return err
	}

	if err = validate(cfg); err != nil {
		return err
	}

	if credentialStillValid(cfg) {
		return nil
	}

	awsConfig, err := external.LoadDefaultAWSConfig(profile + longTermSuffix)
	if err != nil {
		return err
	}

	var devices []string
	if cfg.Section(profile + longTermSuffix).HasKey("aws_mfa_device") {
		device := cfg.Section(profile + longTermSuffix).Key("aws_mfa_device").String()
		if debug {
			fmt.Printf("MFA device %q found in %q for profile %q\n", device, credentialFile, profile+longTermSuffix)
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

	profile, err := internal.GetSessionToken(awsConfig, duration, answer.Device, answer.Code)
	if err != nil {
		return err
	}

	if err = updateAWSCredentials(cfg, profile); err != nil {
		return err
	}

	fmt.Printf("Success! Credentials for profile %q valid until %s \n", profile, profile.Expiration)
	return nil
}

func credentialStillValid(cfg *ini.File) bool {
	if cfg.Section(profile).HasKey("expiration") && !force {
		expirationUnparsed := cfg.Section(profile).Key("expiration").String()
		expiration, err := time.Parse("2006-01-02 15:04:05", expirationUnparsed)
		if err != nil {
			log.Fatalf("Unable to parse %s", expirationUnparsed)
		}

		secondsRemaining := expiration.Unix() - time.Now().Unix()
		if secondsRemaining > 0 {
			fmt.Printf("Credentials for profile %q still valid for %d seconds until %s\n", profile, secondsRemaining, expirationUnparsed)
			return true
		}
	}

	return false
}

func updateAWSCredentials(cfg *ini.File, p *internal.Profile) error {
	if err := cfg.Section(profile).ReflectFrom(p); err != nil {
		return err
	}

	return cfg.SaveTo(credentialFile)
}

func validate(cfg *ini.File) error {
	if _, err := cfg.GetSection(profile + longTermSuffix); err != nil {
		return fmt.Errorf("profile %s does not have long-term suffix", profile)
	}

	return nil
}
