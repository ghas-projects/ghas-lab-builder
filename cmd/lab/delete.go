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
		ctx := cmd.Context()

		ctx = context.WithValue(ctx, config.TokenKey, strings.Split(cmd.Flags().Lookup("token").Value.String(), ","))
		ctx = context.WithValue(ctx, config.BaseURLKey, cmd.Flags().Lookup("base-url").Value.String())
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
