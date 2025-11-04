package cmd

import (
	"fmt"
	"os"

	"github.com/s-samadi/ghas-lab-builder/cmd/lab"
	"github.com/s-samadi/ghas-lab-builder/cmd/repo"
	"github.com/s-samadi/ghas-lab-builder/internal/config"
	"github.com/spf13/cobra"
)

var (
	token          string
	baseURL        string
	enterpriseSlug string
)

var rootCmd = &cobra.Command{
	Use:   "ghas-lab-builder",
	Short: "Builds GitHub Advanced Security Lab environments(orgs, repos, users)",
	Long: `ghas-lab-builder is a CLI tool that helps you set up GitHub Advanced Security Lab environments by 
          automating the creation of organizations, repositories, and addings  users required for hands-on labs.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "GitHub token(s), comma-separated")
	rootCmd.MarkPersistentFlagRequired("token")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "", "GitHub API base URL")
	rootCmd.PersistentFlags().StringVar(&enterpriseSlug, "enterprise-slug", "", "GitHub Enterprise slug")
	rootCmd.MarkPersistentFlagRequired("enterprise-slug")

	if baseURL == "" {
		baseURL = config.DefaultBaseURL
	}

	rootCmd.AddCommand(lab.LabCmd)
	rootCmd.AddCommand(repo.RepoCmd)

}
