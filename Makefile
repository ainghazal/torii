run:
	go build && DEBUG=1 ./torii

dev-deps:
	go install github.com/githubnemo/CompileDaemon

dev-server:
	CompileDaemon -command="./torii"
