lint:
	docker run --rm -i -v "$(PWD):/src" -w /src golangci/golangci-lint:v1.64 golangci-lint run ./... -E gofmt -E revive -E bodyclose -E gosec -E unparam -E unconvert -E gocritic -E nestif -E asciicheck -E errorlint -E copyloopvar -E nilerr --timeout=10m
