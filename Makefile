build:
	go build -o simple-gpg .

build_linux:
	GOOS=linux go build -o simple-gpg .

build_darwin:
	GOOS=darwin go build -o simple-gpg .

install:
	go install .
