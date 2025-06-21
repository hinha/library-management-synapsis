#!/bin/bash

set -euo pipefail

SRC_DIR="gen/gen"
DST_DIR="gen"

if [ -d "$SRC_DIR" ]; then
  echo "Copying contents from '$SRC_DIR' to '$DST_DIR'..."

  shopt -s dotglob

  for item in "$SRC_DIR"/*; do
    if [ -d "$item" ]; then
      base=$(basename "$item")
      mkdir -p "$DST_DIR/$base"
      cp -R "$item"/* "$DST_DIR/$base/"
    else
      cp -R "$item" "$DST_DIR/"
    fi
  done

  shopt -u dotglob

  echo "Deleting source directory '$SRC_DIR'..."
  rm -rf "$SRC_DIR"

  echo "Folder structure has been fixed."
else
  echo "No nested '$SRC_DIR' found — structure is already clean."
fi