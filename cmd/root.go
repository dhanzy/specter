package cmd

import (
	"Specter/core"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cobra"
)

var (
	target     string
	pluginDir  string
	configFile string
)

func Execute() error {

	rootCmd := cobra.Command{
		Use:   "specter",
		Short: "Specter vulnerability scanner",
		Long:  "Specter is a vulnerability scanner that detects vulnerabilities in web applications.",
		RunE:  run,
	}

	rootCmd.PersistentFlags().StringVar(&target, "target", "", "Target URL scan (required)")
	rootCmd.PersistentFlags().StringVar(&pluginDir, "plugin", "", "Plugin directory path")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", ".", "Configuration file path")

	rootCmd.MarkPersistentFlagRequired("target")
	return rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	core.Banner()

	cfg := core.NewConfig(configFile)
	httpClient := &http.Client{Timeout: time.Duration(cfg.HTTPTimeout * int(time.Second))}

	targetUrl, err := url.Parse(target)
	if err != nil {
		return err
	}

	crawler := core.NewCrawler(httpClient, core.CrawlerOptions{
		MaxDepth:         cfg.MaxDepth,
		UserAgent:        cfg.UserAgent,
		BlacklistDomains: cfg.BlacklistDomains,
	})

	return crawler.Crawl(targetUrl)
}
