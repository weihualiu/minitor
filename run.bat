:: go program run bat

echo off
set current_path="%cd%"
:: win7 use
setx GOPATH %current_path%
echo on

go install minitor check ftp config

cd bin
minitor.exe ../config/minitor.conf

pause
