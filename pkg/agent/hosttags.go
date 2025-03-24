package agent

func (c *HostAgent) updateConfigForHostTags(config map[string]interface{}) (map[string]interface{}, error) {

	processorsData, ok := config[Processors].(map[string]interface{})
	if !ok {
		return nil, ErrParseProcessors
	}

	processorsData["resource/host_tags"] = map[string]interface{}{
		"attributes": []map[string]interface{}{
			{
				"key":    "mw.host.tags",
				"action": "insert",
				"value":  c.HostTags,
			},
		},
	}

	serviceData, ok := config[Service].(map[string]interface{})
	if !ok {
		return nil, ErrParseService
	}

	pipelinesData, ok := serviceData[Pipelines].(map[string]interface{})
	if !ok {
		return nil, ErrParsePipelines
	}

	// Iterate over each pipeline and add the processor
	for pipelineName, pipeline := range pipelinesData {
		pipelineMap, ok := pipeline.(map[string]interface{})
		if !ok {
			continue
		}

		// Get existing processors
		processors, exists := pipelineMap["processors"].([]interface{})
		if !exists {
			// If processors do not exist, create a new slice
			pipelineMap["processors"] = []interface{}{"resource/host_tags"}
		} else {
			// Append "resource/host_tags" only if it's not already present
			found := false
			for _, p := range processors {
				if p == "resource/host_tags" {
					found = true
					break
				}
			}
			if !found {
				pipelineMap["processors"] = append([]interface{}{"resource/host_tags"}, processors...)
			}
		}

		// Update the pipeline back
		pipelinesData[pipelineName] = pipelineMap
	}

	return config, nil
}
