package agent

func updateConfigForECS(config map[string]interface{}, infraPlatform InfraPlatform) (map[string]interface{}, error) {

	receiverData, ok := config["receivers"].(map[string]interface{})
	if !ok {
		return nil, ErrParse
	}

	receiverData["awsecscontainermetrics"] = map[string]interface{}{}

	serviceData, ok := config["service"].(map[string]interface{})
	if !ok {
		return nil, ErrParse
	}

	pipelinesData, ok := serviceData["pipelines"].(map[string]interface{})
	if !ok {
		return nil, ErrParse
	}

	metricsData, ok := pipelinesData["metrics"].(map[string]interface{})
	if !ok {
		return nil, ErrParse
	}

	receiversData := []string{}

	if infraPlatform == InfraPlatformECSEC2 {
		for _, receiver := range metricsData["receivers"].([]interface{}) {
			receiverName := receiver.(string)
			receiversData = append(receiversData, receiverName)
		}
	}

	receiversData = append(receiversData, "awsecscontainermetrics")
	metricsData["receivers"] = receiversData

	return config, nil
}
