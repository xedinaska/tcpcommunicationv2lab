fmt:
	gofmt -s -w .

build:
	go build -o bin/app cmd/main.go

build_win:
	GOOS=windows go build -o bin/app-win cmd/main.go

build_linux:
	GOOS=linux go build -o bin/app-linux cmd/main.go

run: build
	./bin/app

run_win: build_win
	./bin/app-win

run_linux: build_linux
	./bin/app-linux

