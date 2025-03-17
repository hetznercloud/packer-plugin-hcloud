NAME=hcloud
BINARY=packer-plugin-$(NAME)
ifeq ($(OS), Windows_NT)
# Prevent "Ignoring plugin match packer-plugin-hcloud, no exe extension"
BINARY=packer-plugin-$(NAME).exe
endif
FQN=$(shell go list | sed 's/packer-plugin-//')

COUNT?=1
TEST?=./...

.PHONY: dev

build:
	go build -o $(BINARY)

dev: build
	packer plugins install --path $(BINARY) $(FQN)

test:
	go test -race -count $(COUNT) -v $(TEST) -timeout=3m -coverprofile=coverage.txt

plugin-check: build
	go tool packer-sdc plugin-check $(BINARY)

testacc: dev
	PACKER_ACC=1 go test -count $(COUNT) -v $(TEST) -timeout=120m -coverprofile=coverage.txt

generate:
	go generate ./...
	rm -rf ".docs"
	go tool packer-sdc renderdocs -src "docs" -partials docs-partials/ -dst ".docs/"
	.web-docs/scripts/compile-to-webdocs.sh "." ".docs" ".web-docs" "hashicorp"
	rm -rf ".docs"
