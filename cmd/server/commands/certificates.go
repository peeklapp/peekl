package commands

import (
	"errors"
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
	caListCmd.AddCommand(caListSignedCmd)
	caListCmd.AddCommand(caListRevokedCmd)

	CaCmd.AddCommand(caListCmd)
	CaCmd.AddCommand(caRevokeCertCmd)

	caRevokeCertCmd.Flags().StringP("certname", "", "", "Name of the certificate to revoke.")
}

var CaCmd = &cobra.Command{
	Use:   "certificates",
	Short: "Commands to interact with the certificates created by Peekl",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				logrus.Fatal(err)
			}
			os.Exit(0)
		}
	},
}

var caListCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificates in the CA",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				logrus.Fatal(err)
			}
			os.Exit(0)
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

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"Name", "Signature Date"})
		for _, p := range signeds {
			if err := table.Append([]string{p.NodeName, p.SignedAt.String()}); err != nil {
				logrus.Fatal(err)
			}
		}
		if err := table.Render(); err != nil {
			logrus.Fatal(err)
		}
	},
}

var caListRevokedCmd = &cobra.Command{
	Use:   "revoked",
	Short: "List revoked certificates",
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

		// Get from database
		revokedCerts, err := dbEngine.ListRevokedCertificates()
		if err != nil {
			logrus.Fatal(err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"Name", "Serial number", "Revocation Date"})
		for _, p := range revokedCerts {
			if err := table.Append([]string{p.NodeName, p.SerialNumber, p.RevokedAt.String()}); err != nil {
				logrus.Fatal(err)
			}
		}
		if err := table.Render(); err != nil {
			logrus.Fatal(err)
		}
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

		dbEngine, err := database.NewDatabaseEngine(&configStruct.Database)
		if err != nil {
			logrus.Fatal(err)
		}

		dbSignedCert, err := dbEngine.GetSignedCertificateByNodeName(certname)
		if err != nil {
			if errors.Is(err, models.SignedCertificateNotFoundByNodeName{}) {
				logrus.Fatalf("No found signed certificate for given certname %s", certname)
			} else {
				logrus.Fatal(err)
			}
		}

		// Load the certificate
		signedCert, err := certs.LoadCertificateFromData(dbSignedCert.Certificate)
		if err != nil {
			logrus.Fatal(err)
		}

		err = dbEngine.InsertRevokedCertificate(dbSignedCert.NodeName, signedCert.SerialNumber.String())
		if err != nil {
			logrus.Fatal(err)
		}

		err = dbEngine.DeleteSignedCertificate(certname)
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("Revoked certificate for '%s'", certname)
	},
}
