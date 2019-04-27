set GOOS=windows
set GOARCH=amd64
go test
if %ERRORLEVEL% neq 0 (
    exit /b 1
)
statik -f -src asset
start go run main.go