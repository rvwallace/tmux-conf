.PHONY: all build install test clean

TARGET ?= tmux-omni
INSTALL_DIR ?= $(HOME)/.config/tmux/scripts

all: build

build: $(TARGET)

$(TARGET): $(wildcard *.go) go.mod go.sum Makefile
	go build -ldflags="-s -w" -o "$@" .

install: build
	mkdir -p "$(INSTALL_DIR)"
	install -m 755 "$(TARGET)" "$(INSTALL_DIR)/$(TARGET)"
	@echo "Installed $(TARGET) to $(INSTALL_DIR)/$(TARGET)"

test:
	go test -v ./...

clean:
	rm -f $(TARGET)
