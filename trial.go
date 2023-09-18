package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type PrometheusConfig struct {
	ScrapeConfigs []ScrapeConfig `yaml:"scrape_configs"`
}

type ScrapeConfig struct {
	JobName        string         `yaml:"job_name"`
	ScrapeInterval string         `yaml:"scrape_interval"`
	StaticConfigs  []StaticConfig `yaml:"static_configs"`
}

type StaticConfig struct {
	Targets []string `yaml:"targets"`
}

type ReceiversConfig struct {
	Prometheus PrometheusReceiverConfig `yaml:"prometheus"`
}

type PrometheusReceiverConfig struct {
	Config PrometheusConfig `yaml:"config"`
}

type JobList []JobItem

type JobItem struct {
	Name       string
	TargetHost string
	TargetPort string
}

func main() {
	jobList := CreateJobList()

	path := "./configyamls-k8s/otel-config.yaml"

	// Read the YAML file
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading YAML file:", zap.Error(err))

	}

	// Unmarshal YAML into a map[string]interface{}
	var configMap map[string]interface{}
	err = yaml.Unmarshal(data, &configMap)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML:", zap.Error(err))

	}

	// Check if "receivers" key exists in the configMap
	receiversMap, ok := configMap["receivers"].(map[interface{}]interface{})
	if !ok {
		log.Fatalf("Error: 'receivers' key not found or is not a map")
	}

	// Now, unmarshal the nested map into a ReceiversConfig
	var receivers ReceiversConfig
	receiversBytes, err := yaml.Marshal(receiversMap)
	if err != nil {
		log.Fatalf("Error marshaling receivers map: %v", err)
	}
	if err := yaml.Unmarshal(receiversBytes, &receivers); err != nil {
		log.Fatalf("Error unmarshaling receivers map: %v", err)
	}

	fmt.Println("receiversBytes --->", string(receiversBytes))

	// for _, job := range receivers.Prometheus.Config.ScrapeConfigs

	var scrapeConfigs []ScrapeConfig

	for _, job := range jobList {
		scrapeConfig := ScrapeConfig{
			JobName:        job.Name,
			ScrapeInterval: "5s",
			StaticConfigs: []StaticConfig{
				{
					Targets: []string{
						job.TargetHost + ":" + job.TargetPort,
					},
				},
			},
		}

		scrapeConfigs = append(scrapeConfigs, scrapeConfig)
	}

	receivers.Prometheus.Config.ScrapeConfigs = append(receivers.Prometheus.Config.ScrapeConfigs, scrapeConfigs...)
	fmt.Println("===>", receivers.Prometheus.Config.ScrapeConfigs)

	configMap["receivers"] = receivers

	// Marshal the updated map back to YAML
	updatedYAML, err := yaml.Marshal(&configMap)
	if err != nil {
		log.Fatalf("Error marshaling YAML:", zap.Error(err))
	}

	// Write the updated YAML to a file or print it to the console
	err = os.WriteFile(path, updatedYAML, 0644)
	if err != nil {
		log.Fatalf("Error writing YAML file:", zap.Error(err))
	}
	// // Check if "pipelines" key exists and is a map
	// prometheus, ok := receivers.(ReceiversConfig)
	// if !ok {
	// 	log.Fatalf("Error: 'prometheus' key not found or is not a map")
	// }

	// fmt.Println("prometheus --->", prometheus)

}

func CreateJobList() JobList {
	// Parse the environment variable
	prometheusScrapeConfig := os.Getenv("PROMETHEUS_SCRAPE_CONFIG")

	if !HasValidPrometheusScrapeConfig(prometheusScrapeConfig) {
		log.Fatal("PROMETHEUS_SCRAPE_CONFIG is not valid.")
	}

	if prometheusScrapeConfig == "" {
		fmt.Println("PROMETHEUS_SCRAPE_CONFIG is not set.")
		return nil
	}

	// Split the input string into job configurations
	prometheusScraperJobs := strings.Split(prometheusScrapeConfig, ",")

	// Create a list of jobs
	var jobList JobList

	for _, job := range prometheusScraperJobs {
		// Split each job configuration into key-value pairs
		jobDetails := strings.Split(job, "@")

		target := strings.Split(jobDetails[1], ":")

		jobItem := JobItem{
			Name:       jobDetails[0],
			TargetHost: target[0],
			TargetPort: target[1],
		}

		fmt.Println("jobItem --->", jobItem)

		jobList = append(jobList, jobItem)
	}

	return jobList
}

func HasValidPrometheusScrapeConfig(config string) bool {
	if config == "" {
		return true
	}
	pairs := strings.Split(config, ",")
	for _, pair := range pairs {
		keyValue := strings.Split(pair, ":")
		if len(keyValue) != 2 {
			return false
		}
	}
	return true
}
