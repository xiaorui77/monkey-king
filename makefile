version ?= 0.1.0

# init project path
HOME_DIR 		:= $(shell pwd)
OUT_DIR  		:= $(HOME_DIR)/output

APP 	:= monkey-king
VERSION := $(version)

# init command params
GO      := $(GO_1_17_BIN)/go


prepare:
	$(GO) env -w GO111MODULE=on
	$(GO) mod download

compile-all: prepare compile-linux compile-darwin compile-windows

compile-linux:
	mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o $(OUT_DIR)/$(APP)-v$(VERSION)-linux-amd64 $(HOME_DIR)/cmd/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -o $(OUT_DIR)/$(APP)-v$(VERSION)-linux-arm64 $(HOME_DIR)/cmd/main.go

compile-darwin:
	mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -o $(OUT_DIR)/$(APP)-v$(VERSION)-darwin-amd64 $(HOME_DIR)/cmd/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -o $(OUT_DIR)/$(APP)-v$(VERSION)-darwin-arm64 $(HOME_DIR)/cmd/main.go

compile-windows:
	mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -o $(OUT_DIR)/$(APP)-v$(VERSION)-windows-amd64 $(HOME_DIR)/cmd/main.go