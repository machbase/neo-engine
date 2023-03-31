@REM 1. Install cygwin with gcc toolchain
@REM      
@REM 2. Install lib2a tool (convert .lib to .a) 
@REM      https://code.google.com/archive/p/lib2a/downloads
@REM    lib2a converting .lib (w/.dll) to .a
@REM

@SET GO11MODULE=on
@SET CGO_ENABLED=1
@SET GOOS=windows
@SET GOARCH=amd64
@REM SET CC=zig cc -target native-native-msvc -library c
@REM SET CXX=zig c++ -target native-native-msvc
@REM SET CXX=
@REM @SET CGO_CFLAGS=-IC:/Users/Eirny/zig-windows-x86_64-0.11.0-dev.2160+49d37e2d1/lib/libc/include/any-windows-any
@REM SET CGO_LDFLAGS=-static ./native/libmachengine_fog_windows_amd64.lib 
SET CGO_LDFLAGS= --static -l.\native\libmachengine_fog_windows_amd64.lib
SET CGO_CFLAGS=

go build -v -buildmode=c-shared -o tmp\a.out -tags=fog_edition .\windows\main.go 
