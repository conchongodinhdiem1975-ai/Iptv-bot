#!/bin/bash
echo "--- Dang build bot ---"
go build -o iptv-bot main.go

echo "--- Dang chay bot (Scrape & Push) ---"
./iptv-bot

echo "--- Hoan tat! ---"
