fmt:
    goimports -w -local github.com/sarrrrry/gh-mrepo .
    gofumpt -w .

lint:
    golangci-lint run ./...

test:
    go test ./...

build:
    go build -o gh-mrepo .
