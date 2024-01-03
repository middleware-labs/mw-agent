module github.com/middleware-labs/mw-agent

go 1.20

// replace checkagent => github.com/middleware-labs/check-agent v0.0.0-20230411103956-f67eb953ba5e
replace checkagent => ./check-agent

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.84.1-0.20231225132042-f4169e2516ad

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.84.1-0.20231225132042-f4169e2516ad

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.84.1-0.20231225132042-f4169e2516ad

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.84.1-0.20231225132042-f4169e2516ad

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.84.1-0.20231225132042-f4169e2516ad

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/k8seventsreceiver v0.84.1-0.20231225132042-f4169e2516ad

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.84.1-0.20231225132042-f4169e2516ad

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/mongodbreceiver v0.84.1-0.20231225132042-f4169e2516ad

replace github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver => github.com/middleware-labs/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver v0.84.1-0.20231225132042-f4169e2516ad

require (
	// checkagent v0.0.0-00010101000000-000000000000
	github.com/k8sgpt-ai/k8sgpt v0.1.8
	github.com/prometheus/common v0.44.0
	github.com/stretchr/testify v1.8.4
	github.com/urfave/cli/v2 v2.25.7
	go.opentelemetry.io/collector/pdata v1.0.0-rcv0014
	go.uber.org/zap v1.25.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.28.1
	k8s.io/apimachinery v0.28.1
	k8s.io/client-go v0.28.1
)

require (
	github.com/kardianos/service v1.2.2
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jmxreceiver v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver v0.84.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver v0.84.0
	go.opentelemetry.io/collector v0.84.0
	go.opentelemetry.io/collector/component v0.84.0
	go.opentelemetry.io/collector/confmap v0.84.0
	go.opentelemetry.io/collector/exporter v0.84.0
	go.opentelemetry.io/collector/exporter/loggingexporter v0.84.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.84.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.84.0
	go.opentelemetry.io/collector/extension v0.84.0
	go.opentelemetry.io/collector/processor v0.84.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.84.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.84.0
	go.opentelemetry.io/collector/receiver v0.84.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.84.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	cloud.google.com/go/compute v1.23.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.4-0.20230617002413-005d2dfb6b68 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.4.2 // indirect
	github.com/Azure/azure-sdk-for-go v65.0.0+incompatible // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.28 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.23 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.19.1 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/alecthomas/participle/v2 v2.0.0 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/antonmedv/expr v1.14.3 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/aws/aws-sdk-go v1.45.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cncf/xds/go v0.0.0-20230607035331-e9ce68804cb4 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/digitalocean/godo v1.98.0 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v24.0.5+incompatible // indirect
	github.com/docker/go-connections v0.4.1-0.20210727194412-58542c764a11 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.10.2 // indirect
	github.com/envoyproxy/go-control-plane v0.11.1-0.20230524094728-9239064ad72f // indirect
	github.com/envoyproxy/protoc-gen-validate v0.10.1 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/go-redis/redis/v7 v7.4.1 // indirect
	github.com/go-resty/resty/v2 v2.7.0 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/go-zookeeper/zk v1.0.3 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.5 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.5 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/gophercloud/gophercloud v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grafana/regexp v0.0.0-20221122212121-6b5c0a4cb7fd // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.17.1 // indirect
	github.com/hashicorp/consul/api v1.24.0 // indirect
	github.com/hashicorp/cronexpr v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.2 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/nomad/api v0.0.0-20230418003350-3067191c5197 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/hetznercloud/hcloud-go v1.42.0 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/ionos-cloud/sdk-go/v6 v6.1.6 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/knadh/koanf/v2 v2.0.1 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/leoluk/perflib_exporter v0.2.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/linode/linodego v1.16.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/miekg/dns v1.1.53 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20220423185008-bf980b35cac4 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/mostynb/go-grpc-compression v1.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.84.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters v0.84.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc4 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20210521082421-73d9475a9142 // indirect
	github.com/ovh/go-ovh v1.4.1 // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/prometheus/prometheus v0.44.0 // indirect
	github.com/prometheus/statsd_exporter v0.22.7 // indirect
	github.com/rs/cors v1.9.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sashabaranov/go-openai v1.14.0 // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.15 // indirect
	github.com/shirou/gopsutil/v3 v3.23.7 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.16.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/vultr/govultr/v2 v2.17.2 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.mongodb.org/mongo-driver v1.12.1 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.84.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v0.84.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.84.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.84.0 // indirect
	go.opentelemetry.io/collector/config/confignet v0.84.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v0.84.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.84.0 // indirect
	go.opentelemetry.io/collector/config/configtls v0.84.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.84.0 // indirect
	go.opentelemetry.io/collector/connector v0.84.0 // indirect
	go.opentelemetry.io/collector/consumer v0.84.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.84.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.0.0-rcv0014 // indirect
	go.opentelemetry.io/collector/semconv v0.84.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.42.1-0.20230612162650-64be7e574a17 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.42.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.17.0 // indirect
	go.opentelemetry.io/otel v1.16.0 // indirect
	go.opentelemetry.io/otel/bridge/opencensus v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.16.0 // indirect
	go.opentelemetry.io/otel/metric v1.16.0 // indirect
	go.opentelemetry.io/otel/sdk v1.16.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.16.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/goleak v1.2.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/exp v0.0.0-20230711023510-fffb14384f22 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/oauth2 v0.11.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/term v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.12.0 // indirect
	gonum.org/v1/gonum v0.14.0 // indirect
	google.golang.org/api v0.138.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230803162519-f966b187b2e5 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/grpc v1.57.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230717233707-2695361300d9 // indirect
	k8s.io/kubelet v0.27.4 // indirect
	k8s.io/utils v0.0.0-20230711102312-30195339c3c7 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.3.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
