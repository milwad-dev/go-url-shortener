build:
	docker compose up -d --build

up:
	docker compose up -d

down:
	docker compose down

# go mod edit -require=github.com/milwad-dev/url-shortener/contracts@v0.0.0 -replace=github.com/milwad-dev/url-shortener/contracts=../proto
