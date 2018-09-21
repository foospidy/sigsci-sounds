install:
	go get github.com/hajimehoshi/go-mp3
	go get github.com/hajimehoshi/oto
	go get github.com/pkg/errors
	go get github.com/faiface/beep
	go get -u golang.org/x/lint/golint
	go get github.com/faiface/beep
	go get github.com/signalsciences/go-sigsci

lint:
	find . -name '*.go' | xargs gofmt -w -s
	go vet

run:
	go run sigsci-sounds.go

test:
	go run sigsci-sounds.go test