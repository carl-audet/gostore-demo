run:
	@set -a; [ -f .env.otel ] && . ./.env.otel; set +a; \
	if [ -n "$$DT_INGEST_TOKEN" ] && [ -z "$$OTEL_EXPORTER_OTLP_HEADERS" ]; then \
		export OTEL_EXPORTER_OTLP_HEADERS="Authorization=Api-Token $$DT_INGEST_TOKEN"; \
	fi; \
	export OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE="$${OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE:-delta}"; \
	vcs_url=$$(git remote get-url origin 2>/dev/null | sed -e 's#^git@\([^:/]*\)[:/]#https://\1/#' -e 's#\.git$$##'); \
	vcs_rev=$$(git rev-parse HEAD 2>/dev/null); \
	[ -n "$$vcs_url" ] && export OTEL_RESOURCE_ATTRIBUTES="$${OTEL_RESOURCE_ATTRIBUTES:+$$OTEL_RESOURCE_ATTRIBUTES,}vcs.repository.url.full=$$vcs_url"; \
	[ -n "$$vcs_rev" ] && export OTEL_RESOURCE_ATTRIBUTES="$${OTEL_RESOURCE_ATTRIBUTES:+$$OTEL_RESOURCE_ATTRIBUTES,}vcs.ref.head.revision=$$vcs_rev"; \
	go run .

build:
	go build -o bin/gostore .

test:
	go test ./...
