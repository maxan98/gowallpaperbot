build-client:
	cd client && go build .
build-server:
	cd server && go build .
gofmt:
	gofmt -w ./server/ && gofmt -w ./client/