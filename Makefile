g:
	git pull
	git add .
	git commit -m "Update Database Schemes and Index"
	git push

up:
	docker compose up -d
	docker logs -f planet_api

ps:
	docker compose ps

psql:
	docker exec -it planet_postgres psql -U admin -d kids_planet