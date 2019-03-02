#!/bin/bash

# Pulls any changes that have been made to dev branch
git pull origin dev

# Push changes to remote repo
git push

# Prompts for user input when done 
# (this is so that the console stays open so that user can read any error messages)
read -p "Press Return to Close..."
