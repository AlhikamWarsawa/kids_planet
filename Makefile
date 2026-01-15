g:
	git pull
	git add .
	git commit -m "changelog"
	git push

up:
	docker compose up -d
	docker logs -f planet_api

ps:
	docker compose ps