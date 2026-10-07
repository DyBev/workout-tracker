CGO_ENABLED := 0
CMD_DIR := ./functions
BUILD_DIR := ./build
GOOS := linux
GOARCH := arm64
CGO_ENABLED := 0
GO_BUILD_FLAGS := -tags lambda.norpc

test:
	CGO_ENABLED=$(CGO_ENABLED) go test ./...

.PHONY: list
list:
	@echo "Available Lambda functions:"
	@find "$(CMD_DIR)" -mindepth 1 -maxdepth 1 -type d -exec basename {} \;

.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf "$(BUILD_DIR)" *.zip
	@echo "✅ Clean complete"

# Usage: make build-myLambda
.PHONY: $(shell find $(CMD_DIR) -mindepth 1 -maxdepth 1 -type d -exec basename {} \;)
$(shell find $(CMD_DIR) -mindepth 1 -maxdepth 1 -type d -exec basename {} \;):
	@FUNCTION=$@ \
	&& CGO_ENABLED=0 go test ./functions/$$FUNCTION/handler -coverprofile=coverage.out \
	&& COVERAGE=$$(go tool cover -func=coverage.out | grep "total" | awk '{print $$3}' | tr -d '%') \
	&& if awk 'BEGIN {exit !($$COVERAGE > 95)}'; then \
		echo "Coverage is $${COVERAGE}% — expected above 95%"; \
		exit 1; \
	fi \
	&& echo "🚀 Building $$FUNCTION..." \
	&& mkdir -p "$(BUILD_DIR)/$$FUNCTION" \
	&& CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GO_BUILD_FLAGS) -o "$(BUILD_DIR)/$$FUNCTION/bootstrap" "$(CMD_DIR)/$$FUNCTION" \
	&& chmod +x "$(BUILD_DIR)/$$FUNCTION/bootstrap" \
	&& (cd "$(BUILD_DIR)/$$FUNCTION" && zip -r "../$$FUNCTION.zip" bootstrap > /dev/null) \
	&& echo "✅ Build complete: $(BUILD_DIR)/$$FUNCTION.zip"
