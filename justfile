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
	@echo "  just mapper-static-only  Run static analysis stage only"
	@echo "  just mapper-runtime-only Run runtime probes stage only"
	@echo "  just mapper-correlate    Run correlation stage only"
	@echo "  just mapper-report       Run reporting stage only"
	@echo "  just mapper-validate     Validate mapper artifacts"
	@echo "  just mapper-smoke        Run smoke sequence + docs report"
	@echo "  just mapper-full         Run full mapper pipeline"
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

# Mapper stage: static analysis only
mapper-static-only:
    ./scripts/mapper-static-only.sh

# Mapper stage: runtime probes only
mapper-runtime-only:
    ./scripts/mapper-runtime-only.sh

# Mapper stage: correlation only
mapper-correlate:
    ./scripts/mapper-correlate.sh

# Mapper stage: reporting only
mapper-report:
    ./scripts/mapper-report.sh

# Mapper stage: artifact validation
mapper-validate:
    ./scripts/mapper-validate.sh

# Mapper smoke sequence with baseline report
mapper-smoke:
    ./scripts/mapper-smoke.sh

# Mapper pipeline: all stages
mapper-full:
    ./scripts/mapper-full.sh
