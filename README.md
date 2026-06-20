# Mutt

An open-source error tracking system for monitoring crashes, handling recovery, and sending real-time alerts.

## Why Mutt

Most error tracking tools are either expensive or lack transparency. Mutt is built to give you full visibility into your application's health — know what's broken, when it broke, and whether it's been fixed.

## What It Does

- **Error Grouping** — Automatically clusters similar errors using fingerprint hashing
- **Real-time Ingestion** — Capture errors from your apps via SDK
- **Status Tracking** — Mark errors as critical, recovered, or resolved
- **Per-Project Alerts** — Toggle notifications on/off per project
- **API Key Auth** — Secure SDK ingestion with hashed API keys

## Status

> **In Development** — Core backend is being built. SDK for Go, Javascript and other languages are coming soon.

## Tech Stack

- Go + Fiber
- PostgreSQL(NeonDB) + GORM
- Redis (Upstash)
- JWT Authentication

## License

[AGPL-3.0](LICENSE)

## Contributing

Contributions are welcome from anyone. Feel free to open issues or submit pull requests.

To understand the system architecture, data flows, and design decisions, please check the [Architecture Diagram](diagram/diagram.md).
