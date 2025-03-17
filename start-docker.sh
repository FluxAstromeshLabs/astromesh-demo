#!/bin/bash

rm -rf ~/.fluxd
cp -r ./chain/.fluxd ~/.fluxd
echo "killing running processes"
docker rm -f $(docker ps -aq)

echo "running fluxd container"
aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws/i1x2i1m1
docker run  --pull=always -d --volume=$HOME/.fluxd:/root/.fluxd -p 26657:26657 -p 26656:26656 -p 10337:10337 -p 9900:9900 --name fluxd public.ecr.aws/i1x2i1m1/fluxd:dev
