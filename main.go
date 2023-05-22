package main

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"github.com/spf13/cobra"
	"github.com/joho/godotenv"

	"github.com/lookatitude/go-orgs-repo-clone/ui"
	"github.com/lookatitude/go-orgs-repo-clone/repo"
)

var (
	clonePath string
	compress  bool
	org       string
	token     string
)

var rootCmd = &cobra.Command{
	Use:   "github-org-cloner",
	Short: "github-org-cloner is a tool for cloning and compressing all repos of a GitHub organization",
	Long:  `A fast and efficient tool for cloning and compressing all repos of a GitHub organization, designed with love.`,
	Run: func(cmd *cobra.Command, args []string) {
		// load values from .env
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file")
		}

		if org == "" {
			org = os.Getenv("GITHUB_ORGANIZATION")
		}

		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		if org == "" || token == "" {
			log.Fatalf("GITHUB_ORGANIZATION and GITHUB_TOKEN must be set")
		}

		ui.Start()
		defer ui.Stop()

		ui.PrintMessage("Fetching repos...")

		repos, err := repo.GetAllRepos(org, token)
		if err != nil {
			ui.PrintError(err)
		}

		ui.PrintMessage("Cloning and compressing repos...")

		var progress int64
		err = repo.CloneAndCompressRepos(repos, token, clonePath, compress, func() {
			atomic.AddInt64(&progress, 1)
			ui.UpdateProgress(int(progress * 100 / int64(len(repos))))
		})
		if err != nil {
			ui.PrintError(err)
		}

		ui.PrintMessage("Done!")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&clonePath, "path", "", "path to clone the repos into")
	rootCmd.PersistentFlags().BoolVar(&compress, "compress", false, "whether to compress the cloned repos")
	rootCmd.PersistentFlags().StringVar(&org, "org", "", "GitHub organization to clone repos from")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "GitHub token for authentication")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
