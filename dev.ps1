# dev.ps1
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd runtime; go run ./cmd/main.go"
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd dashboard; npm run dev"