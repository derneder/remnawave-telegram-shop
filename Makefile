.PHONY: dc-up dc-down dc-migrate dc-logs i18n-check

dc-up:
	docker compose up -d

dc-down:
	docker compose down

dc-migrate:
	docker compose run --rm migrate

dc-logs:
        docker compose logs -f

i18n-check:
	go test ./tests/unit/translation -run TestTranslationsConsistency
