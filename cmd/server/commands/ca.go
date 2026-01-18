package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/redat00/peekl/pkg/certs"
	"github.com/redat00/peekl/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	// Main command
	caCmd.AddCommand(caListPendingCmd)
	caCmd.AddCommand(caSignPendingCmd)
	caCmd.AddCommand(caRevokeCertCmd)

	// List commands
	caListPendingCmd.Flags().Bool("pending", false, "Only list the pending certificates")
	caListPendingCmd.Flags().Bool("signed", false, "Only list the signed certificates")
	caListPendingCmd.Flags().Bool("json", false, "Show output as JSON")

	// Sign command
	caSignPendingCmd.Flags().StringP("certname", "", "", "Name of the pending certificate to sign")

	// Revoke command
	caRevokeCertCmd.Flags().StringP("certname", "", "", "Name of the certificate to revoke.")

	rootCmd.AddCommand(caCmd)
}

var caCmd = &cobra.Command{
	Use:   "ca",
	Short: "Manipulate the CA",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

var caListPendingCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificates",
	Run: func(cmd *cobra.Command, args []string) {
		// Get verbosity
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			logrus.Fatal(err)
		}
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// Handle listing
		onlyPending, err := cmd.Flags().GetBool("pending")
		if err != nil {
			logrus.Fatal(err)
		}
		onlySigned, err := cmd.Flags().GetBool("signed")
		if err != nil {
			logrus.Fatal(err)
		}
		if onlyPending && onlySigned {
			logrus.Fatal("If you want to list all certificates, simply remove flags.")
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

		certsDbEngine, err := certs.NewCertsDatabaseEngine(configStruct.Certificates.DatabasePath)
		if err != nil {
			logrus.Fatal(err)
		}

		if onlyPending {
			// Get from database
			pendings, err := certsDbEngine.ListPendingCertificates()
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
		} else if onlySigned {
			fmt.Println(configStruct.Certificates.SignedDirectory)
			fmt.Println("only listing signed certificates")
		} else {
			fmt.Println(configStruct.Certificates.SignedDirectory)
			fmt.Println(configStruct.Certificates.PendingDirectory)
			fmt.Println("listing all certificates")
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

		certsDbEngine, err := certs.NewCertsDatabaseEngine(configStruct.Certificates.DatabasePath)
		if err != nil {
			logrus.Fatal(err)
		}

		// Get pending certificate
		pendingCert, err := certsDbEngine.GetPendingCertificate(certname)
		if err != nil {
			if errors.Is(err, certs.PendingCertificateNotFound{}) {
				logrus.Fatalf("No currently pending certificate for given certname %s", certname)
			} else {
				logrus.Fatal(err)
			}
		}

		err = certs.SignCertificateSigningRequest(
			configStruct.Certificates.CaDirectory,
			configStruct.Certificates.PendingDirectory,
			configStruct.Certificates.SignedDirectory,
			certname,
		)
		if err != nil {
			logrus.Fatal(err)
		}

		// Create entry for signed certificate
		err = certsDbEngine.InsertSignedCertificate(certname, pendingCert.Signature)
		if err != nil {
			logrus.Fatal(err)
		}

		// Delete the CSR locally
		err = os.Remove(fmt.Sprintf("%s/%s.csr", configStruct.Certificates.PendingDirectory, certname))
		if err != nil {
			logrus.Fatal(err)
		}

		// Delete signature from database as it is know considered signed
		err = certsDbEngine.DeletePendingCertificate(certname)
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

		certsDbEngine, err := certs.NewCertsDatabaseEngine(configStruct.Certificates.DatabasePath)
		if err != nil {
			logrus.Fatal(err)
		}

		err = os.Remove(fmt.Sprintf("%s/%s.crt", configStruct.Certificates.SignedDirectory, certname))
		if err != nil {
			logrus.Fatal(err)
		}

		// TODO: ALL WRONG DON'T BELIEVE IT
		err = certsDbEngine.DeletePendingCertificate(certname)
		if err != nil {
			logrus.Fatal(err)
		}

		// TODO: ADD ACTUAL REVOKATION WITH A CRL
	},
}
