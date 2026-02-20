TARGET := specter


.PHONY: all build clean

all:  build


build:
	@echo "Building $(TARGET)..."
	@go build -o ./build/$(TARGET) .
	@echo "[+] Done."

clean:
	@go clean
	@rm -f ./build/$(TARGET)