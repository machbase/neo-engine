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

@REM go build -o tmp/mach.exe -tags=fog_edition .\windows\main.go

go test -tags=fog_edition -timeout 30s -run ^TestAppendLog$  github.com/machbase/neo-engine/test