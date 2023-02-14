package main

type OtelStruct struct {
	Receivers struct {
		Otlp struct {
			Protocols struct {
				Grpc struct {
					Endpoint string `yaml:"endpoint"`
				} `yaml:"grpc"`
				HTTP struct {
					Endpoint string `yaml:"endpoint"`
				} `yaml:"http"`
			} `yaml:"protocols"`
		} `yaml:"otlp"`
		Hostmetrics struct {
			CollectionInterval string `yaml:"collection_interval"`
			Scrapers           struct {
				CPU struct {
					Metrics struct {
						SystemCPUUtilization struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"system.cpu.utilization"`
					} `yaml:"metrics"`
				} `yaml:"cpu"`
				Load struct {
					CPUAverage bool `yaml:"cpu_average"`
				} `yaml:"load"`
				Memory struct {
					Metrics struct {
						SystemMemoryUtilization struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"system.memory.utilization"`
					} `yaml:"metrics"`
				} `yaml:"memory"`
				Paging struct {
				} `yaml:"paging"`
				Disk struct {
					Metrics struct {
						SystemDiskIoSpeed struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"system.disk.io.speed"`
					} `yaml:"metrics"`
				} `yaml:"disk"`
				Filesystem struct {
					Metrics struct {
						SystemFilesystemUtilization struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"system.filesystem.utilization"`
					} `yaml:"metrics"`
				} `yaml:"filesystem"`
				Network struct {
					Metrics struct {
						SystemNetworkIoBandwidth struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"system.network.io.bandwidth"`
					} `yaml:"metrics"`
				} `yaml:"network"`
				Processes struct {
				} `yaml:"processes"`
				Process struct {
					AvoidSelectedErrors  bool `yaml:"avoid_selected_errors"`
					MuteProcessNameError bool `yaml:"mute_process_name_error"`
				} `yaml:"process"`
			} `yaml:"scrapers"`
		} `yaml:"hostmetrics"`
		DockerStats struct {
			Endpoint           string  `yaml:"endpoint"`
			CollectionInterval string  `yaml:"collection_interval"`
			Timeout            string  `yaml:"timeout"`
			APIVersion         float64 `yaml:"api_version"`
		} `yaml:"docker_stats"`
		Prometheus struct {
			Config struct {
				ScrapeConfigs []struct {
					JobName        string `yaml:"job_name"`
					ScrapeInterval string `yaml:"scrape_interval"`
					StaticConfigs  []struct {
						Targets []string `yaml:"targets"`
					} `yaml:"static_configs"`
				} `yaml:"scrape_configs"`
			} `yaml:"config"`
		} `yaml:"prometheus"`
		Filelog struct {
			Include                 []string `yaml:"include"`
			Exclude                 []string `yaml:"exclude"`
			IncludeFilePath         bool     `yaml:"include_file_path"`
			IncludeFileNameResolved bool     `yaml:"include_file_name_resolved"`
			IncludeFilePathResolved bool     `yaml:"include_file_path_resolved"`
			Operators               []struct {
				Type      string `yaml:"type"`
				If        string `yaml:"if"`
				ID        string `yaml:"id"`
				Field     string `yaml:"field,omitempty"`
				Value     string `yaml:"value,omitempty"`
				Output    string `yaml:"output,omitempty"`
				Regex     string `yaml:"regex,omitempty"`
				ParseFrom string `yaml:"parse_from,omitempty"`
				From      string `yaml:"from,omitempty"`
				To        string `yaml:"to,omitempty"`
			} `yaml:"operators"`
		} `yaml:"filelog"`
		Fluentforward struct {
			Endpoint string `yaml:"endpoint"`
		} `yaml:"fluentforward"`
	} `yaml:"receivers"`
	Processors struct {
		Resource struct {
			Attributes []struct {
				Key           string `yaml:"key"`
				Action        string `yaml:"action"`
				Value         string `yaml:"value,omitempty"`
				FromAttribute string `yaml:"from_attribute,omitempty"`
			} `yaml:"attributes"`
		} `yaml:"resource"`
		Resource2 struct {
			Attributes []struct {
				Key           string `yaml:"key"`
				Action        string `yaml:"action"`
				Value         string `yaml:"value,omitempty"`
				FromAttribute string `yaml:"from_attribute,omitempty"`
			} `yaml:"attributes"`
		} `yaml:"resource/2"`
		Resource3 struct {
			Attributes []struct {
				Key           string `yaml:"key"`
				Action        string `yaml:"action"`
				Value         string `yaml:"value,omitempty"`
				FromAttribute string `yaml:"from_attribute,omitempty"`
			} `yaml:"attributes"`
		} `yaml:"resource/3"`
		AttributesTraces struct {
			Actions []struct {
				Key           string `yaml:"key"`
				FromAttribute string `yaml:"from_attribute"`
				Action        string `yaml:"action"`
			} `yaml:"actions"`
		} `yaml:"attributes/traces"`
		Resourcedetection struct {
			Detectors []string `yaml:"detectors"`
			Timeout   string   `yaml:"timeout"`
			Override  bool     `yaml:"override"`
		} `yaml:"resourcedetection"`
	} `yaml:"processors"`
	Exporters struct {
		Logging struct {
			Loglevel string `yaml:"loglevel"`
		} `yaml:"logging"`
		Otlp2 struct {
			Endpoint string `yaml:"endpoint"`
			Headers  struct {
				Authorization string `yaml:"authorization"`
			} `yaml:"headers"`
			SendingQueue struct {
				Enabled      bool `yaml:"enabled"`
				NumConsumers int  `yaml:"num_consumers"`
				QueueSize    int  `yaml:"queue_size"`
			} `yaml:"sending_queue"`
		} `yaml:"otlp/2"`
	} `yaml:"exporters"`
	Service struct {
		Pipelines struct {
			Metrics struct {
				Receivers  []string `yaml:"receivers"`
				Processors []string `yaml:"processors"`
				Exporters  []string `yaml:"exporters"`
			} `yaml:"metrics"`
			Logs struct {
				Receivers  []string `yaml:"receivers"`
				Processors []string `yaml:"processors"`
				Exporters  []string `yaml:"exporters"`
			} `yaml:"logs"`
			Traces struct {
				Receivers  []string `yaml:"receivers"`
				Processors []string `yaml:"processors"`
				Exporters  []string `yaml:"exporters"`
			} `yaml:"traces"`
		} `yaml:"pipelines"`
		Telemetry struct {
			Logs struct {
				Level string `yaml:"level"`
			} `yaml:"logs"`
		} `yaml:"telemetry"`
	} `yaml:"service"`
}
