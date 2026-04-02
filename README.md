## Run the app locally

This project is set up to run in Docker for a fast, reproducible dev experience with hot reload and optional Go debugging.

### Prerequisites

- Docker Desktop (with Docker Compose)
- make (preinstalled on macOS; otherwise install via Homebrew)
- Optional: Go (only needed if you want to run without Docker)

### 1. Start the app (Docker, hot reload)

Normal dev run (hot reload via reflex):

```bash
make dev
```

Start in **debug mode** (Delve on port 40000):

```bash
make dev DEBUG=true
```

Stop the stack:

```bash
make stop
```

### 2. VS Code debugging (attach to Delve)

When started with `make dev DEBUG=true`, Delve listens on `127.0.0.1:40000`.

Add a minimal `.vscode/launch.json` and use Attach mode:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Attach to Docker Delve",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "remotePath": "/app",
      "port": 40000,
      "host": "127.0.0.1",
      "showLog": true,
      "trace": "verbose",
      "substitutePath": [
        {
          "from": "${workspaceFolder}/internal",
          "to": "/app/internal"
        },
        {
          "from": "${workspaceFolder}/cmd",
          "to": "/app/cmd"
        }
      ]
    }
  ]
}
```

### 3. Run tests

Tests run locally (outside Docker) but inherit env from `.env` and `.e2e.env` if present:

```bash
make test
```
