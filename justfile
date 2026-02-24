# lightweight-charts justfile
# Browser automation recipes

# Load .env file automatically for all recipes
set dotenv-load

# Infer version from git tags for dev builds
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`

# Show available commands
help:
	@echo "Browser Automation Commands"
	@echo ""
	@echo "  just start-browser                   Start Chrome with remote debugging"
	@echo "  just start-browser-logs              Start Chrome with debugging and console logs"
	@echo "  just crash-report                    Bundle latest tv_controller/chromium diagnostics"
	@echo "  just run-researcher                  Build and run passive TradingView researcher"
	@echo "  just run-tv-controller               Build and run Huma control API"
	@echo "  just run-tv-controller-with-browser  Launch browser + controller in one command"
	@echo "  just test-integration                Run integration tests (live server)"
	@echo "  just release-snapshot                Build release binaries locally (no publish)"
	@echo "  just release <tag>                   Cross-compile and publish GitHub release (e.g. just release v1.0.0)"
	@echo ""

# Start Chrome with debugging and console logs
start-browser:
    ./scripts/start-chromium.sh --with-logs

# Bundle crash diagnostics
crash-report:
    ./scripts/collect-crash-report.sh

# Build and run the passive researcher
run-researcher:
    go build -ldflags "-X main.version={{VERSION}}" -o ./bin/researcher ./cmd/researcher && ./bin/researcher

# Build and run the Huma controller API
run-tv-controller:
    go build -ldflags "-X main.version={{VERSION}}" -o ./bin/tv_controller ./cmd/tv_controller && ./bin/tv_controller

# Build and run the controller with auto-launched browser
run-tv-controller-with-browser:
    CONTROLLER_LAUNCH_BROWSER=true go build -ldflags "-X main.version={{VERSION}}" -o ./bin/tv_controller ./cmd/tv_controller && CONTROLLER_LAUNCH_BROWSER=true ./bin/tv_controller

# Run integration tests (requires running browser + tv_controller)
test-integration:
    go test -tags integration -v -count=1 ./test/integration/...

# Build release binaries for all platforms locally (no publish)
release-snapshot:
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p dist
    for os in linux darwin; do
        for arch in amd64 arm64; do
            echo "Building tv_controller ($os/$arch)..."
            GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -ldflags "-s -w -X main.version={{VERSION}}" \
                -o dist/tv_controller_${os}_${arch} ./cmd/tv_controller
            echo "Building researcher ($os/$arch)..."
            GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -ldflags "-s -w -X main.version={{VERSION}}" \
                -o dist/researcher_${os}_${arch} ./cmd/researcher
        done
    done
    echo "Done. Binaries in dist/"
    ls -lh dist/

# Cross-compile and publish a draft GitHub release (e.g. just release v1.0.0)
# Requires: git tag already created, gh auth login completed
release tag:
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p dist
    for os in linux darwin; do
        for arch in amd64 arm64; do
            echo "Building tv_controller ($os/$arch)..."
            GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -ldflags "-s -w -X main.version={{tag}}" \
                -o dist/tv_controller_${os}_${arch} ./cmd/tv_controller
            echo "Building researcher ($os/$arch)..."
            GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -ldflags "-s -w -X main.version={{tag}}" \
                -o dist/researcher_${os}_${arch} ./cmd/researcher
        done
    done
    echo "Creating draft GitHub release {{tag}}..."
    gh release create {{tag}} --draft --generate-notes --title "{{tag}}" \
        dist/tv_controller_linux_amd64 \
        dist/tv_controller_linux_arm64 \
        dist/tv_controller_darwin_amd64 \
        dist/tv_controller_darwin_arm64 \
        dist/researcher_linux_amd64 \
        dist/researcher_linux_arm64 \
        dist/researcher_darwin_amd64 \
        dist/researcher_darwin_arm64
    echo "Draft release created. Review and publish at: https://github.com/Fomo-Driven-Development/MaudeViewTVCore/releases"
