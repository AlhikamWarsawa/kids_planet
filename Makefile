g:
	git pull
	git add .
	git commit -m "Full Stack Infra + Nginx Routing"
	git push

up:
	docker compose up -d
	docker logs -f planet_api

ps:
	docker compose ps

inf:
	docker compose -f infra/docker-compose.yml up -d --build

d:
	docker compose -f infra/docker-compose.yml down

logs-inf:
	docker compose -f infra/docker-compose.yml logs -f --tail=200

psql:
	docker exec -it planet_postgres psql -U admin -d kids_planet