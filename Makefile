install:
	go get -u github.com/golang/lint/golint
	cd ../../.. && go install github.com/foospidy/sigsci-sounds && cd github.com/foospidy/sigsci-sounds

lint:
	clear
	golint sigsci-sounds.go

run:
	go run sigsci-sounds.go  
