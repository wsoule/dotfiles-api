#!/bin/bash
cd /Users/wyatsoule/Sites/Dotfiles_Go/Api
echo "Cleaning Go cache..."
go clean -cache
echo "Building..."
go build -o /tmp/api-test main.go
echo "Build result: $?"
