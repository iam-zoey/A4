#!/bin/bash

# Start Trader 1
go run trader.go -id=1 -address=localhost:8001 -peer=localhost:8002 -post=1 > log/trader1.txt 2>&1 &
echo "Trader 1 started at localhost:8001"

# Start Trader 2
go run trader.go -id=2 -address=localhost:8002 -peer=localhost:8001 -post=2 > log/trader2.txt 2>&1 &
echo "Trader 2 started at localhost:8002"

# Start Seller 1
go run seller/seller.go -id=1 -address=localhost:8003 -trader=localhost:8001 -post=1 > log/seller1.txt 2>&1 &
echo "Seller 1 started at localhost:8003"

# Start Seller 2
go run seller/seller.go -id=2 -address=localhost:8004 -trader=localhost:8002 -post=2 > log/seller2.txt 2>&1 &
echo "Seller 2 started at localhost:8004"