set GOOS=js
set GOARCH=wasm
go build -o main.wasm main.go
gopherjs build -o main.js main.go