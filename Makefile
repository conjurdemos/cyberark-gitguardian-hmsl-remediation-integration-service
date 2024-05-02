# Makefile -*- Makefile -*-

# Updated: <2024/03/19 15:52:05>

# LICENSE

BINDIR := ./bin
STATICDIR := ./static

DOCFILES := $(wildcard *.md)
DOCFILE_HTML_TARGETS := $(addprefix $(STATICDIR)/, $(addsuffix .html, $(basename $(DOCFILES)))) gen-brimstone-doc

BRIMSTONE_OPENAPI_SPEC := api/brimstone.yaml

DATADIR := ./data

.PHONY: Makefile scripts/common.makefile
include scripts/common.makefile
export

all: build-all-bins gen-brimstone-doc docs html  ## build all bins, openapi spec html, process DOCFILES

GO := $(shell command -v go 2> /dev/null)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"


##
## Documentation targets
##
.PHONY: gen-brimstone-doc
gen-brimstone-doc: $(STATICDIR)/brimstone-spec/index.html   ## generate brimstone HTML doc from brimstone openapi spec

$(STATICDIR)/brimstone-spec/index.html: api/brimstone.yaml | redocly-cli
	@mkdir -p $(STATICDIR)/brimstone-spec
	$(REDOCLY_CLI) build-docs ./api/brimstone.yaml -o $(STATICDIR)/brimstone-spec/index.html

##
## OPENAPI code gen
## 

pkg/brimstone/brimstone.gen.go: $(BRIMSTONE_OPENAPI_SPEC) | oapi-codegen-installed
	@mkdir -p pkg/brimstone
	command -v $(OAPI_CODEGEN)
	$(OAPI_CODEGEN) -generate types,server,client -package brimstone api/brimstone.yaml > pkg/brimstone/brimstone.gen.go


api/hasmysecretleaked-openapi.json:
	@mkdir -p api
	curl -s https://api.hasmysecretleaked.com/openapi.json -o api/hasmysecretleaked-openapi.json

pkg/hasmysecretleaked/hasmysecretleaked.gen.go: api/hasmysecretleaked-openapi.json | oapi-codegen-installed
	@mkdir -p pkg/hasmysecretleaked
	$(OAPI_CODEGEN) -generate types,client -package hasmysecretleaked api/hasmysecretleaked-openapi.json > pkg/hasmysecretleaked/hasmysecretleaked.gen.go

# GG API docs - https://api.gitguardian.com/docs
api/gitguardian-openapi.json: 
	@mkdir -p api
	#TODO: wait for GG Openapi spec to get fixed; currently it fails linting with redocly and oapi-codegen fails
	curl -s -o api/gitguardian-openapi-STAGE.json  https://api.gitguardian.com/v1/openapi.json
	$(REDOCLY_CLI) lint --format summary api/gitguardian-openapi-STAGE.json && mv api/gitguardian-openapi-STAGE.json api/gitguardian-openapi.json

pkg/gitguardian/gitguardian.gen.go: api/gitguardian-openapi.json | oapi-codegen-installed
	@mkdir -p pkg/gitguardian
	$(OAPI_CODEGEN) -generate types,client -package gitguardian api/gitguardian-openapi.json > pkg/gitguardian/gitguardian.gen.go

##
## Build binaries
##

.PHONY: build-brimstone
build-brimstone: $(BINDIR)/brimstone  ## build the brimstone server BINDIR/brimstone

$(BINDIR)/brimstone: VERSION pkg/brimstone/brimstone.go pkg/brimstone/brimstone.gen.go pkg/hasmysecretleaked/client.go pkg/hasmysecretleaked/hasmysecretleaked.gen.go $(BRIMSTONE_OPENAPI_SPEC)
	$(GO) build -o $(BINDIR)/brimstone $(LDFLAGS) cmd/brimstone/main.go

.PHONY: build-brimstone-cp
build-brimstone-cp: $(BINDIR)/brimstone-cp  ## build the brimstone server BINDIR/brimstone-cp

$(BINDIR)/brimstone-cp: VERSION pkg/brimstone/brimstone.go pkg/brimstone/brimstone.gen.go pkg/hasmysecretleaked/client.go pkg/hasmysecretleaked/hasmysecretleaked.gen.go $(BRIMSTONE_OPENAPI_SPEC)
	$(GO) build -o $(BINDIR)/brimstone-cp $(LDFLAGS) cmd/brimstone-cp/main.go

.PHONY: build-hailstone
build-hailstone: $(BINDIR)/hailstone  ## build the hailstone loader BINDIR/hailstone

$(BINDIR)/hailstone: VERSION pkg/brimstone/brimstone.gen.go $(BRIMSTONE_OPENAPI_SPEC)
	$(GO) build -o $(BINDIR)/hailstone $(LDFLAGS) cmd/hailstone/main.go

.PHONY: build-hmsl-client
build-hmsl-client: $(BINDIR)/hmsl-client  ## build the hmsl client BINDIR/hmsl-client 

$(BINDIR)/hmsl-client: VERSION cmd/hmslclient/main.go pkg/hasmysecretleaked/client.go pkg/hasmysecretleaked/hasmysecretleaked.gen.go
	$(GO) build -o $(BINDIR)/hmsl-client $(LDFLAGS) cmd/hmslclient/main.go

.PHONY: build-gg-client
build-gg-client: $(BINDIR)/gg-client ## build the gg client BINDIR/gg-client

$(BINDIR)/gg-client: VERSION cmd/ggclient/main.go pkg/gitguardian/gitguardian.gen.go
	$(GO) build -o $(BINDIR)/gg-client $(LDFLAGS) cmd/ggclient/main.go

.PHONY: build-pam-client
build-pam-client: $(BINDIR)/pam-client ## build the gg client BINDIR/gg-client

$(BINDIR)/pam-client: VERSION cmd/pamclient/main.go pkg/privilegeaccessmanager/privilegeaccessmanager.go pkg/utils/utils.go
	$(GO) build -o $(BINDIR)/pam-client $(LDFLAGS) cmd/pamclient/main.go

.PHONY: build-cp-client
build-cp-client: $(BINDIR)/cp-client ## build the gg client BINDIR/gg-client

$(BINDIR)/cp-client: VERSION cmd/cpclient/main.go
	$(GO) build -o $(BINDIR)/cp-client $(LDFLAGS) cmd/cpclient/main.go

.PHONY: build-all-bins
build-all-bins: build-brimstone build-hailstone build-brimstone-cp build-hmsl-client build-gg-client build-pam-client build-randchar build-cp-client

##
## Helpers
##
initialize-cockroach-dev: | start-cockroach-dev  ## Start and Initialize CockraochDB dev instance
	cockroach sql --insecure --user=root --host=127.0.0.1 -f sql/00_initialize.sql

start-cockroach-dev:  ## Start local cockroach db dev instance
	mkdir -p $(DATADIR)
	pgrep cockroach || (cd $(DATADIR) && cockroach start-single-node --insecure --listen-addr=localhost:26257 --http-addr=localhost:9090 --background)

stop-cockroach-dev:  ## Stop local cockroach db dev instance
	pgrep cockroach && kill -TERM $$(pgrep cockroach)

build-randchar: $(BINDIR)/randchar  ## build the randchar server BINDIR/randchar

$(BINDIR)/randchar: VERSION cmd/randchar/main.go
	$(GO) build -o $(BINDIR)/randchar $(LDFLAGS) cmd/randchar/main.go

.PHONY: initialize-cockroach-dev start-cockroach-dev stop-cockroach-dev build-randchar

clean::
	rm -f pkg/brimstone/brimstone.gen.go pkg/hasmysecretleaked/hasmysecretleaked.gen.go
	rm -f $(BINDIR)/brimstone
	rm -f $(BINDIR)/brimstone-cp
	rm -f $(BINDIR)/hailstone
	rm -f $(BINDIR)/hmsl-client
	rm -f $(BINDIR)/gg-client
	rm -f $(BINDIR)/pam-client
	rm -f $(BINDIR)/cp-client
	rm -f $(BINDIR)/randchar
	rm -f $(STATICDIR)/brimstone-spec/index.html

vardump::
	@echo "Makefile: BINDIR: $(BINDIR)"
	@echo "Makefile: DATADIR: $(DATADIR)"
	@echo "Makefile: LDFLAGS: $(LDFLAGS)"
	@echo "Makefile: OAPI_CODEGEN: $(OAPI_CODEGEN)"

.PHONY: vardump clean realclean

.PHONY: scripts/docs.makefile scripts/openapi.makefile
include scripts/docs.makefile
include scripts/openapi.makefile
