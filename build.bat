rem set GOOS=js
rem set GOARCH=wasm
rem go build -o main.wasm main.go
set GOOS=linux
gopherjs build -o main.js main.go