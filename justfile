# lightweight-charts justfile
# Browser automation recipes

# Load .env file automatically for all recipes
set dotenv-load

# Show available commands
help:
	@echo "Browser Automation Commands"
	@echo ""
	@echo "  just start-browser       Start Chrome with remote debugging"
	@echo "  just start-browser-logs  Start Chrome with debugging and console logs"
	@echo "  just crash-report        Bundle latest tv_controller/chromium diagnostics"
	@echo "  just run-researcher      Build and run passive TradingView researcher"
	@echo "  just run-tv-controller   Build and run Huma control API"
	@echo "  just run-tv-controller-with-browser  Launch browser + controller in one command"
	@echo "  just run-tv-multi-controller         Build and run multi-window control API"
	@echo "  just run-tv-multi-controller-with-browser  Launch browser + multi controller"
	@echo "  just test-integration    Run integration tests (live server)"
	@echo "  just test-integration-multi [chart=rzWLrz7t]  Run tests against multi-controller"
	@echo ""

# Start Chrome with debugging and console logs
start-browser:
    ./scripts/start-chromium.sh --with-logs

# Bundle crash diagnostics
crash-report:
    ./scripts/collect-crash-report.sh

# Build and run the passive researcher
run-researcher:
    go build -o ./bin/researcher ./cmd/researcher && ./bin/researcher

# Build and run the Huma controller API
run-tv-controller:
    go build -o ./bin/tv_controller ./cmd/tv_controller && ./bin/tv_controller

# Build and run the controller with auto-launched browser
run-tv-controller-with-browser:
    CONTROLLER_LAUNCH_BROWSER=true go build -o ./bin/tv_controller ./cmd/tv_controller && CONTROLLER_LAUNCH_BROWSER=true ./bin/tv_controller

# Build and run the multi-window controller API
run-tv-multi-controller:
    go build -o ./bin/tv_multi_controller ./cmd/tv_multi_controller && ./bin/tv_multi_controller

# Build and run the multi-window controller with auto-launched browser
run-tv-multi-controller-with-browser:
    CONTROLLER_LAUNCH_BROWSER=true go build -o ./bin/tv_multi_controller ./cmd/tv_multi_controller && CONTROLLER_LAUNCH_BROWSER=true ./bin/tv_multi_controller

# Run integration tests (requires running browser + tv_controller)
test-integration:
    go test -tags integration -v -count=1 ./test/integration/...

# Run integration tests against the multi-window controller on a specific chart
test-integration-multi chart="rzWLrz7t":
    TV_CONTROLLER_URL=http://127.0.0.1:8189 TV_TEST_CHART_ID={{chart}} go test -tags integration -v -count=1 ./test/integration/...
