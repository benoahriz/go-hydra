# go-hydra

`go-hydra` is a Go HTTP service prototype that demonstrates one global API surface where each endpoint behaves like a function: JSON input arrives over HTTP, is piped to a Docker container over stdin, and the container's stdout is returned directly in the HTTP response.

The core idea is to use Docker as a function execution engine for a lightweight Function-as-a-Service pattern that can fan out horizontally for scale.

Current examples include:

- text processing (lowercase to uppercase)
- URL-to-PDF rendering (post a URL, let a container fetch and render, return PDF bytes)

## Local Run (prototype)

1. Ensure Docker is running and the local socket is available.
2. Install Go dependencies.
3. Start the server.

Example:

```bash
go run .
```

## Design Docs Workflow

System design docs are scaffolded under `docs/design/`.

- They are currently ignored in git via `.gitignore` while drafting.
- To start tracking them, remove the `docs/design/` line from `.gitignore`.
- After unignoring, add and commit the folder normally.

## Notes

- Current storage for todos is in-memory.
- Docker-based execution is local-first and intended to demonstrate a container-per-function architecture.
