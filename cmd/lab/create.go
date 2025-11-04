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

var (
	repos             string
	templateReposFile string
	facilitators      string
)

func init() {

	CreateCmd.PersistentFlags().StringVar(&templateReposFile, "template-repos", "", "Path to template repositories file (JSON) (required)")
	CreateCmd.MarkPersistentFlagRequired("template-repos")

	CreateCmd.PersistentFlags().StringVar(&facilitators, "facilitators", "", "lab facilitators usernames, comma-separated")
	CreateCmd.MarkPersistentFlagRequired("facilitators")
}

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a full lab environment (org, repos, users)",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		tokens := strings.Split(cmd.Flags().Lookup("token").Value.String(), ",")
		ctx = context.WithValue(ctx, config.TokenKey, tokens)
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

		return labservice.CreateLabEnvironment(ctx, logger, usersFile, templateReposFile)
	},
}
