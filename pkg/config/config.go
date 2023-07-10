package config

type Config struct {
	MWApiKey string
	MWTarget string

	EnableSytheticMonitoring bool
	ConfigCheckInterval      string

	ApiURLForConfigCheck string

	EnableDebuggingLogging bool
	EnableInfoLogging      bool
	EnableWarningLogging   bool
	EnableErrorLogging     bool
}
