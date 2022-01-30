


all: build deploy

build:
	GOOS=linux GOARCH=arm GOARM=7 go build .

deploy:
	scp t80nxbt pi@procon.local:/home/pi/
