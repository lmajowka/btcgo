windows-up:
	set $(cat .env | xargs) 
	rm -rf go.*
	go mod init btcgo
	go mod tidy
	go build -o btcgo.exe ./src
	./btcgo

linux-up:
	export $(cat .env | xargs) 
	rm -rf go.*
	go mod init btcgo
	go mod tidy
	go build -o btcgo ./src
	./btcgo
	
docker-up:
	docker-compose -f docker/docker-compose.yml up -d
	docker exec -it btcgo ./btcgo

docker-down:
	docker rm btcgo --force