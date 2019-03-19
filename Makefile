UNAME := $(shell uname)
checkSystem:
ifeq ($(OS),Windows_NT)
	go test .\...
endif
ifeq ($(UNAME),Darwin)
	go test ./...
	@echo "Run in an OS X environment"
else
	go test ./...
	@echo "Run in a Linux environment"
endif
