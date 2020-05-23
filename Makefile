.DEFAULT_GOAL=dist

GOOS:=$(shell go env GOOS)
GOARCH:=$(shell go env GOARCH)

DIST_DIR=dist
BIN_DIR=$(GOOS)-$(GOARCH)

EXAMPLES_DIR=example

INSTALL_DIR=/usr/local/bin
SYSTEMD_DIR=/etc/systemd/system

MAIN=main.go
BUILD_CMD=GO111MODULE=on go build

SPIGOT_EXE=spigot
SPIGOT_SERVICE_FILE=spigot.service


.PHONY: dist install clean
dist:
	mkdir -p $(DIST_DIR)/$(BIN_DIR)
	$(BUILD_CMD) -o $(DIST_DIR)/$(BIN_DIR)/$(SPIGOT_EXE) $(MAIN)

install: dist
	install $(DIST_DIR)/$(BIN_DIR)/$(SPIGOT_EXE) $(INSTALL_DIR)/$(SPIGOT_EXE)

systemd: install
	cp -v $(EXAMPLES_DIR)/$(SYSTEMD_DIR)/$(SPIGOT_SERVICE_FILE) $(SYSTEMD_DIR)/$(SPIGOT_SERVICE_FILE)

clean:
	rm -rf $(DIST_DIR)
