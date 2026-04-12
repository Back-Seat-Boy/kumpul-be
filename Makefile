.PHONY: build run test clean docker-build docker-up docker-down migrate-up migrate-down deploy

# Development commands
build:
	go build -o bin/kumpul main.go

run:
	go run main.go server

dev:
	docker-compose up -d
	go run main.go server

test:
	go test -v ./...

clean:
	rm -rf bin/
	go clean

# Docker commands
docker-build:
	docker-compose -f docker-compose.prod.yml build

docker-up:
	docker-compose -f docker-compose.prod.yml up -d

docker-down:
	docker-compose -f docker-compose.prod.yml down

docker-logs:
	docker-compose -f docker-compose.prod.yml logs -f

# Migration commands
migrate-up:
	go run main.go migrate up

migrate-down:
	go run main.go migrate down

migrate-status:
	go run main.go migrate status

# Production deployment
deploy:
	./deploy.sh

# Production helpers
prod-up:
	docker-compose -f docker-compose.prod.yml up -d

prod-down:
	docker-compose -f docker-compose.prod.yml down

prod-logs:
	docker-compose -f docker-compose.prod.yml logs -f

prod-migrate:
	docker-compose -f docker-compose.prod.yml exec app ./kumpul migrate up

prod-backup:
	docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U postgres kumpul > backup_$$(date +%Y%m%d_%H%M%S).sql

# SSL certificate
ssl-cert:
	docker-compose -f docker-compose.prod.yml run --rm certbot certonly --webroot --webroot-path=/var/www/certbot -d $$(DOMAIN) --agree-tos --email $$(EMAIL) --no-eff-email

ssl-renew:
	docker-compose -f docker-compose.prod.yml run --rm certbot renew
