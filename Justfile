fmt:
    goimports -w -local github.com/sarrrrry/gh-mrepo .
    gofumpt -w .

lint:
    golangci-lint run ./...

test:
    go test ./...

vet:
    go vet ./...

build:
    go build -o gh-mrepo .

# テスト・静的解析・ビルドを一括実行
check: test vet lint build
