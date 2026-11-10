NAME=ansible-navigator
BINARY=packer-plugin-${NAME}
PLUGIN_FQN="$(shell grep -E '^module' <go.mod | sed -E 's/module *//')"

COUNT?=1
TEST?=$(shell go list ./...)
HASHICORP_PACKER_PLUGIN_SDK_VERSION?=$(shell go list -m github.com/hashicorp/packer-plugin-sdk | cut -d " " -f2)

.PHONY: dev test

build:
	@go build -o ${BINARY}

dev:
	@go build -ldflags="-X '${PLUGIN_FQN}/version.VersionPrerelease=dev'" -o '${BINARY}'
	packer plugins install --path ${BINARY} "$(shell echo "${PLUGIN_FQN}" | sed 's/packer-plugin-//')"

test:
	@go test -race -count $(COUNT) $(TEST) -timeout=3m

install-packer-sdc: ## Install packer sofware development command
	@go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@${HASHICORP_PACKER_PLUGIN_SDK_VERSION}

plugin-check: install-packer-sdc build
	@packer-sdc plugin-check ${BINARY}

testacc: dev
	@PACKER_ACC=1 go test -count $(COUNT) -v $(TEST) -timeout=120m

generate: install-packer-sdc
	@go generate ./...
	@rm -rf .docs
	@packer-sdc renderdocs -src "docs" -partials docs-partials/ -dst ".docs/"
	@./.web-docs/scripts/compile-to-webdocs.sh "." ".docs" ".web-docs" "SolomonHD"
	@rm -r ".docs"

# Explicit integration qualification; these targets never run as part of `test`.
ACCEPTANCE_PREPARED_DIR?=/tmp/packer-navigator-acceptance-prepared
ACCEPTANCE_RECORD?=acceptance-record.json
ACCEPTANCE_ARGS?=

.PHONY: acceptance-prepare acceptance-linux-wsl test-acceptance-runner
acceptance-prepare:
	@python3 scripts/linux_wsl_acceptance.py prepare --prepared-dir "$(ACCEPTANCE_PREPARED_DIR)" $(ACCEPTANCE_ARGS)

acceptance-linux-wsl:
	@python3 scripts/linux_wsl_acceptance.py run --prepared-dir "$(ACCEPTANCE_PREPARED_DIR)" --record "$(ACCEPTANCE_RECORD)" $(ACCEPTANCE_ARGS)

test-acceptance-runner:
	@PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts -p 'test_linux_wsl_acceptance.py' -v
