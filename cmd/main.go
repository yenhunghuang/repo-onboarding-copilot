package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yenhunghuang/repo-onboarding-copilot/internal/security/validator"
	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

var (
	// Version will be set during build
	Version = "dev"
	// BuildDate will be set during build
	BuildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "repo-onboarding-copilot [repository-url]",
	Short: "Repository onboarding analysis tool",
	Long: `Repo Onboarding Copilot is a CLI tool that analyzes Git repositories
to generate comprehensive onboarding documentation and identify key patterns,
dependencies, and architectural insights.

Examples:
  # Analyze a GitHub repository
  repo-onboarding-copilot https://github.com/owner/repo.git
  
  # Analyze using SSH URL
  repo-onboarding-copilot git@github.com:owner/repo.git
  
  # Show version information
  repo-onboarding-copilot --version`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		// Initialize logger
		log := logger.New()
		
		// Validate repository URL
		validator := validator.New()
		validatedURL, err := validator.ValidateRepositoryURL(args[0])
		if err != nil {
			log.Error(fmt.Sprintf("Invalid repository URL: %v", err))
			os.Exit(1)
		}

		log.Info(fmt.Sprintf("Starting analysis of repository: %s", validatedURL.Raw))
		fmt.Printf("✓ Repository URL validated successfully\n")
		fmt.Printf("✓ Scheme: %s, Host: %s\n", validatedURL.Scheme, validatedURL.Host)
		
		// TODO: Implement repository analysis workflow
		fmt.Printf("Repository analysis workflow will be implemented in future stories.\n")
	},
}

func init() {
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Show version information")
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Show help information")
	
	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Repo Onboarding Copilot %s (built %s)\n", Version, BuildDate)
		},
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}