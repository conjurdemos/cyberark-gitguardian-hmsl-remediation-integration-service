# docs.makefile

# Updated: <2023-10-24 00:57:48 david.hisel>

# https://github.com/plantuml/plantuml/releases
PLANTUML_JAR =  $(BINDIR)/plantuml.jar
PLANTUML_DEFAULT_URL = https://github.com/plantuml/plantuml/releases/download/v1.2023.12/plantuml.jar
PLANTUML_API_URL = https://api.github.com/repos/plantuml/plantuml/releases/latest

DOT := $(shell command -v dot 2> /dev/null)
NPM := $(shell command -v npm 2> /dev/null)
DOCTOC_ARGS = --entryprefix '*'

NPMDIR = $(shell npm root)

npm-installed:
ifndef NPM
	$(error "npm is not available, please install npm")
endif

dot-installed:
ifndef DOT
	$(error "dot is not available, please install graphviz")
endif


docs: $(DOCFILES) | dot-installed doctoc $(PLANTUML_JAR) ## process DOCFILES files using plantuml (requires graphviz)
	$(foreach var,$(DOCFILES),if ! git diff --exit-code $(var)>/dev/null; then java -jar $(PLANTUML_JAR) -tsvg $(var); fi;)
	$(foreach var,$(DOCFILES),if ! git diff --exit-code $(var)>/dev/null; then npm exec -- doctoc $(DOCTOC_ARGS) $(var); fi;)

# Determine plantuml.jar download URL; do it this way in case github
# returns api-rate limit msg instead of json; set the download url
# value from the api call JSON response or the default value set above
PLANTUML_DOWNLOAD_URL = $(shell curl -s $(PLANTUML_API_URL) 2>/dev/null |jq -er '.assets[]|select(.name=="plantuml.jar")|.browser_download_url' 2>/dev/null|| echo "$(PLANTUML_DEFAULT_URL)")

$(BINDIR):
	@mkdir -p $(BINDIR)

$(PLANTUML_JAR): | $(BINDIR)
	curl -sL $(PLANTUML_DOWNLOAD_URL) -o $(PLANTUML_JAR)


# <https://github.com/thlorenz/doctoc>
doctoc: npm-installed
	npm list doctoc >/dev/null || npm install doctoc

# <https://www.npmjs.com/package/markdown-it>
markdown-it: | $(PLANTUML_JAR) npm-installed  ## install the markdown-it tool with plantuml support
	npm list markdown-it >/dev/null || npm install markdown-it
	npm list markdown-it-meta-header >/dev/null || npm install markdown-it-meta-header
	npm list markdown-it-plantuml-ex >/dev/null || npm install markdown-it-plantuml-ex
	cp $(PLANTUML_JAR) $(NPMDIR)/markdown-it-plantuml-ex/lib/plantuml.jar


ifeq ($(origin DOCFILE_HTML_TARGETS),undefined)
DOCFILE_HTML_TARGETS := $(foreach var,$(DOCFILES),$(var).html)
endif

html: $(DOCFILE_HTML_TARGETS)  ## build html docs from markdown DOCFILES; process .md to .md.html

$(STATICDIR)%.md.html : %.md | markdown-it
	@mkdir -p $(STATICDIR)
	npm exec -- markdown-it -o $@ $<

vardump::
	@echo "docs.makefile: NPMDIR: $(NPMDIR)"
	@echo "docs.makefile: PLANTUML_JAR: $(PLANTUML_JAR)"
	@echo "docs.makefile: PLANTUML_API_URL: $(PLANTUML_API_URL)"
	@echo "docs.makefile: PLANTUML_DEFAULT_URL : $(PLANTUML_DEFAULT_URL)"
	@echo "docs.makefile: PLANTUML_DOWNLOAD_URL: $(PLANTUML_DOWNLOAD_URL)"
	@echo "docs.makefile: DOCFILES: $(DOCFILES)"
	@echo "docs.makefile: DOCFILE_HTML_TARGETS: $(DOCFILE_HTML_TARGETS)"

clean::
	$(foreach var,$(DOCFILES),rm -f $(var).orig.*;)
	$(foreach var,$(DOCFILES),rm -f $(var).toc.*;)

realclean::
	rm -rf $(PLANTUML_JAR)
	rm -rf $(DOCFILE_HTML_TARGETS)
