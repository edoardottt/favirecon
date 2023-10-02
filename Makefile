REPO=github.com/edoardottt/favirecon

remod:
	@rm -rf go.*
	@go mod init ${REPO}
	@go get ./...
	@go mod tidy -v
	@echo "Done."

update:
	@go get -u ./...
	@go mod tidy -v
	@echo "Done."

lint:
	@golangci-lint run

linux:
	@go build -o favirecon ./...
	@sudo mv favirecon /usr/local/bin/
	@echo "Done."

unlinux:
	@sudo rm -rf /usr/local/bin/favirecon
	@echo "Done."

test:
	@go test -race ./...
	@echo "Done."