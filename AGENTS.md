# Agent Instructions (iot-app-template-go)

This file contains specific instructions for AI agents working in this repository. 
If you operate in this codebase, you must adhere to the following guidelines.

## Architecture Context
This repository serves as a **boilerplate template** for building robust, stateless IoT microservices (bridges) in the homelab. It acts as the standard implementation of the "Go-Microservice-Standard" defined by the IoT Engineer Agent.

- **`cmd/`**: CLI setup using `cobra`. Defines configuration flags (container args) and initializes the application.
- **`internal/mqtt/`**: MQTT client logic. Handles reconnections, Home Assistant Auto-Discovery, and state publishing.
- **`internal/source/`**: A template package for interacting with external hardware/APIs.
- **`internal/service/`**: The core business logic. Periodically polls the `source` and publishes to `mqtt`. Connects dependencies via injection.

## Strict Rules
1. **Four-Eyes-Principle**: You must **never push directly to `main`**. All features and fixes must be developed on a dedicated branch and merged via Pull Request (`gh pr create`).
2. **Container Arguments Only**: Configuration must only be handled via CLI flags (Cobra). Do not introduce environment variables or config files. 
3. **Stateless First**: The template must remain completely stateless.
4. **Logging**: Use Go's native structured logging `log/slog`.
5. **No specific domain logic**: This is a template. Do not add logic specific to a certain heat pump, weather API, or smart plug. Keep the `source` package generic or as a dummy implementation.
