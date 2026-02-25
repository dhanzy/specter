package cmd

import (
	"Specter/core"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	target       string
	pluginDir    string
	configFile   string
	activePlugin string
)

func Execute() error {

	rootCmd := cobra.Command{
		Use:   "specter",
		Short: "Specter vulnerability scanner",
		Long:  "Specter is a vulnerability scanner that detects vulnerabilities in web applications.",
		RunE:  run,
	}

	rootCmd.PersistentFlags().StringVar(&target, "target", "", "Target URL scan (required)")
	rootCmd.PersistentFlags().StringVar(&pluginDir, "plugin-dir", "plugins", "Plugin directory path")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", ".", "Configuration file path")
	rootCmd.PersistentFlags().StringVarP(&activePlugin, "plugin", "p", "", "Plugin to use from")

	rootCmd.MarkPersistentFlagRequired("target")
	rootCmd.MarkPersistentFlagRequired("plugin")
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

	pluginpath := path.Join(pluginDir, activePlugin)
	if ok := strings.HasPrefix(pluginpath, ".yaml"); !ok {
		pluginpath = pluginpath + ".yaml"
	}
	info, err := os.Stat(pluginpath)
	if os.IsNotExist(err) {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("Plugin should be a file")
	}

	crawler := core.NewCrawler(httpClient, core.CrawlerOptions{
		MaxDepth:         cfg.MaxDepth,
		UserAgent:        cfg.UserAgent,
		BlacklistDomains: cfg.BlacklistDomains,
		QueueSize:        cfg.QueueSize,
		Workers:          cfg.Concurrency,
	})

	seed := core.Target{URL: targetUrl}

	// targets := make([]core.DetectionResult, 0)

	targetsCh := make(chan core.DetectionResult, cfg.QueueSize)

	go func() {
		if err := crawler.Crawl([]core.Target{seed}, targetsCh); err != nil {
			fmt.Printf("Crawlser stopped error %v\n", err)
		}
	}()

	engine, err := core.NewPluginEngine(pluginpath, cfg)
	if err != nil {
		return err
	}

	for target := range targetsCh {
		engine.Execute(target)
	}
	return nil
}
