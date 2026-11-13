# Adding OBI (eBPF network metrics) to mw-agent

mw-agent uses OBI (OpenTelemetry eBPF Instrumentation) as a collector receiver
named `obi`. OBI comes from our fork,
`middleware-labs/opentelemetry-ebpf-instrumentation`, which is added as a git
submodule. Grafana Beyla and Red Hat's OpenShift collector use the same setup.

## Why a submodule and not a normal Go dependency

OBI's eBPF programs are compiled into generated files (`*_bpfel.go` and
`*_bpfel.o`). These files are in `.gitignore`, so a plain
`go get` or a pseudo-version has no eBPF code, and the build fails:

```
pkg/internal/statsolly/ebpf/stats_tracer.go:151:13: undefined: StatsObjects
```

So we add the OBI source as a submodule, generate the eBPF files there, and
point `go.mod` at it with `replace`.

## Requirements

- mw-agent on collector `v0.163.x` (component/receiver `v1.69.0`). This is the
  same version OBI uses.
- No CGO. OBI builds with `CGO_ENABLED=0`.
- Linux `amd64` and `arm64` only. On Windows and macOS the receiver compiles
  but returns "unsupported platform", so we register it only in Linux-only
  files.

## Steps

### 1. Add the submodule pinned to a commit

```sh
git submodule add https://github.com/middleware-labs/opentelemetry-ebpf-instrumentation.git .obi-src
git -C .obi-src checkout 33120d9d6120
git add .gitmodules .obi-src
```

The submodule records this exact commit. No tag is needed.

To move to a newer OBI commit (full steps in [obi-update.md](obi-update.md)):

```sh
git -C .obi-src fetch
git -C .obi-src checkout <new-commit>
git add .obi-src
```

### 2. Point `go.mod` at the submodule

```
require go.opentelemetry.io/obi/collector v1.0.0

replace go.opentelemetry.io/obi => ./.obi-src

replace go.opentelemetry.io/obi/collector => ./.obi-src/collector
```

The receiver (`collector/`) is its own Go module in OBI, with its own `go.mod`,
so it needs its own `replace`. `go.opentelemetry.io/obi` stays in `go.mod` as
an indirect requirement of it. `v1.0.0` is only a placeholder. Because of the `replace`, Go builds whatever
commit `.obi-src` is on. Then run `go mod tidy`.

### 3. Register the receiver

The receiver is registered only by the agents that use it: the host agent
and the kube agent. Each has a Linux-only file:

```go
// cmd/host-agent/obi_linux.go (same in cmd/kube-agent/obi_linux.go)
package main

import (
	"github.com/middleware-labs/mw-agent/pkg/agent"
	obireceiver "go.opentelemetry.io/obi/collector"
)

func init() {
	agent.ExtraReceivers = append(agent.ExtraReceivers, obireceiver.NewFactory())
}
```

`pkg/agent` declares `ExtraReceivers` in `definitions.go` and adds it to the
receiver list in `hostagent_linux.go` and `kubeagent.go`. `pkg/agent` itself
does not import OBI.

Why: opsai and synthetics also import `pkg/agent`. If OBI were registered in
`pkg/agent`, they would import it too. Then their Linux builds would need the
generated eBPF files, and each binary would carry about 2.8 MB of unused OBI
code.

| Binary | Imports OBI on Linux |
|---|---|
| host-agent, kube-agent | yes |
| opsai, synthetics, kube-config-updater | no |
| any agent on Windows or macOS | no |

The receiver only runs when the backend config includes an `obi` receiver.
That is controlled by `agent-features.cloud-network-monitoring`
(`MW_AGENT_FEATURES_CLOUD_NETWORK_MONITORING`, default `false`). When it is
on, the agent sends `enable_cloud_network_monitoring=true` to the backend,
the same way as `enable_datadog_receiver`.

### 4. Generate the eBPF files before building (Makefile)

```make
OBI_SRC=.obi-src
generate-obi:
	test -f ${OBI_SRC}/go.mod || git submodule update --init ${OBI_SRC}
	$(MAKE) -C ${OBI_SRC} docker-generate
```

These targets depend on `generate-obi`: `build-linux`, `build-linux-amd64`,
`build-linux-arm64`, `build-kube`, `test` and `test_coverage`. They build or
test the host agent or the kube agent on Linux. `build-opsai`,
`build-synthetics`, `build-kube-config-updater` and the Windows and macOS
builds don't need it.

`git submodule update` runs only when `.obi-src` is not checked out, so it
never resets a commit you checked out there yourself.

`docker-generate` uses OBI's generator image, so clang and llvm are not
needed on the build machine. Only Docker is needed. You only have to run it
again after `.obi-src` moves to a new commit.

### 5. Generate the eBPF files in Docker builds

Add a generate stage before the Go build in `Dockerfiles/DockerfileLinux` and
`Dockerfiles/DockerfileKube.base`:

```dockerfile
FROM --platform=$BUILDPLATFORM ghcr.io/open-telemetry/obi-generator:0.2.17 AS bpf-generator
WORKDIR /src/.obi-src
COPY .obi-src .
RUN /generate.sh

FROM golang:1.26.0 AS build
# ... existing steps ...
COPY . .
COPY --from=bpf-generator /src/.obi-src ./.obi-src
RUN CGO_ENABLED=0 go build ...
```

- `/generate.sh` is the generator image's own script. It sets the clang,
  llvm and bpf2go paths, then runs `make generate`.
- One run generates both `amd64` and `arm64`, so the stage runs on the build
  platform.
- The generator image version must match `GEN_IMG` in `.obi-src/Makefile`.
- `DockerfileKube.base` runs `go mod download` before `COPY . .`. Because of
  the `replace`, it also needs OBI's `go.mod` copied in first:

  ```dockerfile
  COPY go.mod go.sum ./
  COPY .obi-src/go.mod .obi-src/go.sum ./.obi-src/
  COPY .obi-src/collector/go.mod .obi-src/collector/go.sum ./.obi-src/collector/
  RUN go mod download
  ```

`DockerfileKube.base` builds three images (`APP_BINARY` = `mw-kube-agent`,
`mw-opsai` or `mw-synthetics`). Only `mw-kube-agent` needs the generated
files, so the stage is chosen by `APP_BINARY`, and Docker skips
`bpf-generator` for opsai and synthetics:

```dockerfile
ARG APP_BINARY=mw-kube-agent
# ... bpf-generator stage as above ...
FROM --platform=$BUILDPLATFORM golang:1.26.0 AS obi-skip
RUN mkdir -p /src/.obi-src

FROM bpf-generator AS obi-mw-kube-agent
FROM obi-skip AS obi-mw-opsai
FROM obi-skip AS obi-mw-synthetics
FROM obi-${APP_BINARY} AS obi
# build stage: COPY --from=obi /src/.obi-src ./.obi-src
```

A new image built from this Dockerfile needs its own `obi-<APP_BINARY>` line.

`DockerfileKubeLocal.base` copies a binary that's already built, so it needs
no change.

### 6. Check out submodules in CI

Every workflow that builds the agent needs the submodule checked out:

```yaml
- uses: actions/checkout@v4
  with:
    submodules: true
```

Most workflows already use `submodules: 'recursive'`. `code-analysis.yaml`
did not, so it was added there.

The deb and rpm workflows run `go build` directly on the runner, so they run
the generator first:

```yaml
- name: Generate OBI eBPF files
  run: make -C agent generate-obi
```

`code-analysis.yaml` runs `make test_coverage`, which already depends on
`generate-obi`.

## Example receiver config

```yaml
receivers:
  obi:
    # same options as OBI's config file
```

See `.obi-src/examples/otel-collector/config.yaml` for a full example.

## Tested so far

Tested in a copy of mw-agent at collector v0.163, with OBI from a local path:

- `go mod tidy` works with no version conflicts. About 75 existing deps get
  minor or patch bumps. Bigger ones: `otel/sdk/log` v0.21 → v1.47 and
  `contrib/otelconf` v0.25 → v0.27.
- `CGO_ENABLED=0` builds pass for `linux/amd64` and `linux/arm64`.
- Binary size (stripped, `-s -w`):

  | Arch | Without OBI | With OBI | Added |
  |---|---|---|---|
  | amd64 | 253.7 MB | 300.8 MB | +47.1 MB (+19%) |
  | arm64 | 234.4 MB | 279.1 MB | +44.8 MB (+19%) |
  | amd64 gzipped | 65.4 MB | 79.4 MB | +14.0 MB |

Not tested yet:

- The Docker generate stage (step 5).
- Running the agent with the `obi` receiver enabled.
- `make test` with the bumped dependencies.
- Windows and macOS builds.
