# Dev notes (this machine)

- Interactive shell is zsh — quote anything containing `=` in echo/args (zsh word expansion bites).
- Telemetry ingestion into the Dynatrace tenant lags 30–90s (logs land slower than spans). To wait
  for it, poll: `until <check>; do sleep 15; done` — long bare `sleep`s are blocked by the sandbox.
- Run the service with `PORT=3100 make run` (or `PORT=3100 go run .`); kill stragglers with
  `lsof -ti :3100 | xargs kill`.
- The `bluebox` CLI is preconfigured for the dev stack — `bluebox ask --service <name> "<question>"`
  works out of the box (no login needed).
