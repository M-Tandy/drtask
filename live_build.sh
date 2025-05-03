#!/bin/sh

# in foreground, continously run app, looking for changes
watchexec -r -e go --wrap-process session -- "go run *.go"
