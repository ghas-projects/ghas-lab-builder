package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/s-samadi/ghas-lab-builder/cmd/lab"
	"github.com/s-samadi/ghas-lab-builder/cmd/orgs"
	"github.com/s-samadi/ghas-lab-builder/cmd/repo"
	"github.com/s-samadi/ghas-lab-builder/internal/config"
	"github.com/spf13/cobra"
)

var (
	appId          string
	privateKeyPath string
	token          string
	baseURL        string
	enterpriseSlug string
)

var rootCmd = &cobra.Command{
	Use:   "ghas-lab-builder",
	Short: "Builds GitHub Advanced Security Lab environments(orgs, repos, users)",
	Long: `ghas-lab-builder is a CLI tool that helps you set up GitHub Advanced Security Lab environments by 
          automating the creation of organizations, repositories, and addings  users required for hands-on labs.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate that either token OR (app-id + private-key-path) is provided, but not both
		hasToken := token != ""
		hasAppCreds := appId != "" || privateKeyPath != ""

		if !hasToken && !hasAppCreds {
			return fmt.Errorf("authentication required: provide either --token OR both --app-id and --private-key-path")
		}

		if hasToken && hasAppCreds {
			return fmt.Errorf("conflicting authentication methods: provide either --token OR (--app-id and --private-key-path), not both")
		}

		// If using app credentials, both app-id and private-key-path must be provided
		if hasAppCreds {
			if appId == "" {
				return fmt.Errorf("--app-id is required when using GitHub App authentication")
			}
			if privateKeyPath == "" {
				return fmt.Errorf("--private-key-path is required when using GitHub App authentication")
			}
		}

		// Set default base URL if not provided
		if baseURL == "" {
			baseURL = config.DefaultBaseURL
		}

		// Store authentication information in context
		ctx := cmd.Context()
		if token != "" {
			// Using PAT authentication
			ctx = context.WithValue(ctx, config.TokenKey, token)
		} else {
			// Using GitHub App authentication
			ctx = context.WithValue(ctx, config.AppIDKey, appId)
			ctx = context.WithValue(ctx, config.PrivateKeyPathKey, privateKeyPath)
		}

		ctx = context.WithValue(ctx, config.BaseURLKey, baseURL)
		ctx = context.WithValue(ctx, config.EnterpriseSlugKey, enterpriseSlug)

		cmd.SetContext(ctx)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// GitHub App authentication flags
	rootCmd.PersistentFlags().StringVar(&appId, "app-id", "", "GitHub App ID (required if not using --token)")
	rootCmd.PersistentFlags().StringVar(&privateKeyPath, "private-key-path", "", "Path to GitHub App private key PEM file (required if not using --token)")

	// PAT authentication flag
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "GitHub Personal Access Token (required if not using GitHub App authentication)")

	// Common flags
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "", "GitHub API base URL")
	rootCmd.PersistentFlags().StringVar(&enterpriseSlug, "enterprise-slug", "", "GitHub Enterprise slug")
	rootCmd.MarkPersistentFlagRequired("enterprise-slug")

	if baseURL == "" {
		baseURL = config.DefaultBaseURL
	}

	rootCmd.AddCommand(lab.LabCmd)
	rootCmd.AddCommand(repo.RepoCmd)
	rootCmd.AddCommand(orgs.OrgsCmd)
}
