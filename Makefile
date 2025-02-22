test_cmd := go test $(shell go list ./... | grep -v "dev")

.PHONY: test-no-race
test-no-race:
	@go clean -testcache
	@$(test_cmd) --cover

.PHONY: test-race
test-race:
	@go clean -testcache
	@$(test_cmd) --cover --race

.PHONY: test
test:
	@go mod tidy
	make test-no-race
	make test-race
