#!/bin/bash

set -x

# Private project dependencies
private_deps=( )
support_deps=( "launchpad.net/gocheck" "github.com/scottferg/go2xunit" )

BINARYNAME=vip

ROOTDIR=$(pwd)

LOCAL_GOPATH=$ROOTDIR/gopath
GOROOT=$HOME/go

rm -rf $LOCAL_GOPATH

export GOPATH=$LOCAL_GOPATH
export PATH=$GOROOT/bin:$LOCAL_GOPATH/bin:$GOPATH/bin:/usr/local/bin:$PATH

IMPORTROOT=github.com/vokalinteractive

IMPORTPATH=$IMPORTROOT/$BINARYNAME

mkdir -p $LOCAL_GOPATH/src/$IMPORTPATH

rm -rf $LOCAL_GOPATH/src/$IMPORTPATH 
mkdir -p $LOCAL_GOPATH/src/$IMPORTPATH 

cp *.go $LOCAL_GOPATH/src/$IMPORTPATH

for dir in */
do 
    if [ $dir != "gopath/" ]; then
        mkdir -p $LOCAL_GOPATH/src/$IMPORTPATH
        cp -r $dir $LOCAL_GOPATH/src/$IMPORTPATH
    fi
done

# Clone necessary dependencies
for dep in "${private_deps[@]}"
do
    rm -rf $LOCAL_GOPATH/src/$IMPORTROOT/$dep
    git clone git@github.com:vokalinteractive/$dep.git $LOCAL_GOPATH/src/$IMPORTROOT/$dep
done

# Build any ancillary tools/libraries
for dep in "${support_deps[@]}"
do
    go get $dep
    go install $dep
done

TESTFILE=$ROOTDIR/tests.xml

COVERAGEOUT=$ROOTDIR/coverage.out
COVERAGEHTML=$ROOTDIR/coverage.html
COVERAGEJSON=$ROOTDIR/coverage.json
COVERAGEXML=$ROOTDIR/coverage.xml

cd $LOCAL_GOPATH/src/$IMPORTPATH

go get

go test -i
go test -gocheck.v -coverprofile=$COVERAGEOUT | $LOCAL_GOPATH/bin/go2xunit -fail -output $TESTFILE
if [ $? -ne 0 ]; then
    exit 1
fi

go tool cover -html=$COVERAGEOUT -o $COVERAGEHTML
if [ $? -ne 0 ]; then
    exit 2
fi

mv $LOCAL_GOPATH/bin/$BINARYNAME $ROOTDIR/$BINARYNAME
