#!/bin/sh
cdFTC='/home/itsc/gopath/src/FTC/'

echo "进入"$cdFTC
cd $cdFTC
cd devservice/

echo "启动etcClient"
nohup ./etcClient &

echo "进入./ep/"
cd ./ep/
echo "启动epService"
nohup ./epService &


echo "进入"$cdFTC
cd $cdFTC
echo "启动main"
nohup ./main &

echo "============start end=================================="








 
