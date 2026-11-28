WAILS := $(shell go env GOPATH)/bin/wails
GOBIN := $(shell go env GOPATH)/bin

.PHONY: dev build build-app build-cli install install-plugin test clean

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

install-plugin:
	mkdir -p "$(HOME)/.gemini/config/plugins/kingfisher"
	cp -R plugins/antigravity-kingfisher/. "$(HOME)/.gemini/config/plugins/kingfisher/"
	@echo "Installed Antigravity plugin to $(HOME)/.gemini/config/plugins/kingfisher"

test:
	go test -v ./...
	cd frontend && npm run check && npm run build

clean:
	rm -rf build/bin frontend/dist
