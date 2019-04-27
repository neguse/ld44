go test
if %ERRORLEVEL% neq 0 (
    exit /b 1
)
statik -f -src asset
go run main.go