IMG=registry.cn-huhehaote.aliyuncs.com/feng-566/tax-crawler:v0.0.1

docker pull $IMG

docker rm -f tax

docker volume create tax; \
docker run -d --name tax --restart on-failure -v tax:/opt/data $IMG