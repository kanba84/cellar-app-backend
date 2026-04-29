#!/bin/bash

# 1. Build the application for Linux ARM
echo "(1/5) Building the application for Linux ARM..."
GOOS=linux GOARCH=arm GOARM=6 go build -o cellar-app

# 2. Stop the running service
echo "(2/5) Stopping the running service..."
ssh cellar-app "sudo systemctl stop cellar-app.service"

# 3. Copy the new binary to the server
echo "(3/5) Copying the new binary to the server..."
scp cellar-app cellar-app:~

# 4. Start the updated service
echo "(4/5) Starting the updated service..."
ssh cellar-app "sudo systemctl restart cellar-app.service"

#5. Check the status of the service
echo "(5/5) Checking the status of the service..."
ssh cellar-app "sudo systemctl status cellar-app.service"