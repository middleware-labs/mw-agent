module github.com/middleware-labs/mw-agent

go 1.26.0

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter => github.com/middleware-labs/opentelemetry-collector-contrib/internal/filter v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl => github.com/middleware-labs/opentelemetry-collector-contrib/pkg/ottl v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor => github.com/middleware-labs/opentelemetry-collector-contrib/processor/transformprocessor v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker => github.com/middleware-labs/opentelemetry-collector-contrib/internal/docker v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/mwmongodbreceiver v0.0.0-20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/mwpostgresqlreceiver v0.0.0-20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/sqlquery => github.com/middleware-labs/opentelemetry-collector-contrib/internal/sqlquery v0.0.0-20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/apachereceiver v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/mwmysqlreceiver v0.0.0-20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/rabbitmqreceiver v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/sqlserverreceiver v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/nginxreceiver v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor => github.com/middleware-labs/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders => github.com/middleware-labs/opentelemetry-collector-contrib/internal/metadataproviders v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zookeeperreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/zookeeperreceiver v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor => github.com/middleware-labs/opentelemetry-collector-contrib/processor/filterprocessor v0.91.1-0.20261112055836-5a01843f508c

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/saphanareceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/saphanareceiver v0.91.1-0.20261112055836-5a01843f508c

replace go.opentelemetry.io/collector => go.opentelemetry.io/collector v0.163.0

require (
	github.com/prometheus/common v0.72.1
	github.com/stretchr/testify v1.12.1
	github.com/urfave/cli/v2 v2.25.7
	go.opentelemetry.io/collector/pdata v1.69.0 // indirect
	go.uber.org/zap v1.28.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.37.1
	k8s.io/apimachinery v0.37.1
	k8s.io/client-go v0.37.1
)

require (
	github.com/grafana/pyroscope-go v1.1.1
	github.com/kardianos/service v1.2.2
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatorateprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor v1.2.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/logdedupprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/redactionprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/saphanareceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zookeeperreceiver v0.163.0
	go.opentelemetry.io/collector/component v1.69.0
	go.opentelemetry.io/collector/confmap v1.69.0
	go.opentelemetry.io/collector/exporter v1.69.0
	go.opentelemetry.io/collector/exporter/debugexporter v0.163.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.163.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.163.0
	go.opentelemetry.io/collector/extension v1.69.0
	go.opentelemetry.io/collector/processor v1.69.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.163.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.163.0
	go.opentelemetry.io/collector/receiver v1.69.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.163.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/middleware-labs/mw-injector v1.2.3
	github.com/middleware-labs/synthetics-agent v1.0.65
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampextension v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/datadogreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlqueryreceiver v0.163.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcpcheckreceiver v0.163.0
	go.opentelemetry.io/collector/component/componentstatus v0.163.0
	go.opentelemetry.io/collector/confmap/provider/envprovider v1.69.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.69.0
	go.opentelemetry.io/collector/confmap/provider/yamlprovider v1.69.0
	go.opentelemetry.io/collector/connector v0.163.0
	go.opentelemetry.io/collector/consumer v1.69.0
	go.opentelemetry.io/collector/exporter/nopexporter v0.163.0
	go.opentelemetry.io/collector/featuregate v1.69.0
	go.opentelemetry.io/collector/otelcol v0.163.0
	go.opentelemetry.io/collector/receiver/nopreceiver v0.163.0
	go.opentelemetry.io/collector/service v0.163.0
	go.uber.org/zap/exp v0.3.0
)

require (
	github.com/AlessandroPomponio/go-gibberish v0.0.0-20191004143433-a2d4156f0396 // indirect
	github.com/aws/aws-sdk-go-v2/service/route53 v1.70.4 // indirect
	github.com/caarlos0/env/v11 v11.4.1 // indirect
	github.com/cilium/ebpf v0.22.0 // indirect
	github.com/containers/common v0.64.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.5 // indirect
	github.com/grafana/go-offsets-tracker v0.1.7 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/microsoft/go-winmd v0.0.0-20260929074142-364b6859d0c4 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/ohler55/ojg v1.28.7 // indirect
	github.com/oschwald/maxminddb-golang/v2 v2.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/tidwall/jsonc v0.3.3 // indirect
	github.com/vektah/gqlparser/v2 v2.5.62 // indirect
	github.com/vishvananda/netlink v1.3.1 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/xwb1989/sqlparser v0.0.0-20180606152119-120387863bf2 // indirect
	github.com/yl2chen/cidranger v1.0.2 // indirect
	go.opentelemetry.io/contrib/detectors/aws/ec2/v2 v2.6.0 // indirect
	go.opentelemetry.io/ebpf-profiler v0.0.202640 // indirect
	go.opentelemetry.io/proto/otlp/processcontext/v1development v0.4.0 // indirect
	golang.org/x/arch v0.31.0 // indirect
)

require (
	cloud.google.com/go/auth v0.23.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute v1.72.0 // indirect
	cloud.google.com/go/compute/metadata v0.10.0 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.23.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.14.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v8 v8.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v8 v8.0.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.2.0 // indirect
	github.com/Azure/go-ntlmssp v0.1.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.8.0 // indirect
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/ClickHouse/ch-go v0.73.0 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.47.0 // indirect
	github.com/Code-Hex/go-generics-cache v1.5.1 // indirect
	github.com/DataDog/agent-payload/v5 v5.0.211 // indirect
	github.com/DataDog/datadog-agent/comp/core/tagger/origindetection v0.83.3 // indirect
	github.com/DataDog/datadog-agent/pkg/obfuscate v0.83.3 // indirect
	github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/otlp/attributes v0.84.0-devel.0.20260930220258-a8c2777f7050 // indirect
	github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/otlp/metrics v0.83.0 // indirect
	github.com/DataDog/datadog-agent/pkg/proto v0.83.3 // indirect
	github.com/DataDog/datadog-agent/pkg/remoteconfig/state v0.83.3 // indirect
	github.com/DataDog/datadog-agent/pkg/template v0.83.3 // indirect
	github.com/DataDog/datadog-agent/pkg/trace v0.83.3 // indirect
	github.com/DataDog/datadog-agent/pkg/trace/log v0.83.3 // indirect
	github.com/DataDog/datadog-agent/pkg/trace/stats v0.83.3 // indirect
	github.com/DataDog/datadog-agent/pkg/trace/traceutil v0.83.3 // indirect
	github.com/DataDog/datadog-agent/pkg/util/hostname/validate v0.83.0 // indirect
	github.com/DataDog/datadog-agent/pkg/util/log v0.83.3 // indirect
	github.com/DataDog/datadog-agent/pkg/util/quantile v0.83.0 // indirect
	github.com/DataDog/datadog-agent/pkg/util/scrubber v0.83.3 // indirect
	github.com/DataDog/datadog-agent/pkg/version v0.83.3 // indirect
	github.com/DataDog/datadog-api-client-go/v2 v2.66.0 // indirect
	github.com/DataDog/datadog-go/v5 v5.9.1 // indirect
	github.com/DataDog/go-sqllexer v0.2.4 // indirect
	github.com/DataDog/go-tuf v1.1.1-0.5.2 // indirect
	github.com/DataDog/sketches-go v1.4.8 // indirect
	github.com/DataDog/zstd v1.5.8-0.20260421145859-31a7e515a571 // indirect
	github.com/DeRuina/timberjack v1.4.8 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.38.0 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/Microsoft/go-winio v0.6.3 // indirect
	github.com/SAP/go-hdb v1.19.2 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/adakailabs/go-traceroute v0.0.0-20210727014431-97524352ab91 // indirect
	github.com/alecthomas/participle/v2 v2.1.4 // indirect
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/alexbrainman/sspi v0.0.0-20250919150558-7d374ff0d59e // indirect
	github.com/andybalholm/brotli v1.2.6 // indirect
	github.com/antchfx/xmlquery v1.5.1 // indirect
	github.com/antchfx/xpath v1.3.8 // indirect
	github.com/apache/arrow-go/v18 v18.4.0 // indirect
	github.com/apache/thrift v0.25.0 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/aws/aws-msk-iam-sasl-signer-go v1.0.4 // indirect
	github.com/aws/aws-sdk-go v1.55.8 // indirect
	github.com/aws/aws-sdk-go-v2 v1.47.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.20 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.33.9 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.20.9 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.20.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager v0.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.6 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.6 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.336.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ecs v1.100.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.60.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.11.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.20.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/kafka v1.63.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.64.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/rds v1.128.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.113.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.10.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.38.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.43.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.51.4 // indirect
	github.com/aws/smithy-go v1.28.5 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/bboreham/go-loser v0.0.0-20230920113527-fcc2c21820a3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v4 v4.10.2 // indirect
	github.com/brunoscheufler/aws-ecs-metadata-go v0.0.0-20221221133751-67e37ae746cd // indirect
	github.com/bufbuild/protocompile v0.14.1 // indirect
	github.com/buger/jsonparser v1.1.2 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cenkalti/backoff/v7 v7.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20260202195803-dba9d589def2 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/digitalocean/go-metadata v0.0.0-20250129100319-e3650a3df44b // indirect
	github.com/digitalocean/godo v1.206.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.8.2 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dvsekhvalnov/jose2go v1.7.0 // indirect
	github.com/ebitengine/purego v0.11.1 // indirect
	github.com/edsrzf/mmap-go v1.2.1-0.20241212181136-fad1cd13edbd // indirect
	github.com/elastic/go-grok v0.3.1 // indirect
	github.com/elastic/lunes v0.2.2 // indirect
	github.com/elastic/proxy-connect-dialer-go v0.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.39.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.3 // indirect
	github.com/expr-lang/expr v1.17.8 // indirect
	github.com/facette/natsort v0.0.0-20181210072756-2cd4dd1e2dcb // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/felixge/fgprof v0.9.5 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20251226215517-609e4778396f // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.4 // indirect
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.8 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-ldap/ldap/v3 v3.4.15 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/analysis v1.0.0 // indirect
	github.com/go-openapi/errors v0.22.9 // indirect
	github.com/go-openapi/jsonpointer v1.0.1 // indirect
	github.com/go-openapi/jsonreference v1.0.1 // indirect
	github.com/go-openapi/loads v0.25.2 // indirect
	github.com/go-openapi/spec v1.0.0 // indirect
	github.com/go-openapi/strfmt v0.27.2 // indirect
	github.com/go-openapi/swag v0.29.2 // indirect
	github.com/go-openapi/swag/cmdutils v0.29.2 // indirect
	github.com/go-openapi/swag/conv v0.29.2 // indirect
	github.com/go-openapi/swag/fileutils v0.29.2 // indirect
	github.com/go-openapi/swag/jsonutils v0.29.2 // indirect
	github.com/go-openapi/swag/loading v0.29.2 // indirect
	github.com/go-openapi/swag/mangling v0.29.2 // indirect
	github.com/go-openapi/swag/netutils v0.29.2 // indirect
	github.com/go-openapi/swag/pools v0.29.2 // indirect
	github.com/go-openapi/swag/stringutils v0.29.2 // indirect
	github.com/go-openapi/swag/typeutils v0.29.2 // indirect
	github.com/go-openapi/swag/yamlutils v0.29.2 // indirect
	github.com/go-openapi/validate v1.0.0 // indirect
	github.com/go-resty/resty/v2 v2.17.2 // indirect
	github.com/go-sql-driver/mysql v1.10.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/go-zookeeper/zk v1.0.4 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/google/go-tpm v0.9.8 // indirect
	github.com/google/pprof v0.0.0-20260906184651-6331bc6350fe // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.20 // indirect
	github.com/googleapis/gax-go/v2 v2.26.2 // indirect
	github.com/gophercloud/gophercloud/v2 v2.14.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/grafana/clusterurl v0.3.0 // indirect
	github.com/grafana/pyroscope-go/godeltaprof v0.1.8 // indirect
	github.com/grafana/regexp v0.0.0-20250905093917-f7b3be9d1853 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.31.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hashicorp/consul/api v1.34.4 // indirect
	github.com/hashicorp/cronexpr v1.1.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.6.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/nomad/api v0.0.0-20260907080526-08ef8f3d26da // indirect
	github.com/hashicorp/serf v0.10.4 // indirect
	github.com/hetznercloud/hcloud-go/v2 v2.49.0 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/ionos-cloud/sdk-go/v6 v6.3.11 // indirect
	github.com/jaegertracing/jaeger-idl v0.13.1 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jhump/protoreflect v1.17.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/k0kubun/pp v3.0.1+incompatible // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/knadh/koanf/v2 v2.3.8 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-syslog/v4 v4.6.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20190525184631-5f46317e436b // indirect
	github.com/lib/pq v1.12.3 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/likexian/gokit v0.25.15 // indirect
	github.com/likexian/whois v1.15.5 // indirect
	github.com/likexian/whois-parser v1.24.20 // indirect
	github.com/linode/go-metadata v0.3.1 // indirect
	github.com/linode/linodego v1.69.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20260330125221-c963978e514e // indirect
	github.com/magefile/mage v1.15.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.23 // indirect
	github.com/mdlayher/socket v0.6.0 // indirect
	github.com/mdlayher/vsock v1.3.0 // indirect
	github.com/microsoft/go-mssqldb v1.11.2 // indirect
	github.com/middleware-labs/innoParser v0.0.0-20240729092319-ddbdd8e42266 // indirect
	github.com/miekg/dns v1.1.73 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/moby/api v1.56.1 // indirect
	github.com/moby/moby/client v0.6.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mongodb-forks/digest v1.1.0 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/oklog/ulid/v2 v2.1.2 // indirect
	github.com/open-telemetry/opamp-go v0.25.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/k8sleaderelector v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampcustommessages v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/datadog v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/exp/metrics v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/gopsutilenv v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/healthcheck v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sinventory v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kafka v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/pdatautil v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sqlquery v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/core/xidutils v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/datadog v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/expohisto v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/kafka/configkafka v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/kafka/topic v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/xprofile v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/xottl v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sampling v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/status v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/zipkin v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatocumulativeprocessor v0.163.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/scraper/zookeeperscraper v0.163.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20251015124057-db0dee36e235 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/oracle/oci-go-sdk/v65 v65.124.1 // indirect
	github.com/outcaste-io/ristretto v0.2.3 // indirect
	github.com/outscale/osc-sdk-go/v2 v2.35.0 // indirect
	github.com/ovh/go-ovh v1.9.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/paulmach/orb v0.13.0 // indirect
	github.com/pb33f/jsonpath v0.8.2 // indirect
	github.com/pb33f/libopenapi v0.38.7 // indirect
	github.com/pb33f/ordered-map/v2 v2.3.1 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.33 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20260805114148-88456608a4f6 // indirect
	github.com/prometheus-community/pro-bing v0.1.0 // indirect
	github.com/prometheus/alertmanager v0.34.0 // indirect
	github.com/prometheus/client_golang v1.25.0 // indirect
	github.com/prometheus/client_golang/exp v0.0.0-20260907100614-57bb367da472 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common/assets v0.2.0 // indirect
	github.com/prometheus/exporter-toolkit v0.20.0 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/prometheus/prometheus v0.315.0 // indirect
	github.com/prometheus/sigv4 v0.5.0 // indirect
	github.com/puzpuzpuz/xsync/v4 v4.5.0 // indirect
	github.com/redis/go-redis/v9 v9.23.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/scaleway/scaleway-sdk-go v1.38.0 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.11.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shirou/gopsutil/v4 v4.26.9 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/shurcooL/httpfs v0.0.0-20230704072500-f1e31cf0ba5c // indirect
	github.com/sijms/go-ora/v2 v2.9.0 // indirect
	github.com/snowflakedb/gosnowflake/v2 v2.2.0 // indirect
	github.com/sony/gobreaker/v2 v2.4.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stackitcloud/stackit-sdk-go/core v0.27.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/thda/tds v0.1.7 // indirect
	github.com/tilinna/clock v1.1.0 // indirect
	github.com/tinylib/msgp v1.6.4 // indirect
	github.com/tklauser/go-sysconf v0.4.0 // indirect
	github.com/tklauser/numcpus v0.12.0 // indirect
	github.com/twmb/franz-go v1.22.1 // indirect
	github.com/twmb/franz-go/pkg/kadm v1.19.0 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.14.0 // indirect
	github.com/twmb/franz-go/pkg/sasl/kerberos v1.1.0 // indirect
	github.com/twmb/franz-go/plugin/kzap v1.1.2 // indirect
	github.com/twmb/murmur3 v1.2.1 // indirect
	github.com/ua-parser/uap-go v0.0.0-20251207011819-db9adb27a0b8 // indirect
	github.com/valyala/fastjson v1.6.10 // indirect
	github.com/vultr/govultr/v3 v3.33.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	go.mongodb.org/atlas v0.38.0 // indirect
	go.mongodb.org/mongo-driver/v2 v2.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/client v1.69.0 // indirect
	go.opentelemetry.io/collector/component/componenttest v0.163.0 // indirect
	go.opentelemetry.io/collector/config/configauth v1.69.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.69.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v1.69.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.163.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.69.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.69.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.69.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v1.69.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.69.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.163.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.69.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.163.0 // indirect
	go.opentelemetry.io/collector/connector/connectortest v0.163.0 // indirect
	go.opentelemetry.io/collector/connector/xconnector v0.163.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.163.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.163.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.163.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.163.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper v0.163.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.163.0 // indirect
	go.opentelemetry.io/collector/exporter/exportertest v0.163.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.163.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.69.0 // indirect
	go.opentelemetry.io/collector/extension/extensioncapabilities v0.163.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.163.0 // indirect
	go.opentelemetry.io/collector/extension/extensiontest v0.163.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.163.0 // indirect
	go.opentelemetry.io/collector/filter v0.163.0 // indirect
	go.opentelemetry.io/collector/internal/componentalias v0.163.0 // indirect
	go.opentelemetry.io/collector/internal/fanoutconsumer v0.163.0 // indirect
	go.opentelemetry.io/collector/internal/memorylimiter v0.163.0 // indirect
	go.opentelemetry.io/collector/internal/sharedcomponent v0.163.0 // indirect
	go.opentelemetry.io/collector/internal/statusutil v0.163.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.163.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.163.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.163.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.163.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.69.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.163.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper v0.163.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper/xprocessorhelper v0.163.0 // indirect
	go.opentelemetry.io/collector/processor/processortest v0.163.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.163.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.163.0 // indirect
	go.opentelemetry.io/collector/receiver/receivertest v0.163.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.163.0 // indirect
	go.opentelemetry.io/collector/scraper v0.163.0 // indirect
	go.opentelemetry.io/collector/scraper/scraperhelper v0.163.0 // indirect
	go.opentelemetry.io/collector/semconv v0.128.1-0.20250610090210-188191247685 // indirect
	go.opentelemetry.io/collector/service/hostcapabilities v0.163.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.20.0 // indirect
	go.opentelemetry.io/contrib/detectors/aws/ecs v1.47.0 // indirect
	go.opentelemetry.io/contrib/detectors/aws/eks v1.47.0 // indirect
	go.opentelemetry.io/contrib/detectors/aws/elasticbeanstalk v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/azure/azureappservice v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/azure/azurecontainerapps v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/azure/azurefunctions v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/azure/azurevm v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.47.0 // indirect
	go.opentelemetry.io/contrib/detectors/hetzner v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/ibmcloud/vpc v0.19.0 // indirect
	go.opentelemetry.io/contrib/detectors/vultr v0.19.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.71.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.71.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0 // indirect
	go.opentelemetry.io/contrib/otelconf v0.27.0 // indirect
	go.opentelemetry.io/contrib/propagators/autoprop v0.72.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.47.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.47.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.47.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.47.0 // indirect
	go.opentelemetry.io/obi v1.0.0
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.23.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.23.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.69.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.23.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/sdk v1.47.0 // indirect
	go.opentelemetry.io/otel/sdk/log v1.47.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.1 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go.uber.org/goleak v1.3.0 // indirect
	go.uber.org/mock v0.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.6 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/exp v0.0.0-20260908205506-85c1c2202aba // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/telemetry v0.0.0-20261009151841-97e8ea77aa2a // indirect
	golang.org/x/term v0.47.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gonum.org/v1/gonum v0.17.0 // indirect
	google.golang.org/api v0.297.0 // indirect
	google.golang.org/genproto v0.0.0-20260715232425-e75dac1f907d // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260928230214-8a89bd6388cc // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260928230214-8a89bd6388cc // indirect
	google.golang.org/grpc v1.84.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260821135717-be32def86098 // indirect
	k8s.io/kubelet v0.37.1 // indirect
	k8s.io/utils v0.0.0-20260707023825-cf1189d6abe3 // indirect
	sigs.k8s.io/json v0.0.0-20260909141634-11ed52e25bc5 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

replace go.opentelemetry.io/obi => ./.obi-src
