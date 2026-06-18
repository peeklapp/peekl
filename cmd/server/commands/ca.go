package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/peeklapp/peekl/internal/certs"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/database"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	// List commands
	caListPendingCmd.Flags().Bool("json", false, "Show output as JSON")
	caListSignedCmd.Flags().Bool("json", false, "Show output as JSON")
	caListCmd.AddCommand(caListPendingCmd)
	caListCmd.AddCommand(caListSignedCmd)

	// Main command
	CaCmd.AddCommand(caListCmd)
	CaCmd.AddCommand(caSignPendingCmd)
	CaCmd.AddCommand(caRevokeCertCmd)

	// Sign command
	caSignPendingCmd.Flags().StringP("certname", "", "", "Name of the pending certificate to sign")

	// Revoke command
	caRevokeCertCmd.Flags().StringP("certname", "", "", "Name of the certificate to revoke.")
}

var CaCmd = &cobra.Command{
	Use:   "ca",
	Short: "Manipulate the CA",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

var caListCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificates in the CA",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

var caListPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List pending certificates",
	Run: func(cmd *cobra.Command, args []string) {
		// Get verbosity
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			logrus.Fatal(err)
		}
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// Output mode
		jsonOutput, err := cmd.Flags().GetBool("json")
		if err != nil {
			logrus.Fatal(err)
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

		// Get from database
		pendings, err := dbEngine.ListPendingCertificates()
		if err != nil {
			logrus.Fatal(err)
		}

		// Output as JSON if asked to
		if jsonOutput {
			jsonData, err := json.Marshal(pendings)
			if err != nil {
				logrus.Fatal(err)
			}
			fmt.Println(string(jsonData))
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.Header([]string{"Name", "Submission Date"})
			for _, p := range pendings {
				table.Append([]string{p.NodeName, p.SubmittedAt.String()})
			}
			table.Render()
		}
	},
}

var caListSignedCmd = &cobra.Command{
	Use:   "signed",
	Short: "List signed certificates",
	Run: func(cmd *cobra.Command, args []string) {
		// Get verbosity
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			logrus.Fatal(err)
		}
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// Output mode
		jsonOutput, err := cmd.Flags().GetBool("json")
		if err != nil {
			logrus.Fatal(err)
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

		// Get from database
		signeds, err := dbEngine.ListSignedCertificates()
		if err != nil {
			logrus.Fatal(err)
		}

		// Output as JSON if asked to
		if jsonOutput {
			jsonData, err := json.Marshal(signeds)
			if err != nil {
				logrus.Fatal(err)
			}
			fmt.Println(string(jsonData))
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.Header([]string{"Name", "Signature Date"})
			for _, p := range signeds {
				table.Append([]string{p.NodeName, p.SignedAt.String()})
			}
			table.Render()
		}

	},
}

var caSignPendingCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign a pending certificate",
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

		certname, err := cmd.Flags().GetString("certname")
		if err != nil {
			logrus.Fatal(err)
		}
		if certname == "" {
			logrus.Fatal("You must specify a certname using the `--certname` parameter")
		}

		dbEngine, err := database.NewDatabaseEngine(&configStruct.Database)
		if err != nil {
			logrus.Fatal(err)
		}

		// Get pending certificate
		pendingCert, err := dbEngine.GetPendingCertificate(certname)
		if err != nil {
			if errors.Is(err, models.PendingCertificateNotFound{}) {
				logrus.Fatalf("No currently pending certificate for given certname %s", certname)
			} else {
				logrus.Fatal(err)
			}
		}

		cert, err := certs.SignCertificateSigningRequest(
			pendingCert.Data,
			configStruct.Certificates.CaCertificateFilePath,
			configStruct.Certificates.CaCertificateKeyPath,
		)
		if err != nil {
			logrus.Fatal(err)
		}

		// Create entry for signed certificate
		csrSignature := certs.GetCertificateSigningRequestSignature(pendingCert.Data)
		err = dbEngine.InsertSignedCertificate(certname, csrSignature, cert)
		if err != nil {
			logrus.Fatal(err)
		}

		// Delete signature from database as it is know considered signed
		err = dbEngine.DeletePendingCertificate(certname)
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Info(fmt.Sprintf("Signed certificate for '%s'", certname))
	},
}

var caRevokeCertCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke a certificate",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Fatal("Command not implemented yet.")
	},
}
