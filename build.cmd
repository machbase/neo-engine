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
@REM SET CGO_LDFLAGS=-static ./native/libmachengine_fog_windows_amd64.a
SET CGO_LDFLAGS= --static -lmachengine_fog_windows_amd64 -LC:/Users/ratse/work/neo-engine/native  -lm  -lws2_32  -lnetapi32 -ladvapi32 -liphlpapi -ldbghelp -lshell32 -luser32
SET CGO_CFLAGS=

go build -buildmode=c-shared -o tmp\mach.exe -tags=fog_edition .\windows\main.go
