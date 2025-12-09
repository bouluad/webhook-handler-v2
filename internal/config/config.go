package config

import (
	"log"
	"os"
)

// Config holds all base application configuration settings.
type Config struct {
	// Server settings (primarily for receiver)
	Port string

	// SaaS Source (GitHub) security
	GitHubSecret string

	// Azure Service Bus settings (used by both)
	ServiceBusConnectionString string
	ServiceBusQueueName      string
    
    // Forwarder specific settings
    TargetToolURL string
    TargetToolAuthToken string
}

// LoadBaseConfig initializes the configuration from environment variables.
// Note: It loads all possible variables, and the main function will enforce strict requirements.
func LoadBaseConfig() *Config {
	cfg := &Config{
		Port:                       GetEnv("PORT", "8080"),
		ServiceBusConnectionString: GetEnv("AZURE_SERVICE_BUS_CONN_STRING", ""),
		ServiceBusQueueName:        GetEnv("AZURE_SERVICE_BUS_QUEUE_NAME", ""),
        
        GitHubSecret:               GetEnv("GITHUB_WEBHOOK_SECRET", ""),
        TargetToolURL:              GetEnv("TARGET_TOOL_URL", ""),
        TargetToolAuthToken:        GetEnv("TARGET_TOOL_AUTH_TOKEN", ""),
	}
	log.Println("Base configuration loaded.")
	return cfg
}

// GetEnv retrieves the environment variable or returns the fallback.
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// GetEnvStrict retrieves the environment variable or terminates the application.
func GetEnvStrict(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("Environment variable %s is required and not set.", key)
	return ""
}
