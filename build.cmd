@REM 1. Install zig tool chain
@REM      https://ziglang.org/download/
@REM    Download zig tools and set PATH
@REM 2. Install lib2a tool (convert .lib to .a) 
@REM      https://code.google.com/archive/p/lib2a/downloads
@REM    lib2a converting .lib (w/.dll) to .a
@REM

@SET GO11MODULE=on
@SET CGO_ENABLED=1
@SET GOOS=windows
@SET GOARCH=amd64
@SET CC=zig cc -target x86_64-windows.win10_fe...win10_fe-gnu
@SET CX=zig c++ -target x86_64-windows.win10_fe...win10_fe-gnu

go test -v -count 1 -tags=fog_edition ./test