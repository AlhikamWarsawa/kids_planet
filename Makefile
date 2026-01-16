g:
	git pull
	git add .
	git commit -m "DB migrations V1"
	git push

up:
	docker compose up -d
	docker logs -f planet_api

ps:
	docker compose ps

psql:
	docker exec -it planet_postgres psql -U admin -d kids_planet