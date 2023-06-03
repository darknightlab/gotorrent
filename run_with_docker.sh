#!/bin/bash

mkdir -p config
mkdir -p caddy
wget https://raw.githubusercontent.com/darknightlab/gotorrent/main/docker-compose.yml
wget https://raw.githubusercontent.com/darknightlab/gotorrent/main/config/config.example.yaml -O config/config.yaml
wget https://raw.githubusercontent.com/darknightlab/gotorrent/main/caddy/Caddyfile.example -O caddy/Caddyfile

read -p "Please enter your domain name: " domain
sed -i "s/example.com/$domain/g" caddy/Caddyfile
read -p "Please enter your email: " email
sed -i "s/<your email>/$email/g" caddy/Caddyfile
read -p "Please enter your cloudflare api token: " token
sed -i "s/<your api token>/$token/g" caddy/Caddyfile

docker-compose up -d
