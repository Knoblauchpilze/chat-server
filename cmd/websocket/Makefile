
APPLICATION ?= websocket

setup:
	mkdir -p build/bin
	cp configs/config-template-dev.yml configs/config-dev.yml

release:
	go build -o build/bin/${APPLICATION} main.go

run: release
	./build/bin/${APPLICATION} config-dev

clean:
	rm -rf build configs/config-dev.yml
