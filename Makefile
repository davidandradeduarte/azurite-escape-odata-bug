azurite:
	docker run --name azurite -d --restart unless-stopped -p 10000:10000 -p 10001:10001 -p 10002:10002 mcr.microsoft.com/azure-storage/azurite:3.24.0

run:
	go run main.go