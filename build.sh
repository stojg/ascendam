#!/bin/sh -ex

BUILDDIR=./build
LATEST_TAG=`git describe --abbrev=0 --tags`

rm -rf ${BUILDDIR} && mkdir -p ${BUILDDIR}
gox -output="$BUILDDIR/{{.Dir}}_${LATEST_TAG}_{{.OS}}_{{.Arch}}" -os "linux darwin windows"

for file in ${BUILDDIR}/*
do
	zip "$file.zip" "$file"
	rm "$file"
done

