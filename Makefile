install:
	go get -u golang.org/x/lint/golint
	go get github.com/faiface/beep
	#cd ../../.. && go install github.com/foospidy/sigsci-sounds && cd github.com/foospidy/sigsci-sounds

lint:
	clear
	golint sigsci-sounds.go

run:
	go get github.com/hajimehoshi/go-mp3
	go get github.com/hajimehoshi/oto
	go get github.com/pkg/errors
	go get github.com/faiface/beep
	go run sigsci-sounds.go  

test:
	go get github.com/hajimehoshi/go-mp3
	go get github.com/hajimehoshi/oto
	go get github.com/pkg/errors
	go get github.com/faiface/beep
	go run test-sounds.go 