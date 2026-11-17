# Updating the OBI commit

mw-agent uses OBI from the `.obi-src` git submodule. The submodule points to
one exact commit of `middleware-labs/opentelemetry-ebpf-instrumentation`.
Pushing to the OBI repo does not change what mw-agent builds. To use new OBI
code, you have to move the pointer to the new commit and commit that in mw-agent.

See [obi-integration.md](obi-integration.md) for how the submodule is set up.

## 1. Push the OBI change

In your OBI repo:

```sh
cd ~/opentelemetry-ebpf-instrumentation
git add <files>
git commit -m "<message>"
git push origin network-performance-monitoring
git rev-parse HEAD        # copy this commit hash
```

mw-agent can only use commits that are pushed to GitHub.

## 2. Check out the new commit in `.obi-src`

In mw-agent:

```sh
cd ~/mw-agent
git -C .obi-src fetch origin
git -C .obi-src checkout <new-commit-hash>
```

To use the latest commit of the branch instead of a specific hash:

```sh
git -C .obi-src checkout origin/network-performance-monitoring
```

Check that it moved:

```sh
git -C .obi-src log --oneline -1
git status                 # shows: modified: .obi-src (new commits)
```

## 3. Regenerate the eBPF files

Go does not regenerate the eBPF files on its own, so after a checkout they
may still be the old ones. Regenerate:

```sh
make generate-obi
```

This only rebuilds the eBPF programs whose sources changed. The generated
files are ignored by OBI's `.gitignore`, so they never show up in
`git status`.

## 4. Update Go modules

```sh
go mod tidy
```

This matters when the new OBI commit changed its `go.mod`. Check:

- **Collector version.** mw-agent and OBI must use the same collector
  release (now `v0.163.x`, component `v1.69.0`). If OBI moved to a newer
  collector, mw-agent has to be upgraded to that version too. Otherwise the
  build fails.
- **Other bumped dependencies.** `git diff go.mod` shows what changed.
- **New OBI modules.** If `go mod tidy` says a package is in no module (OBI
  split it into its own `go.mod`, as it did for `collector/`), add a
  `require` and a `replace` pointing into `.obi-src`, and copy that `go.mod`
  in `Dockerfiles/DockerfileKube.base` before `go mod download`.

## 5. Check the generator image version

The Dockerfiles pin the image that generates the eBPF files. It must match
OBI's Makefile:

```sh
grep '^GEN_IMG' .obi-src/Makefile
grep obi-generator Dockerfiles/DockerfileLinux Dockerfiles/DockerfileKube.base
```

If they differ, update the `FROM ghcr.io/open-telemetry/obi-generator:<version>`
line in both Dockerfiles.

## 6. Build and test

```sh
make build-linux-amd64
make build-linux-arm64
make test
```

## 7. Commit the new pointer in mw-agent

```sh
git add .obi-src go.mod go.sum
git commit -m "Update OBI to <short-hash>"
git push
```

`git add .obi-src` saves the new commit hash (the submodule pointer), not
OBI's files. To check what is recorded:

```sh
git ls-files -s .obi-src
# 160000 <commit-hash> 0	.obi-src
```

## Other people's checkouts

After pulling an mw-agent commit that moved the pointer, run:

```sh
git pull
git submodule update --init
make generate-obi
```

`git pull` does not move `.obi-src` on its own. `git submodule update`
checks out the commit that mw-agent records.

## Going back to an older OBI commit

Same steps, with the older hash:

```sh
git -C .obi-src checkout <old-commit-hash>
make generate-obi
go mod tidy
git add .obi-src go.mod go.sum
git commit -m "Revert OBI to <short-hash>"
```

## Rules

- Don't edit code inside `.obi-src`. Make changes in the OBI repo, push, then
  move the pointer. A commit made inside `.obi-src` that is never pushed
  breaks the build for everyone else.
- Only point to commits that exist on GitHub.
