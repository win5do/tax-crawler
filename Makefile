IMG := registry.cn-huhehaote.aliyuncs.com/feng-566/tax-crawler:v0.0.1

build:
	docker build -t $(IMG) .

push:
	docker push $(IMG)

image: build push

ARGS := --verbose debug --cron 3 --range 21600

run: ARGS := $(ARGS) --db tax.db
run:
	go run . $(ARGS)

docker-run:
	docker volume create tax; \
	docker run -d --name tax --restart on-failure -v tax:/opt/data $(IMG) $(ARGS)
