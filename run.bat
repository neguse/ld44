set GOOS=windows
set GOARCH=amd64
go test
if %ERRORLEVEL% neq 0 (
    exit /b 1
)
start go run main.go