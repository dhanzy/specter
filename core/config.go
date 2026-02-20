package core

type Config struct {
	MaxDepth         int      `yaml:"max_depth"`
	Concurrency      int      `yaml:"concurrency"`
	QueueSize        int      `yaml:"queue_size"`
	PluginDir        string   `yaml:"plugin_dir"`
	UserAgent        string   `yaml:"user_agent"`
	HTTPTimeout      int      `yaml:"http_timeout"`
	BlacklistDomains []string `yaml:"blacklist_domains"`
}

func NewConfig(configFile string) *Config {
	return &Config{
		MaxDepth:         2,
		Concurrency:      8,
		QueueSize:        256,
		PluginDir:        "plugins",
		UserAgent:        "Specter/1.0",
		BlacklistDomains: []string{"google.com", "facebook.com", "twitter.com", "linkedin.com", "github.com", "instagram.com", "youtube.com", "wikipedia.org", "amazon.com", "netflix.com"},
	}
}
