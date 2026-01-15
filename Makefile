g:
	git pull
	git add .
	git commit -m "Docker Dev Stack + DB Connectivity"
	git push

up:
	docker compose up -d
	docker logs -f planet_api

ps:
	docker compose ps