
.PHONY: 
all: help

## license: Adds license header to missing files.
license:
	@echo "  >  \033[32mAdding license headers...\033[0m "
	GO111MODULE=off go get -u github.com/google/addlicense
	addlicense -v -c "Sygma" -f ./scripts/header.txt -y 2023 -ignore ".idea/**"  .

## license-check: Checks for missing license headers
license-check:
	@echo "  >  \033[Checking for license headers...\033[0m "
	GO111MODULE=off go get -u github.com/google/addlicense
	addlicense -check -c "Sygma" -f ./scripts/header.txt -y 2021 -ignore ".idea/**" .

coverage:
	go tool cover -func cover.out | grep total | awk '{print $3}'

test:
	./scripts/tests.sh

genmocks:
	mockgen -source=./chains/evm/listener/events/handlers/deposit.go -destination=./mock/deposit.go -package mock
	mockgen -source=./chains/evm/listener/events/handlers/rotate.go -destination=./mock/rotate.go -package mock
	mockgen -source=./chains/evm/listener/listener.go -destination=./mock/listener.go -package mock
	mockgen -source=./chains/evm/executor/executor.go -destination=./mock/executor.go -package mock
	mockgen -source=./chains/evm/prover/prover.go -destination=./mock/prover.go -package mock


PLATFORMS := linux/amd64 darwin/amd64 darwin/arm64 linux/arm
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o 'build/${os}-${arch}/relayer'; \

build-all: $(PLATFORMS)
