.PHONY: dc-up dc-down dc-migrate dc-logs i18n-check cover

dc-up:
	docker compose up -d

dc-down:
	docker compose down

dc-migrate:
	docker compose run --rm migrate

dc-logs:
        docker compose logs -f

i18n-check:
       go test ./tests -run TestTranslationsConsistency

cover:
       go test ./tests -cover -coverpkg=./... -coverprofile=coverage.out
       go tool cover -func=coverage.out
