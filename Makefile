build-windows:
	GOOS=windows CGO_ENABLED=0 go build -o build/mw-windows-agent.exe cmd/host-agent/main.go
build-linux:
	GOOS=linux CGO_ENABLED=0 go build -o build/mw-host-agent cmd/host-agent/main.go

build-kube:
	GOOS=linux CGO_ENABLED=0 go build -o build/mw-host-agent cmd/kube-agent/main.go
build: build-linux build-windows build-kube

package-windows: build-windows
	makensis scripts/package-windows/package-windows.nsi 

clean:
	go clean
	rm mw-host-agent
	rm mw-windows-agent

test:
	go test ./...

test_coverage:
	go test ./... -coverprofile=coverage.out

vet:
	go vet

lint:
	golangci-lint run --enable-all
