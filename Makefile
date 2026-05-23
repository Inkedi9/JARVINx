APP=jarvinx
CMD=cmd/main.go

.PHONY: run build build-linux build-windows clean help

help:
	@echo ""
	@echo "  JARVINx — Autonomous Agent Runtime"
	@echo ""
	@echo "  make run            Lancer en mode dev (go run)"
	@echo "  make build          Compiler pour l'OS actuel"
	@echo "  make build-linux    Cross-compiler pour Linux amd64"
	@echo "  make build-windows  Cross-compiler pour Windows amd64"
	@echo "  make clean          Supprimer les binaires"
	@echo ""

run:
	go run $(CMD)

build:
	go build -ldflags="-s -w" -o $(APP) $(CMD)
	@echo "Binaire : ./$(APP)"

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(APP)-linux $(CMD)
	@echo "Binaire : ./$(APP)-linux"

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(APP).exe $(CMD)
	@echo "Binaire : ./$(APP).exe"

clean:
	rm -f $(APP) $(APP)-linux $(APP).exe
	@echo "Binaires supprimés"