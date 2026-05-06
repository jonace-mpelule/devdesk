APP := devdesk
MODULE := $(shell go list -m)
CMD := ./cmd/devdesk
DIST := dist
BUMP ?= patch
VERSION ?= $(shell ./scripts/next-version.sh $(BUMP))
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GOCACHE ?= $(CURDIR)/.gocache
GOMODCACHE ?= $(CURDIR)/.gomodcache
LDFLAGS := -s -w -X '$(MODULE)/internal/devdesk.Version=$(VERSION)' -X '$(MODULE)/internal/devdesk.Commit=$(COMMIT)' -X '$(MODULE)/internal/devdesk.Date=$(DATE)'

.PHONY: run test clean version build build-mac build-mac-arm64 build-mac-amd64 package-mac checksum tag release install-local

run:
	go run $(CMD)

test:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go test ./...

clean:
	rm -rf $(DIST)

version:
	@echo $(VERSION)

build:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST)/$(APP) $(CMD)

build-mac: build-mac-arm64 build-mac-amd64

build-mac-arm64:
	mkdir -p $(DIST)/$(APP)_$(VERSION)_darwin_arm64
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST)/$(APP)_$(VERSION)_darwin_arm64/$(APP) $(CMD)

build-mac-amd64:
	mkdir -p $(DIST)/$(APP)_$(VERSION)_darwin_amd64
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST)/$(APP)_$(VERSION)_darwin_amd64/$(APP) $(CMD)

package-mac: clean build-mac
	cd $(DIST) && tar -czf $(APP)_$(VERSION)_darwin_arm64.tar.gz $(APP)_$(VERSION)_darwin_arm64
	cd $(DIST) && tar -czf $(APP)_$(VERSION)_darwin_amd64.tar.gz $(APP)_$(VERSION)_darwin_amd64
	$(MAKE) checksum

checksum:
	cd $(DIST) && shasum -a 256 *.tar.gz > checksums.txt

tag: test
	./scripts/tag-version.sh $(VERSION)

release: package-mac tag
	gh release create $(VERSION) $(DIST)/*.tar.gz $(DIST)/checksums.txt --title "$(VERSION)" --notes "DevDesk $(VERSION)"

install-local: build
	mkdir -p $(HOME)/.local/bin
	cp $(DIST)/$(APP) $(HOME)/.local/bin/$(APP)
