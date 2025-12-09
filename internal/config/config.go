package config

import (
	"log"
	"os"
)

// Config holds all base application configuration settings shared or used by either service.
type Config struct {
	// Server settings (primarily for receiver)
	Port string

	// SaaS Source (GitHub) security
	GitHubSecret string

	// Azure Service Bus settings (used by both)
	ServiceBusConnectionString string
	ServiceBusQueueName      string
    
    // Forwarder specific settings (loaded separately in forwarder main)
    TargetToolURL string
    TargetToolAuthToken string
}

// LoadBaseConfig initializes the configuration from environment variables.
func LoadBaseConfig() *Config {
	// Load common settings
	cfg := &Config{
		Port:                       getEnv("PORT", "8080"),
		ServiceBusConnectionString: getEnvStrict("AZURE_SERVICE_BUS_CONN_STRING"),
		ServiceBusQueueName:        getEnvStrict("AZURE_SERVICE_BUS_QUEUE_NAME"),
        
        // These will be loaded by the specific component that needs them
        GitHubSecret: getEnv("GITHUB_WEBHOOK_SECRET", ""),
        TargetToolURL: getEnv("TARGET_TOOL_URL", ""),
        TargetToolAuthToken: getEnv("TARGET_TOOL_AUTH_TOKEN", ""),
	}
	log.Println("Base configuration loaded successfully.")
	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvStrict(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("Environment variable %s is required and not set.", key)
	return ""
}
