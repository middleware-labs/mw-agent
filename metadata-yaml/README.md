# Auto Metadata YAML Generator

Metadata YAML Generator will use to create the metadata yaml for the metrics from the json exported data.

## Steps to run and configuration

### Requirements

- GO-Lang 1.20
- Python 3

### Configuration

- Configure the fileexporter as mentioned below
  ```yaml
  exporters:
    file/json:
      path: ./metadata-yaml/input/todo/example.json
  service:
    pipelines:
      metrics:
        exporters:
        - file/json
  ```
- Configure integration receiver only at a time to collect the data effectively.
- Keep JSON one entry at a time.
- Set up your integration.
- Build and Run the mw-agent manually with below command
  ```shell
  make build-linux
  
  export MW_API_URL_FOR_CONFIG_CHECK="<https://your.target.endpoint.url>" 
  export MW_API_KEY=<Your Api Key> 
  export MW_TARGET=<https://your.target.endpoint.url>:443
  export MW_FETCH_ACCOUNT_OTEL_CONFIG=false 
  export MW_CONFIG_CHECK_INTERVAL=0 
  build/mw-host-agent start --otel-config-file otel-config.yaml
  ```
- It will generate the metrics in provided file location.

### Run the metadata yaml generator

- Navigate to `metadata-yaml` directory
  ```shell
  cd metadata-yaml
  ```
- Run below command to generate metadata yaml.
  ```shell
  python3 main.py
  ```
- After run successfully all the successfully processed json will be moved to the `completed` directory.
- All the metadata yamls will be in the `metadata-yaml/output` directory.
  - Example:
    - `metadata-yaml/output/example_metadata.yaml`
  - Here already added two json files in todo directory. To try it just run the metadata yaml generator.
