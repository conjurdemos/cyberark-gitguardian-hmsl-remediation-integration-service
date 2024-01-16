# openapi.makefile

# Updated: <2024-01-16 16:59:09 david.hisel>

NPM := $(shell command -v npm 2> /dev/null)
NPMDIR := $(shell npm root)
REDOCLY_CLI := $(NPM) exec -- @redocly/cli

npm-installed:
ifndef NPM
	$(error "npm is not available, please install npm")
endif

#####
# Redocly - for linting and generating docs

.PHONY: redocly-cli
redocly-cli: | npm-installed  ## install redocly-cli - USED FOR GENERATING DOCS
	$(NPM) list @redocly/cli@latest >/dev/null || $(NPM) install @redocly/cli@latest

#####
# Go - code generation

OAPI_CODEGEN := $(shell command -v ./bin/oapi-codegen 2> /dev/null)

.PHONY: oapi-codegen-installed
oapi-codegen-installed:
ifndef OAPI_CODEGEN
	$(error "OAPI_CODEGEN is not installed; try 'make oapi-codegen'")
endif

.PHONY: go-installed
go-installed:
ifndef GO
	$(error "GO is not available, please install GO")
endif

.PHONY: oapi-codegen
oapi-codegen: bin/oapi-codegen ## install oapi-codegen tool used for generating go code

bin/oapi-codegen: | go-installed  ## install oapi-codegen tool used for generating go code
	GOBIN=$(shell pwd)/bin $(GO) install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

.PHONY: vardump
vardump::
	@echo "openapi.makefile: NPM: $(NPM)"
	@echo "openapi.makefile: NPMDIR: $(NPMDIR)"
	@echo "openapi.makefile: REDOCLY_CLI: $(REDOCLY_CLI)"
	@echo "openapi.makefile: OAPI_CODEGEN: $(OAPI_CODEGEN)"
