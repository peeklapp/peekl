package commands

import (
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/database"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	// Parameters for create command
	createEnrollCmd.Flags().StringP("ip", "", "", "IP address to which the enrollment token should be bound")
	if err := createEnrollCmd.MarkFlagRequired("ip"); err != nil {
		logrus.Fatal(err)
	}
	createEnrollCmd.Flags().StringP("duration", "", "30m", "Duration for which the enrollment token should be valid for")

	// Parameters for expire command
	deleteEnrollCmd.Flags().StringP("ip", "", "", "IP address of the enrollment token you want to delete")
	if err := deleteEnrollCmd.MarkFlagRequired("ip"); err != nil {
		logrus.Fatal(err)
	}

	// Parameters for expire command
	createEnrollCmd.Flags().StringP("token", "", "", "Enrollment token to expire immediately")

	// Add sub-commands to main command
	EnrollCmd.AddCommand(createEnrollCmd)
	EnrollCmd.AddCommand(deleteEnrollCmd)
	EnrollCmd.AddCommand(listEnrollCmd)
}

var EnrollCmd = &cobra.Command{
	Use:   "enroll",
	Short: "Enrollment commands",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				logrus.Fatal(err)
			}
			os.Exit(0)
		}
	},
}

var createEnrollCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new enrollment token",
	Run: func(cmd *cobra.Command, args []string) {
		// Get verbosity
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			logrus.Fatal(err)
		}
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// Load configuration
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			logrus.Fatal(err)
		}
		configStruct, err := config.NewServerConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		// Create db engine
		dbEngine, err := database.NewDatabaseEngine(&configStruct.Database)
		if err != nil {
			logrus.Fatal(err)
		}

		// Get IP
		ip, err := cmd.Flags().GetString("ip")
		if err != nil {
			logrus.Fatal(err)
		}

		// Try to find if an enrollment token already exist
		tokenAlreadyExist, err := dbEngine.DoesATokenExistAndIsValid(ip)
		if err != nil {
			logrus.Fatal(err)
		}
		if tokenAlreadyExist {
			logrus.Fatal("A token already exist and is valid for this IP")
		}

		// Get duration
		duration, err := cmd.Flags().GetString("duration")
		if err != nil {
			logrus.Fatal(err)
		}
		timeDuration, err := time.ParseDuration(duration)
		if err != nil {
			logrus.Fatal(err)
		}
		validUntil := time.Now().Add(timeDuration)

		// Generate token
		tokenUuid := uuid.New().String()

		// Hash token
		hashingParams := utils.DefaultParams()
		hashedToken, err := utils.HashPassword(tokenUuid, hashingParams)
		if err != nil {
			logrus.Fatal(err)
		}

		// Compute valid until value
		if err := dbEngine.InsertEnrollmentToken(hashedToken, ip, validUntil); err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("Enrollment token generated : %s", tokenUuid)
	},
}

var deleteEnrollCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an enrollment token",
	Run: func(cmd *cobra.Command, args []string) {
		// Get verbosity
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			logrus.Fatal(err)
		}
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// Load configuration
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			logrus.Fatal(err)
		}
		configStruct, err := config.NewServerConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		// Get IP
		ip, err := cmd.Flags().GetString("ip")
		if err != nil {
			logrus.Fatal(err)
		}

		dbEngine, err := database.NewDatabaseEngine(&configStruct.Database)
		if err != nil {
			logrus.Fatal(err)
		}

		if err = dbEngine.DeleteEnrollmentToken(ip); err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("Expired enrollment for IP : %s", ip)
	},
}

var listEnrollCmd = &cobra.Command{
	Use:   "list",
	Short: "List enrollment tokens",
	Run: func(cmd *cobra.Command, args []string) {
		// Get verbosity
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			logrus.Fatal(err)
		}
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// Load configuration
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			logrus.Fatal(err)
		}
		configStruct, err := config.NewServerConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		dbEngine, err := database.NewDatabaseEngine(&configStruct.Database)
		if err != nil {
			logrus.Fatal(err)
		}

		enrollmentTokens, err := dbEngine.ListEnrollmentToken()
		if err != nil {
			logrus.Fatal(err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"IP", "Created At", "Valid Until"})
		for _, e := range enrollmentTokens {
			if err := table.Append([]string{e.Ip, e.CreatedAt.String(), e.ValidUntil.String()}); err != nil {
				logrus.Fatal(err)
			}
		}
		if err := table.Render(); err != nil {
			logrus.Fatal(err)
		}
	},
}
