run:
	@set -a; [ -f .env.otel ] && . ./.env.otel; set +a; \
	VCS_URL=$$(git remote get-url origin 2>/dev/null || echo local/demo-gostore); \
	VCS_REV=$$(git rev-parse HEAD 2>/dev/null || echo unknown); \
	OTEL_RESOURCE_ATTRIBUTES="$${OTEL_RESOURCE_ATTRIBUTES:+$${OTEL_RESOURCE_ATTRIBUTES},}vcs.repository.url.full=$${VCS_URL},vcs.ref.head.revision=$${VCS_REV}" \
	go run .

build:
	go build -o bin/gostore .

test:
	go test ./...
