package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

// DatabaseConfig represents a single database connection configuration.
type DatabaseConfig struct {
	Name    string `json:"name"`
	Driver  string `json:"driver"`
	DSN     string `json:"dsn"`
	Enabled bool   `json:"enabled"`
}

// ElasticsearchConfig represents Elasticsearch connection configuration.
type ElasticsearchConfig struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	APIKey  string `json:"api_key,omitempty"`
	User    string `json:"user,omitempty"`
	Pass    string `json:"pass,omitempty"`
	Enabled bool   `json:"enabled"`
}

// PrometheusConfig represents Prometheus connection configuration.
type PrometheusConfig struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// RedisConfig represents Redis connection configuration.
type RedisConfig struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password,omitempty"`
	DB       int    `json:"db"`
	Enabled  bool   `json:"enabled"`
}

// S3Config represents S3-compatible storage configuration.
type S3Config struct {
	Name            string `json:"name"`
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region,omitempty"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Bucket          string `json:"bucket,omitempty"`
	UseSSL          bool   `json:"use_ssl"`
	Enabled         bool   `json:"enabled"`
}

// DataSourcesConfig holds all datasource configurations from file.
type DataSourcesConfig struct {
	Databases     []DatabaseConfig      `json:"databases,omitempty"`
	Elasticsearch []ElasticsearchConfig `json:"elasticsearch,omitempty"`
	Prometheus    []PrometheusConfig    `json:"prometheus,omitempty"`
	Redis         []RedisConfig         `json:"redis,omitempty"`
	S3            []S3Config            `json:"s3,omitempty"`
}

// Config holds the application configuration.
type Config struct {
	ServerName          string
	ServerVersion       string
	LogLevel            string
	PrometheusAddr      string
	DBDriver            string
	DBDSN               string
	PostgresDSN         string
	MySQLDSN            string
	DataSourcesFile     string // Path to datasources.json (or databases.json for backward compat)
	DataSourcesConfig   *DataSourcesConfig
	ElasticsearchURL    string
	ElasticsearchUser   string
	ElasticsearchPass   string
	ElasticsearchAPIKey string
	Transport           string // "stdio" or "http"
	Host                string // Hostname for HTTP server (default localhost)
	Port                string // Port for HTTP server (default 8080)
	PublicURL           string // Full public URL for HTTP mode (e.g., https://psops.pcsindonesia.com)
}

// Load loads the configuration from environment variables.
func Load() *Config {
	_ = godotenv.Load() // Load .env file if it exists
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" && os.Getenv("APP_POSTGRESQL_HOST") != "" {
		dbDSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			getEnv("APP_POSTGRESQL_HOST", "localhost"),
			getEnv("APP_POSTGRESQL_PORT", "5432"),
			getEnv("APP_POSTGRESQL_USERNAME", "postgres"),
			getEnv("APP_POSTGRESQL_PASSWORD", ""),
			getEnv("APP_POSTGRESQL_DB_NAME", "postgres"),
			getEnv("APP_POSTGRESQL_SSL_MODE", "disable"),
		)
	}

	cfg := &Config{
		ServerName:          getEnv("SERVER_NAME", "Production MCP Server"),
		ServerVersion:       getEnv("SERVER_VERSION", "1.0.0"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		PrometheusAddr:      os.Getenv("PROMETHEUS_ADDR"),
		DBDriver:            getEnv("DB_DRIVER", "postgres"),
		DBDSN:               dbDSN,
		PostgresDSN:         getEnv("POSTGRES_DSN", ""),
		MySQLDSN:            getEnv("MYSQL_DSN", ""),
		DataSourcesFile:     getEnv("DATASOURCES_CONFIG", getEnv("DATABASES_CONFIG", "datasources.json")),
		ElasticsearchURL:    getEnv("ELASTICSEARCH_URL", ""),
		ElasticsearchUser:   getEnv("ELASTICSEARCH_USER", ""),
		ElasticsearchPass:   getEnv("ELASTICSEARCH_PASS", ""),
		ElasticsearchAPIKey: getEnv("ELASTICSEARCH_API_KEY", ""),
		Transport:           getEnv("MCP_TRANSPORT", "stdio"),
		Host:                getEnv("HOST", "localhost"),
		Port:                getEnv("PORT", "8180"),
		PublicURL:           getEnv("PUBLIC_URL", ""),
	}

	// Load datasources from config file if it exists
	cfg.DataSourcesConfig = loadDataSourcesConfig(cfg.DataSourcesFile)

	// Override with file-based Elasticsearch config if available (use first enabled)
	if cfg.DataSourcesConfig != nil && len(cfg.DataSourcesConfig.Elasticsearch) > 0 {
		for _, es := range cfg.DataSourcesConfig.Elasticsearch {
			if !es.Enabled {
				continue
			}
			// Use first enabled ES cluster
			if cfg.ElasticsearchURL == "" {
				cfg.ElasticsearchURL = es.URL
			}
			if cfg.ElasticsearchAPIKey == "" {
				cfg.ElasticsearchAPIKey = es.APIKey
			}
			if cfg.ElasticsearchUser == "" {
				cfg.ElasticsearchUser = es.User
			}
			if cfg.ElasticsearchPass == "" {
				cfg.ElasticsearchPass = es.Pass
			}
			break // Only use first enabled
		}
	}

	// Override with file-based Prometheus config if available (use first enabled)
	if cfg.DataSourcesConfig != nil && len(cfg.DataSourcesConfig.Prometheus) > 0 {
		for _, prom := range cfg.DataSourcesConfig.Prometheus {
			if !prom.Enabled {
				continue
			}
			// Use first enabled Prometheus
			if cfg.PrometheusAddr == "" {
				cfg.PrometheusAddr = prom.URL
			}
			break // Only use first enabled
		}
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		return cast.ToInt(value)
	}
	return defaultValue
}

// loadDataSourcesConfig loads datasource configurations from JSON file.
func loadDataSourcesConfig(filePath string) *DataSourcesConfig {
	data, err := os.ReadFile(filePath)
	if err != nil {
		// File doesn't exist or can't be read - return nil (will use ENV-based config)
		return nil
	}

	var dsConfig DataSourcesConfig
	if err := json.Unmarshal(data, &dsConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to parse %s: %v\n", filePath, err)
		return nil
	}

	return &dsConfig
}
