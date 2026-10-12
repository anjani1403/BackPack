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
- [ ] **Repo Setup:** Initialize the project structure (directories for `src/backpack`, `tests`, `docs`, etc.).
- [ ] **Dependency Management:** Set up `pyproject.toml` (if Python) or `go.mod` (if Go).
- [ ] **CI/CD Pipeline:** Create a basic GitHub Actions workflow for linting and running tests.
- [ ] **Core Interfaces:** Define the `StorageBackend` interface/contract.

### Phase 1: MVP Core
- [ ] **CLI Skeleton:** Implement the CLI entry point with commands: `backup`, `restore`, and `test-connection`.
- [ ] **Config & Flags:** Implement configuration loading (parsing CLI flags and environment variables).
- [ ] **Local Storage Backend:** Implement the `Local` storage adapter that satisfies the `StorageBackend` interface (upload/download to local filesystem).
- [ ] **Structured Logging:** Implement the logger to record start/end times, duration, status, and size.

---

## 👤 Rajat's Tasks (Database & Orchestration)
*Focus: Adapters, Backup/Restore Pipeline, and Environment.*

### Phase 0: Foundation
- [ ] **Testing Environment:** Create the `docker-compose.yml` with a PostgreSQL container (and MinIO for later) to serve as the local test target.
- [ ] **Core Interfaces:** Define the `DatabaseAdapter` interface/contract (methods: `test_connection`, `backup`, `restore`).

### Phase 1: MVP Core
- [ ] **PostgreSQL Adapter:** Implement the logic to wrap `pg_dump` for backups and `pg_restore` for restores.
- [ ] **Connection Testing:** Implement the `test_connection` logic for PostgreSQL (using `psycopg` or `pgx`).
- [ ] **Pipeline Orchestrator (Backup):** Create the pipeline that streams the output of `pg_dump` -> applies `gzip` compression -> passes the stream to the storage backend.
- [ ] **Pipeline Orchestrator (Restore):** Create the pipeline that pulls from storage -> decompresses `gzip` -> feeds into `pg_restore`.

---

## 🔄 Integration & Review (Together)
Once both have completed their Phase 1 tasks:
1. **Connect CLI to Pipeline:** Wire Anjani's CLI commands to trigger Rajat's pipeline orchestrator.
2. **End-to-End Test:** Run a full backup and restore against the Docker Compose PostgreSQL instance using the local storage backend.
3. **Review & Refactor:** Ensure proper error handling (e.g., failed dump cleans up partial files) and test coverage before moving to Phase 2.
