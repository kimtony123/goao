# Makefile

.PHONY: test test-unit test-integration test-e2e test-cover benchmark clean

# Run all tests
test: test-unit test-integration

# Run unit tests only
test-unit:
	go test -v ./... -race

# Run integration tests (mocked HTTP)
test-integration:
	go test -v ./... -tags=integration -race

# Run E2E tests (requires AO testnet)
test-e2e:
	go test -v ./... -tags=e2e -timeout=5m

# Run tests with coverage
test-cover:
	go test -v ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

# Run benchmarks
benchmark:
	go test -v ./... -bench=. -benchmem -cpuprofile=cpu.prof

# Run specific package tests
test-signer:
	go test -v ./signer/... -race

test-dataitem:
	go test -v ./dataitem/... -race

test-encrypt:
	go test -v ./encrypt/... -race

test-compute:
	go test -v ./compute/... -race

# Clean test artifacts
clean:
	rm -f coverage.out coverage.html cpu.prof test-results.json report.xml

# CI/CD pipeline
ci: test-unit test-cover benchmark
	@echo "✅ All tests passed!"
