WAILS := $(shell go env GOPATH)/bin/wails
GOBIN := $(shell go env GOPATH)/bin

.PHONY: dev build build-app build-cli install test clean

dev:
	$(WAILS) dev

build: build-app build-cli

build-app:
	$(WAILS) build

build-cli:
	mkdir -p build/bin
	go build -o build/bin/kf ./cmd/kf

install: build
	rm -rf /Applications/Kingfisher.app
	cp -R build/bin/Kingfisher.app /Applications/Kingfisher.app
	mkdir -p $(GOBIN)
	cp build/bin/kf $(GOBIN)/kf
	@echo "Installed /Applications/Kingfisher.app and $(GOBIN)/kf"

test:
	go test -v ./...
	cd frontend && npm run check && npm run build

clean:
	rm -rf build/bin frontend/dist
