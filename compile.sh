#!/bin/sh
cdFTC='/home/itsc/gopath/src/FTC/'
echo "============begin=================================="

echo "进入"$cdFTC
cd $cdFTC
cd devservice/

#echo "删除日志"
rm -rf logs etcClient
echo "编译etcClient"
go build etcClient.go 

echo "进入./ep/"
cd ./ep/
#echo "删除日志"
rm -rf logs epService
echo "编译epService"
go build epService.go blsdkbriage.go 

cd ../winep/
#echo "删除日志"
rm -rf logs winep.exe
echo "编译winep"
env CGO_ENABLED=1 GOOS=windows GOARCH=386 CC=i686-w64-mingw32-gcc go build -o winep.exe epService.go  blsdkbriage.go 

echo "进入"$cdFTC
cd $cdFTC
#echo "删除日志"
rm -rf logs main
echo "编译main"
go build main.go

echo "============end=================================="








 
