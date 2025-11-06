package lab

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/s-samadi/ghas-lab-builder/internal/config"
	labservice "github.com/s-samadi/ghas-lab-builder/internal/services"
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a full lab environment (org, repos, users)",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		// Traverse up to find and call the root command's PersistentPreRunE
		root := cmd
		for root.Parent() != nil {
			root = root.Parent()
		}

		// Call root's PersistentPreRunE if it exists
		if root.PersistentPreRunE != nil {
			if err := root.PersistentPreRunE(cmd, args); err != nil {
				return err
			}
		}

		// Get context AFTER calling root's PersistentPreRunE (which sets BaseURLKey)
		ctx := cmd.Context()
		ctx = context.WithValue(ctx, config.EnterpriseSlugKey, cmd.Flags().Lookup("enterprise-slug").Value.String())
		ctx = context.WithValue(ctx, config.FacilitatorsKey, strings.Split(facilitators, ","))
		ctx = context.WithValue(ctx, config.LabDateKey, labDate)

		cmd.SetContext(ctx)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

		return labservice.DestroyLabEnvironment(ctx, logger, labDate, usersFile)
	},
}
