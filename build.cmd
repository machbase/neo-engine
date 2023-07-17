@REM Install cygwin with gcc toolchain
@REM      
@REM    - Prefer using TDM-GCC-64
@REM

@SET GOOS=windows
@SET GOARCH=amd64
@SET CGO_ENABLED=1
@SET CC=C:\TDM-GCC-64\bin\gcc.exe
@SET CXX=C:\TDM-GCC-64\bin\g++.exe
@SET CGO_LDFLAGS=
@SET CGO_CFLAGS=
@SET GO11MODULE=on

@REM go build -o tmp/mach.exe .\windows\main.go

@go test -v -count=1 -timeout 30s  github.com/machbase/neo-engine/test