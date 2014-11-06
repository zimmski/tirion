#!/bin/sh

for FOLDER in golang.org/x/net/websocket github.com/howeyc/fsnotify github.com/lib/pq github.com/robfig/config github.com/streadway/simpleuuid github.com/robfig/pathtree github.com/robfig/revel github.com/robfig/revel/revel; do
	echo "Process $FOLDER"
	rm -rf $GOPATH/pkg/*/$FOLDER*
	if [ -d "$GOPATH/src/$FOLDER" ]; then
		go clean $FOLDER
		go install $FOLDER
	else
		go get $FOLDER
	fi
done

