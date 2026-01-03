#!/bin/bash
set -e

echo "📦 Building and Publishing Engram Python SDK..."

cd sdk/python

# Clean old builds
rm -rf dist/ build/ *.egg-info/

# Setup temp venv for build tools
python3 -m venv .venv_publish
source .venv_publish/bin/activate

# Install/Update build tools
pip install --upgrade pip
pip install --upgrade build twine

# Build package
python3 -m build

# Upload to PyPI
# Twine will use TWINE_USERNAME and TWINE_PASSWORD env vars if set
python3 -m twine upload --non-interactive dist/*

# Cleanup
deactivate
rm -rf .venv_publish

echo "✅ Published engram-sdk to PyPI!"
