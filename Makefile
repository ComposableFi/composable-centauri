#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.0.0-rc8
BUILDDIR ?= $(CURDIR)/build
HTTPS_GIT := https://github.com/notional-labs/composable-centauri.git

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += gcc cleveldb
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=pica \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=picad \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" 

ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq ($(LINK_STATICALLY),true)
  ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif


all: install

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/picad

build:
	go build $(BUILD_FLAGS) -o bin/picad ./cmd/picad

docker-build-debug:
	@DOCKER_BUILDKIT=1 docker build -t centauri:debug -f Dockerfile .

lint:
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' -not -name '*.gw.go' | xargs go run mvdan.cc/gofumpt -w .
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' -not -name '*.gw.go' | xargs go run github.com/client9/misspell/cmd/misspell -w
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' -not -name '*.gw.go' | xargs go run golang.org/x/tools/cmd/goimports -w -local github.com/notional-labs/pica
.PHONY: lint

###############################################################################
###                                  Proto                                  ###
###############################################################################

protoVer=0.14.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
containerProtoGen=proto-gen-$(protoVer)
containerProtoFmt=proto-fmt-$(protoVer)

proto-all: proto-format proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-format:
	@echo "Formatting Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./ -not -path "./third_party/*" -name "*.proto" -exec clang-format -i {} \; ; fi

proto-lint:
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against $(HTTPS_GIT)#branch=main

.PHONY: proto-all proto-gen proto-format proto-lint proto-check-breaking 
###                             Interchain test                             ###
###############################################################################

# Executes start chain tests via interchaintest
ictest-start-cosmos:
	cd tests/interchaintest && go test -race -v -run TestStartCentauri .

ictest-validator:
	cd tests/interchaintest && go test -race -v -run TestValidator .

# Executes start chain tests via interchaintest
ictest-start-polkadot:
	cd tests/interchaintest && go test -timeout=25m -race -v -run TestPolkadotcentauriChainStart .

# Executes IBC tests via interchaintest
ictest-ibc:
	cd tests/interchaintest && go test -timeout=25m -race -v -run TestCentauriPicassoIBCTransfer .

# Executes Basic Upgrade Chain tests via interchaintest
ictest-upgrade:
	cd tests/interchaintest && go test -timeout=25m -race -v -run TestCentauriUpgrade .

# Executes all tests via interchaintest after compling a local image as juno:local
ictest-all: ictest-start-cosmos ictest-start-polkadot ictest-ibc ictest-ibc-cosmos ictest-chain-core ictest-pfm ictest-pfm-router ictest-pfm-timeout ictest-miscellaneous ictest-ibc-hooks

# Executes push wasm client tests via interchaintest
ictest-push-wasm:
	cd tests/interchaintest && go test -race -v -run TestPushWasmClientCode .

ictest-ibc-cosmos:
	cd tests/interchaintest && go test -race -v -run TestComposableGaiaIBCTransfer .

ictest-chain-core:
	cd tests/interchaintest && go test -race -v -run TestCoreSDKCommands .

ictest-pfm-timeout:
	cd tests/interchaintest && go test -race -v -run TestTimeoutOnForward .

ictest-pfm:
	cd tests/interchaintest && go test -race -v -run TestPacketForwardMiddleware .

ictest-pfm-router:
	cd tests/interchaintest && go test -race -v -run TestPacketForwardMiddlewareRouter .

ictest-ibc-hooks:
	cd tests/interchaintest && go test -race -v -run TestComposableIBCHooks .

ictest-miscellaneous:
	cd tests/interchaintest && go test -race -v -run TestICTestMiscellaneous .


# Init 2 cosmos chains and setup ibc between them
init-test-interchain: clean-testing-data install
	./scripts/test-upgrade-cosmos-chains.sh

###  Upgrade Test ###
test-upgrade: clean-testing-data
	@echo "Starting upgrade test"
	./scripts/test-upgrade.sh

clean-testing-data:
	@echo "Killing binary and removing previous data"
	echo "stopping picachain..."
	-@pkill picad 2>/dev/null
	-@rm -rf ./mytestnet

	echo "stopping parachain..."
	-@killall parachain-node
	-@killall polkadot
	
	netstat -ltup | grep LISTEN

.PHONY: ictest-start-cosmos ictest-start-polkadot ictest-ibc ictest-push-wasm ictest-all

include contrib/make/release.mk


test-upgrade: clean-testing-data
	@echo "Starting upgrade test"
	./scripts/tweak-test-upgrade.sh



## Scripts for testing sdk 50
init-deps:
	@echo "Installing dependencies"
	bash ./scripts/upgrade/init-deps.sh

localnet-pica:
	@echo "Starting test"
	rm -rf screenlog.0
	-@pkill picad 2>/dev/null
	bash ./scripts/run-node.sh picad
	bash ./scripts/50/store-wasm-code.sh

localnet-parachain:
	@echo "Starting localnet"
	bash ./scripts/upgrade/setup-polkadot-node.sh

relayer-create-clients:
	@echo "Starting relayer"
	bash ./scripts/relayer_hyperspace/create-clients.sh

pica-upgrade:
	@echo "Starting upgrade"
	bash ./scripts/upgrade/upgrade.

relayer-test-cleanup:
	@echo "Cleaning up"
	./scripts/relayer_hyperspace/cleanup.sh
###############################################################################
###                        Integration Tests                                ###
###############################################################################

integration-test-all: init-test-framework \
	test-ibc-hooks

init-test-framework: clean-testing-data install
	@echo "Initializing both blockchains..."
	./scripts/tests/init-test-framework.sh
	./scripts/relayer/relayer-init.sh

test-ibc-hooks:
	@echo "Testing ibc-hooks..."
	./scripts/tests/ibc-hooks/increment.sh
