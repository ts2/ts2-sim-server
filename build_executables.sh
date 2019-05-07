#!/usr/bin/env bash

set -x

package="github.com/ts2/ts2-sim-server"
if [[ -z "$package" ]]; then
  echo "usage: $0 <package-name>"
  exit 1
fi
package_split=(${package//\// })
package_name=${package_split[-1]}


HERE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
BUILD_DIR="$HERE_DIR/build/downloads"

mkdir -p $BUILD_DIR

INDEX_FILE="$HERE_DIR/build/index.md"

cd $HERE_DIR/server 
rice embed-go
cd $HERE_DIR

echo "# Downloads" > $INDEX_FILE
echo "" >> $INDEX_FILE

DATE=`date '+%Y-%m-%d %H:%M:%S'`
echo "Built: $DATE" >> $INDEX_FILE 
echo "" >> $INDEX_FILE 

# ALL Wanted
#platforms=( "linux/386" "linux/amd64" "linux/arm" "windows/amd64" "windows/386" "darwin/amd64" "darwin/386" "darwin/arm64" )

## Working
platforms=(  "linux/arm" "windows/amd64" "windows/386" "linux/amd64" )

#platforms=( "linux/arm" "linux/386" "linux/amd64" )

echo "-----------------"

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    build_opts=""    
    output_name=$package_name'-'$GOOS'-'$GOARCH

    if [ $GOOS = "windows" ]; then
        ## Windows
        output_name+='.exe'
        if [ $GOARCH = "amd64" ]; then
            #build_opts="CXX_FOR_TARGET=i686-w64-mingw32-g++ CC_FOR_TARGET=i686-w64-mingw32-gcc"
            
            #build_opts="CC=x86_64-w64-mingw32-gcc"
        else 
            #build_opts="CC=i686-w64-mingw32-gcc"
        fi

    elif [ $GOOS = "linux" ]; then

        ## Linux ARM - sudo apt-get install libc6-armel-cross libc6-dev-armel-cross binutils-arm-linux-gnueabi libncurses5-dev
        if [ $GOARCH = "arm" ]; then
            #build_opts="CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ GOARM=7"
            #build_opts="CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ GOARM=7"
            output_name+='7'
       
        fi

    fi 
    #  -ldflags="-s -w"   CGO_ENABLED=1
    env GOOS=$GOOS GOARCH=$GOARCH $build_opts go build -o "$BUILD_DIR/$output_name" $package
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi

    echo " - <a href='/ts2-sim-server/downloads/$output_name'>$output_name </a>" >> $INDEX_FILE
done
