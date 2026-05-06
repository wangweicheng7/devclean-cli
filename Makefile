BINARY_NAME ?= cleandev
PKG ?= ./...
FORMULA_PATH ?= homebrew-tap/Formula/cleandev.rb

.PHONY: fmt test build run install-user uninstall-user brew-formula-update brew-install-local sha256-url

fmt:
	go fmt $(PKG)

test:
	go test $(PKG)

build:
	go build -o bin/$(BINARY_NAME) ./cmd/cleandev

run:
	go run ./cmd/cleandev $(ARGS)

install-user:
	mkdir -p "$(HOME)/bin"
	go build -o "$(HOME)/bin/$(BINARY_NAME)" ./cmd/cleandev
	@echo "installed to $(HOME)/bin/$(BINARY_NAME)"

uninstall-user:
	rm -f "$(HOME)/bin/$(BINARY_NAME)"
	@echo "removed $(HOME)/bin/$(BINARY_NAME)"

sha256-url:
	@URL="$${URL:-}"; \
	if [ -z "$$URL" ]; then echo "usage: make sha256-url URL=https://..." >&2; exit 2; fi; \
	tmp="$$(mktemp -t cleandev-sha.XXXXXX)"; \
	curl -L -o "$$tmp" "$$URL"; \
	shasum -a 256 "$$tmp" | awk '{print $$1}'; \
	rm -f "$$tmp"

brew-formula-update:
	@if [ -z "$${TAG:-}" ]; then echo "usage: make brew-formula-update TAG=v0.1.0" >&2; exit 2; fi
	bash scripts/update_formula_from_tag.sh "$${TAG}"

brew-install-local:
	@echo "brew install from local formula file: $(FORMULA_PATH)" >&2
	brew install --formula --build-from-source "$(FORMULA_PATH)"

