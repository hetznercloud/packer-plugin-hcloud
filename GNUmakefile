NAME=hcloud
BINARY=packer-plugin-$(NAME)
ifeq ($(OS), Windows_NT)
# Prevent "Ignoring plugin match packer-plugin-hcloud, no exe extension"
BINARY=packer-plugin-$(NAME).exe
endif

COUNT?=1
TEST?=./...
HASHICORP_PACKER_PLUGIN_SDK_VERSION?=$(shell go list -m github.com/hashicorp/packer-plugin-sdk | cut -d " " -f2)

.PHONY: dev

build:
	go build -o $(BINARY)

dev: build
	mkdir -p ~/.config/packer/plugins
	mv $(BINARY) ~/.config/packer/plugins

test:
	go test -race -count $(COUNT) -v $(TEST) -timeout=3m -coverprofile=coverage.txt

install-packer-sdc: ## Install packer software development command
	go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@$(HASHICORP_PACKER_PLUGIN_SDK_VERSION)

plugin-check: install-packer-sdc build
	packer-sdc plugin-check $(BINARY)

testacc: build
	PACKER_ACC=1 PACKER_PLUGIN_PATH=$(PWD) go test -count $(COUNT) -v $(TEST) -timeout=120m -coverprofile=coverage.txt

generate: install-packer-sdc
	go generate ./...
	rm -rf .docs
	packer-sdc renderdocs -src "docs" -partials docs-partials/ -dst ".docs/"
	./.web-docs/scripts/compile-to-webdocs.sh "." ".docs" ".web-docs" "hashicorp"
	rm -r ".docs"
