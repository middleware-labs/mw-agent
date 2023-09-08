
# Middleware Agent (mw-agent) Configuration Guide

The Middleware Agent (`mw-agent`) is a versatile host agent designed to collect observability signals. This guide will walk you through the configuration options for the `mw-agent` using input flags and environment variables.

## Input Flags

Input flags are command-line arguments that you can use to configure the `mw-agent` when starting it. Here are the available input flags:

1. `--api-key` (Environment Variable: `MW_API_KEY`):
   - Description: Middleware API key for your account.
   - Example: `--api-key=<YOUR_API_KEY>`

2. `--target` (Environment Variable: `MW_TARGET`):
   - Description: Middleware target for your account.
   - Example: `--target=https://app.middleware.io`

3. `--enable-synthetic-monitoring` (Environment Variable: `MW_ENABLE_SYNTHETIC_MONITORING`):
   - Description: Enable synthetic monitoring.
   - Example: `--enable-synthetic-monitoring`

4. `--config-check-interval` (Environment Variable: `MW_CONFIG_CHECK_INTERVAL`):
   - Description: Duration string to periodically check for configuration updates. Setting the value to `0` disables this feature.
   - Example: `--config-check-interval=60s`

5. `--docker-endpoint` (Environment Variable: `MW_DOCKER_ENDPOINT`):
   - Description: Set the endpoint for the Docker socket if different from the default.
   - Example: `--docker-endpoint=unix:///var/run/docker.sock`

6. `--host-tags` (Environment Variable: `MW_HOST_TAGS`):
   - Description: Tags for this host.
   - Example: `--host-tags=tag1=value1,tag2=value2`

7. `--logfile` (Environment Variable: `MW_LOGFILE`):
   - Description: Log file to store Middleware agent logs.
   - Example: `--logfile=/path/to/logfile.log`

8. `--logfile-size` (Environment Variable: `MW_LOGFILE_SIZE`):
   - Description: Log file size to store Middleware agent logs. This flag only applies if `--logfile` flag is specified.
   - Example: `--logfile-size=10` (for 10 MB log file)

9. `--config-file` (Environment Variable: `MW_CONFIG_FILE`):
   - Description: Location of the configuration file for this agent. Default location varies by the operating system.

Here's an example of how to start the `mw-agent` with input flags:

```bash
mw-agent start \
  --api-key=YOUR_API_KEY \
  --target=https://app.middleware.io \
  --enable-synthetic-monitoring \
  --config-check-interval=60s \
  --docker-endpoint=unix:///var/run/docker.sock \
  --host-tags=tag1=value1,tag2=value2 \
  --logfile=/path/to/logfile.log \
  --logfile-size=10
```

## Environment Variables

In addition to input flags, you can also configure the `mw-agent` using environment variables. These variables correspond to the input flags mentioned above:

- `MW_API_KEY`: Middleware API key for your account.
- `MW_TARGET`: Middleware target for your account.
- `MW_ENABLE_SYNTHETIC_MONITORING`: Enable synthetic monitoring (set to any non-empty value to enable).
- `MW_CONFIG_CHECK_INTERVAL`: Duration string to periodically check for configuration updates. Setting to `0` disables this feature.
- `MW_DOCKER_ENDPOINT`: Set the endpoint for the Docker socket if different from the default.
- `MW_HOST_TAGS`: Tags for this host.
- `MW_LOGFILE`: Log file to store Middleware agent logs.
- `MW_LOGFILE_SIZE`: Log file size to store Middleware agent logs (in MB).
- `MW_CONFIG_FILE`: Location of the configuration file for this agent.

To start the `mw-agent` using environment variables mentioned above, you can use the following command. Replace `YOUR_API_KEY` and other values with your actual configuration:

```bash
export MW_API_KEY=YOUR_API_KEY
export MW_TARGET=YOUR_MIDDLEWARE_TARGET
export MW_ENABLE_SYNTHETIC_MONITORING=true
export MW_CONFIG_CHECK_INTERVAL=60s
export MW_HOST_TAGS=tag1=value1,tag2=value2
export MW_LOGFILE=/path/to/logfile.log
export MW_LOGFILE_SIZE=10

mw-agent start
```

## Configuration file

```yaml
api-key: YOUR_API_KEY
target: https://app.middleware.io
enable-synthetic-monitoring: true
config-check-interval: 60s
docker-endpoint: unix:///var/run/docker.sock
host-tags: tag1=value1,tag2=value2
logfile: /path/to/logfile.log
logfile-size: 10
```

You can save this configuration to a YAML file, such as `mw-agent-config.yaml`. Then, you can specify the configuration file using the `--configuration-file` flag when starting the `mw-agent`:

```bash
mw-agent start --configuration-file=mw-agent-config.yaml
```

This allows you to keep your configuration in a separate file for easier management and reuse.

