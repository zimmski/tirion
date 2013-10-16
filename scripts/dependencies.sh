#!/bin/sh

for FOLDER in code.google.com/p/go.net/websocket github.com/howeyc/fsnotify github.com/lib/pq github.com/robfig/config github.com/robfig/pathtree github.com/robfig/revel github.com/robfig/revel/revel github.com/streadway/simpleuuid; do
	echo "Process $FOLDER"
	rm -rf $GOPATH/pkg/*/$FOLDER*
	if [ -d "$GOPATH/src/$FOLDER" ]; then
		go clean $FOLDER
		go install $FOLDER
	else
		go get $FOLDER
	fi
done

