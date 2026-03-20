# IoT App Template (Go)

This repository serves as a boilerplate for building robust, stateless IoT microservices (bridges) in the homelab. It acts as the standard implementation of the "Go-Microservice-Standard" defined by the IoT Engineer Agent.

## Architecture

The template follows a clean, modular structure inspired by enterprise Go applications (like Deepup's scan-controllerd), but simplified for IoT environments:

* `cmd/`: Contains the CLI setup using `cobra`. Defines configuration flags (container args) and initializes the application.
* `internal/mqtt/`: Contains the MQTT client logic. Handles reconnections, Home Assistant Auto-Discovery, and state publishing.
* `internal/source/`: A template package for interacting with external hardware/APIs (e.g., Modbus TCP, REST, Zigbee). Replace this with your specific protocol.
* `internal/service/`: The core business logic. Periodically polls the `source` and publishes to `mqtt`. Connects dependencies via injection.

## Features

* **Configuration via CLI Arguments**: No hidden environment variables. Everything is configured via explicit flags (e.g., `-mqtt-host`, `-poll-interval`).
* **Structured Logging**: Uses Go's native `log/slog` for structured, level-based logging.
* **Graceful Shutdown**: Listens to OS signals (`SIGTERM`, `SIGINT`) to close connections cleanly before the Pod shuts down.
* **Minimal Footprint**: Uses a Multi-Stage Dockerfile resulting in a `<10MB` Scratch container.
* **CI/CD Built-in**: Includes a GitHub Actions workflow to automatically build and push the image to GHCR on branch merges.

## Usage

1. **Clone this template** (Use the "Use this template" button on GitHub).
2. **Rename imports**: Search for `github.com/dennisschroeder/iot-app-template-go` and replace it with your new repository name.
3. **Implement Source**: Replace `internal/source` with your specific API/hardware logic.
4. **Update Service**: Map the incoming source data to MQTT topics in `internal/service`.

## Local Testing

```bash
go build -o service .
./service -mqtt-host=localhost -poll-interval=10s -log-level=debug
```