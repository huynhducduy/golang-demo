docker-compose up development

docker-compose up -d production

docker-compose logs --tail=100 -f production

docker-compose exec production sh
