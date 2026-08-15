.PHONY: all build install test clean

TARGET ?= tmux-omni
INSTALL_DIR ?= $(HOME)/.config/tmux/scripts

all: build

build:
	go build -ldflags="-s -w" -o $(TARGET) .

install: build
	mkdir -p $(INSTALL_DIR)
	install -m 755 $(TARGET) $(INSTALL_DIR)/$(TARGET)
	@echo "Installed $(TARGET) to $(INSTALL_DIR)/$(TARGET)"

test:
	go test -v ./...

clean:
	rm -f $(TARGET)
