package repo

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/s-samadi/ghas-lab-builder/internal/config"
	reposervice "github.com/s-samadi/ghas-lab-builder/internal/services"
	"github.com/spf13/cobra"
)

var (
	repos string
)

func init() {
	CreateCmd.PersistentFlags().StringVar(&repos, "repos", "", "Path to template repositories file (JSON) (required)")
	CreateCmd.MarkPersistentFlagRequired("repos")
}

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create repositories within a lab environment",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		tokens := strings.Split(cmd.Flags().Lookup("token").Value.String(), ",")
		ctx = context.WithValue(ctx, config.TokenKey, tokens)
		ctx = context.WithValue(ctx, config.BaseURLKey, cmd.Flags().Lookup("base-url").Value.String())
		ctx = context.WithValue(ctx, config.OrgKey, org)

		cmd.SetContext(ctx)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

		return reposervice.CreateReposInLabOrg(ctx, logger, repos)
	},
}
