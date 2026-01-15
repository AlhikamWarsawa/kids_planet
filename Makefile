g:
	git pull
	git add .
	git commit -m "update architecture docs"
	git push

up:
	docker compose up -d
	docker logs -f planet_api

ps:
	docker compose ps