# common.makefile

# Updated: <2024-01-16 17:00:10 david.hisel>

VERSION := $(shell if [ -f VERSION ]; then cat VERSION; else printf "v0.0.1"; fi)
NEXTVERSION := $(shell echo "$(VERSION)" | awk -F. '{print $$1"."$$2"."$$3+1}')

help: ## show help
	@echo "The following build targets have help summaries:"
	@gawk 'BEGIN{FS=":.*[#][#]"} /[#][#]/ && !/^#/ {h[$$1":"]=$$2}END{n=asorti(h,d);for (i=1;i<=n;i++){printf "%-26s%s\n", d[i], h[d[i]]}}' $(MAKEFILE_LIST)
	@echo

versionbump:  ## increment BUILD number in VERSION file
	echo "$(VERSION)" | awk -F. '{print $$1"."$$2"."$$3+1}' > VERSION

.PHONY: help versionbump

vardump::  ## echo make variables
	@echo "common.makefile: VERSION: $(VERSION)"
	@echo "common.makefile: NEXTVERSION: $(NEXTVERSION)"

clean:: ## clean ephemeral build resources

realclean:: clean  ## clean all resources that can be re-made

.PHONY: vardump clean realclean
