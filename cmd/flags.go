package cmd

import (
	"github.com/spf13/cobra"
)

type Flag string

const (
	Email     = Flag("email")
	Domain    = Flag("domain")
	Overwrite = Flag("overwrite")
)

func addEmailAndDomainFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(string(Email), "e", "", "Email address of your Zattoo account.")
	cmd.MarkFlagRequired(string(Email))

	cmd.Flags().StringP(string(Domain), "d", "zattoo.com", "Domain of your Zattoo subscription.")
}

func addDownloadFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP(string(Overwrite), "y", false, "Overwrite existing files?")
}
