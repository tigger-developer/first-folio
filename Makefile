INSTALL_DIR ?= $(HOME)/.local/bin
BUILD_DIR ?= $(CURDIR)/dist
HOMEBREW_TAP_DIR ?= $(HOME)/code/tigoss/homebrew-tap
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
VERSION ?= $(patsubst v%,%,$(CURRENT_VERSION))
RELEASE_VERSION ?= $(basename $(CURRENT_VERSION)).$(shell expr $(patsubst .%,%,$(suffix $(CURRENT_VERSION))) + 1)
LDFLAGS := -X github.com/tigger-developer/first-folio/internal/app.Version=$(VERSION)

.PHONY: build install uninstall test lint check-release-deps release sync

build:
	@mkdir -p "$(BUILD_DIR)"
	@go build -trimpath -ldflags "$(LDFLAGS)" -o "$(BUILD_DIR)/folio" ./cmd/folio

install: build
	@mkdir -p "$(INSTALL_DIR)"
	@ln -sf "$(BUILD_DIR)/folio" "$(INSTALL_DIR)/folio"
	@echo "Linked $(INSTALL_DIR)/folio -> $(BUILD_DIR)/folio"

uninstall:
	@test ! -L "$(INSTALL_DIR)/folio" || unlink "$(INSTALL_DIR)/folio"
	@echo "Removed $(INSTALL_DIR)/folio"

lint:
	@go vet ./...

test:
	@go test ./...

check-release-deps:
	@command -v go >/dev/null
	@command -v typst >/dev/null
	@command -v pandoc >/dev/null

sync:
	@git add --all
	@git commit -m "sync: $$(date +%Y-%m-%d)" || true
	@git pull --rebase
	@git push

release: check-release-deps
ifneq ($(SKIP_TESTS),1)
	@$(MAKE) test
else
	@echo "Skipping release regression tests (SKIP_TESTS=1)"
endif
	@git tag -a "$(RELEASE_VERSION)" -m "$(RELEASE_VERSION)"
	@git push
	@git push origin "$(RELEASE_VERSION)"
	@go run ./cmd/update-homebrew --publish-tap "$(HOMEBREW_TAP_DIR)" "$(patsubst v%,%,$(RELEASE_VERSION))"
