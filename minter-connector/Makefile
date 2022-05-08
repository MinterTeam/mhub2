all: install

install: go.sum
		go install ./cmd/mhub-minter-connector

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify
proto-gen:
	cd proto && protoc --go_out=../tx_committer --go_opt=paths=source_relative \
        --go-grpc_out=../tx_committer --go-grpc_opt=paths=source_relative \
        tx_committer.proto