IMG := registry.cn-huhehaote.aliyuncs.com/feng-566/tax-crawler:v0.0.1

build:
	docker build -t $(IMG) .

push:
	docker push $(IMG)

image: build push

run:
	go run ./

docker-run:
	docker run -d --restart on-failure $(IMG) --verbose debug --cron 3 --range 21600
