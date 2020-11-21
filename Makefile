IMG := registry.cn-huhehaote.aliyuncs.com/feng-566/tax-crawler:v0.0.1

build:
	docker build -t $(IMG) .

push:
	docker push $(IMG)

image: build push

run:
	go run ./

docker-run:
	docker run -d --name tax --restart on-failure $(IMG) --verbose info --cron 30 --range 21600
