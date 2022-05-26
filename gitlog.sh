#!/bin/bash

git --no-pager log origin/main..HEAD --pretty=oneline  --abbrev-commit
