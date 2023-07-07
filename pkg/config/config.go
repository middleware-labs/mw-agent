package config

type Config struct {
	MWApiKey     string
	MWBackendURL string

	EnableSytheticMonitoring bool
	ConfigCheckInterval      string

	ApiURLForConfigCheck string

	EnableDebuggingLogging bool
	EnableInfoLogging      bool
	EnableWarningLogging   bool
	EnableErrorLogging     bool
}
