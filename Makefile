g:
	git pull
	git add .
	git commit -m "Catalog UI for Player + Filters"
	git push

dev:
	docker compose up -d
	docker logs -f planet_api

ps:
	docker compose ps

prod:
	docker compose -f infra/docker-compose.yml up -d --build
	docker compose logs

d:
	docker compose -f infra/docker-compose.yml down

logs-inf:
	docker compose -f infra/docker-compose.yml logs -f --tail=200

psql:
	docker exec -it planet_postgres psql -U admin -d kids_planet

psqll:
	docker exec -it planet_postgres_infra psql -U admin -d kids_planet