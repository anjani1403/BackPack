# BackPack: Phase 0 & 1 Development Plan

Based on the roadmap and team roles outlined in the project documentation, here is a foolproof, step-by-step plan for the initial stages of development. This plan focuses on **Phase 0 (Setup & Interfaces)** and **Phase 1 (MVP: PostgreSQL, local storage, gzip, logging)**. 

Since the language (Python vs. Go) is yet to be finalized, this plan applies cleanly to both options.

## 🤝 Pre-requisite: Joint Decisions
Before writing any code, both developers must agree on:
1. **Language & Stack:** Choose between Python (Typer/Rich) and Go (Cobra/Viper).
2. **Project Management:** Set up a GitHub Project board to track these tasks.
3. **Branching Strategy:** Agree on a simple flow (e.g., feature branches merging into `main` via PRs).

---

## 👤 Anjani's Tasks (Platform & Infrastructure)
*Focus: CLI, Storage, Configuration, and Observability.*

### Phase 0: Foundation
- [x] **Repo Setup:** Initialize the project structure (directories for `cmd/backpack`, `adapters`, `pipeline`, `storage`, `config`, `logger`).
- [x] **Dependency Management:** Set up `go.mod`.
- [x] **CI/CD Pipeline:** Create a basic GitHub Actions workflow for linting and running tests (`.github/workflows/ci.yml`).
- [x] **Core Interfaces:** Define the `StorageBackend` interface/contract (`storage/storage.go`).

### Phase 1: MVP Core
- [x] **CLI Skeleton:** Implement the CLI entry point with commands: `backup`, `restore`, `list`, and `test-connection` (`cmd/backpack/main.go`).
- [x] **Config & Flags:** Implement configuration loading (parsing CLI flags and environment variables) (`config/config.go`).
- [x] **Local Storage Backend:** Implement the `Local` storage adapter that satisfies the `StorageBackend` interface (upload/download to local filesystem) (`storage/local.go`).
- [x] **Structured Logging:** Implement the logger to record start/end times, duration, status, and size (`logger/logger.go`).

---

## 👤 Rajat's Tasks (Database & Orchestration)
*Focus: Adapters, Backup/Restore Pipeline, and Environment.*

### Phase 0: Foundation
- [x] **Testing Environment:** Create the `docker-compose.yml` with a PostgreSQL container (and MinIO for later) to serve as the local test target.
- [x] **Core Interfaces:** Define the `DatabaseAdapter` interface/contract (methods: `test_connection`, `backup`, `restore`).

### Phase 1: MVP Core
- [x] **PostgreSQL Adapter:** Implement the logic to wrap `pg_dump` for backups and `pg_restore` for restores.
- [x] **Connection Testing:** Implement the `test_connection` logic for PostgreSQL (using `psycopg` or `pgx`).
- [x] **Pipeline Orchestrator (Backup):** Create the pipeline that streams the output of `pg_dump` -> applies `gzip` compression -> passes the stream to the storage backend.
- [x] **Pipeline Orchestrator (Restore):** Create the pipeline that pulls from storage -> decompresses `gzip` -> feeds into `pg_restore`.

---

## 🔄 Integration & Review (Together)
- [x] **Connect CLI to Pipeline:** Wired CLI commands directly to the pipeline orchestrator, adapters, and storage backends.
- [x] **Unit & Component Tests:** Added test suites verifying gzip compression, data integrity, and local storage operations (`pipeline_test.go`, `local_test.go`, `postgres_test.go`).
- [ ] **End-to-End Test:** Run a live backup and restore against the Docker Compose PostgreSQL instance.
- [ ] **Review & Refactor:** Review code quality, error handling, and documentation before moving to Phase 2.
