#!/bin/bash

sudo fuser -k 8003/tcp
sudo fuser -k 8001/tcp
sudo fuser -k 8002/tcp
sudo fuser -k 8004/tcp