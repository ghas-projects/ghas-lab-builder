package repo

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/s-samadi/ghas-lab-builder/internal/config"
	reposervice "github.com/s-samadi/ghas-lab-builder/internal/services"
	userspec "github.com/s-samadi/ghas-lab-builder/internal/util"
	"github.com/spf13/cobra"
)

var (
	deleteRepos string
)

func init() {
	DeleteCmd.PersistentFlags().StringVar(&deleteRepos, "repos", "", "Path to file containing repository names to delete (JSON). If empty, all repos in the org will be deleted")
}

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete repositories within a lab environment",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		tokens := strings.Split(cmd.Flags().Lookup("token").Value.String(), ",")
		ctx = context.WithValue(ctx, config.TokenKey, tokens)
		ctx = context.WithValue(ctx, config.OrgKey, org)
		ctx = context.WithValue(ctx, config.BaseURLKey, cmd.Flags().Lookup("base-url").Value.String())

		cmd.SetContext(ctx)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

		var repoNames []string
		var err error

		// If repos file is provided, load from file
		// Otherwise, delete all repos in the org
		if deleteRepos != "" {
			repoNames, err = userspec.LoadFromJsonFile(deleteRepos)
			if err != nil {
				logger.Error("Failed to load repository names",
					slog.String("file", deleteRepos),
					slog.Any("error", err))
				return err
			}
		} else {
			logger.Info("No repos file specified, will delete all repositories in the organization")
			repoNames = nil // nil signals to delete all repos
		}

		return reposervice.DeleteReposInLabOrg(ctx, logger, repoNames)
	},
}
