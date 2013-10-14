#!/bin/sh

if [ -z "${GOARCH+x}" ]; then
	echo "GOARCH environment variable is not set!"

	exit 1
fi

if [ -z "${GOOS+x}" ]; then
	echo "GOOS environment variable is not set!"

	exit 1
fi

if [ ! -f $GOROOT/src/all.bash ]; then
    echo "GOROOT must contain a compileable Go installation because auf cross-compiling"

	exit 1
fi

# compile go

TMPOLD="$(pwd)"

cd $GOROOT/src

./make.bash

cd $TMPOLD

# compile requirements

go clean github.com/lib/pq
go install github.com/lib/pq
go clean github.com/robfig/revel
go install github.com/robfig/revel
go clean github.com/robfig/revel/revel
go install github.com/robfig/revel/revel

# compile tirion

make -C $GOPATH/src/github.com/zimmski/tirion clean
make -C $GOPATH/src/github.com/zimmski/tirion libs
make -C $GOPATH/src/github.com/zimmski/tirion tirion-agent

# init

TMPFOLDER="$(mktemp --directory)"

echo "Build package in folder $TMPFOLDER"

# lib

mkdir $TMPFOLDER/lib

mkdir $TMPFOLDER/lib/c
cp $GOPATH/src/github.com/zimmski/tirion/clients/c-client/libtirion.a $TMPFOLDER/lib/c/libtirion.a
cp $GOPATH/src/github.com/zimmski/tirion/clients/c-client/tirion.h $TMPFOLDER/lib/c/tirion.h

mkdir -p $TMPFOLDER/lib/go
mkdir -p $TMPFOLDER/lib/go/pkg/${GOOS}_${GOARCH}/github.com/zimmski
cp $GOPATH/pkg/${GOOS}_${GOARCH}/github.com/zimmski/tirion.a $TMPFOLDER/lib/go/pkg/${GOOS}_${GOARCH}/github.com/zimmski/tirion.a

# server

mkdir -p $TMPFOLDER/share
revel build github.com/zimmski/tirion/tirion-server $TMPFOLDER/share

rm $TMPFOLDER/share/run.bat
rm $TMPFOLDER/share/run.sh

mv $TMPFOLDER/share/src/github.com $TMPFOLDER/share/github.com
rmdir $TMPFOLDER/share/src

rm -r $TMPFOLDER/share/github.com/robfig/revel/modules
rm -r $TMPFOLDER/share/github.com/zimmski/tirion/tirion-server/app/controllers
rm -r $TMPFOLDER/share/github.com/zimmski/tirion/tirion-server/app/routes
rm -r $TMPFOLDER/share/github.com/zimmski/tirion/tirion-server/app/tmp
rm -r $TMPFOLDER/share/github.com/zimmski/tirion/tirion-server/app/init.go
rm -r $TMPFOLDER/share/github.com/zimmski/tirion/tirion-server/conf/app.conf
rm -r $TMPFOLDER/share/github.com/zimmski/tirion/tirion-server/tests
rm -r $TMPFOLDER/share/github.com/zimmski/tirion/tirion-server/README.md

mv $TMPFOLDER/share/github.com/zimmski/tirion/tirion-server/scripts $TMPFOLDER/scripts

# bin

mkdir $TMPFOLDER/bin

cp $GOPATH/bin/tirion-agent $TMPFOLDER/bin/tirion-agent
mv $TMPFOLDER/share/tirion-server $TMPFOLDER/bin/tirion-server

chmod +x $TMPFOLDER/bin/*

# zip
TMPOLD="$(pwd)"
cd $TMPFOLDER
ZIPNAME=$TMPOLD/tirion-$GOOS-$GOARCH.zip
rm $ZIPNAME
zip -9 -r $ZIPNAME *
cd $TMPOLD

# cleanup

# TODO uncomment
#rm -r $TMPFOLDER
