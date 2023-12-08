package agent

func (c *HostAgent) updateConfigForECS(config map[string]interface{}) (map[string]interface{}, error) {

	receiverData, ok := config[Receivers].(map[string]interface{})
	if !ok {
		return nil, ErrParseReceivers
	}

	receiverData[AWSECSContainerMetrics] = map[string]interface{}{}

	serviceData, ok := config[Service].(map[string]interface{})
	if !ok {
		return nil, ErrParseService
	}

	pipelinesData, ok := serviceData[Pipelines].(map[string]interface{})
	if !ok {
		return nil, ErrParsePipelines
	}

	metricsData, ok := pipelinesData[Metrics].(map[string]interface{})
	if !ok {
		return nil, ErrParseMetrics
	}

	receiversData := []string{}

	if c.InfraPlatform == InfraPlatformECSEC2 {
		for _, receiver := range metricsData[Receivers].([]interface{}) {
			receiverName := receiver.(string)
			receiversData = append(receiversData, receiverName)
		}
	}

	receiversData = append(receiversData, AWSECSContainerMetrics)
	metricsData[Receivers] = receiversData

	return config, nil
}
