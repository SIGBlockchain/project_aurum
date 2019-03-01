#!/bin/bash

# Run all go test files in project directory
go test ../...

# Prompts for user input when done 
# (this is so that the console stays open so that user can read any error messages)
read -p "Press Return to Close..."
