APP=jarvinx
CMD=cmd/main.go
VERSION = 1.10.0

LDFLAGS = -ldflags "-s -w -X main.Version=$(VERSION)"


.PHONY: run build build-linux build-windows clean help

help:
	@echo ""
	@echo "  JARVINx — Autonomous Agent Runtime v$(VERSION)"
	@echo ""
	@echo "  make run            Lancer en mode dev"
	@echo "  make build          Compiler pour l'OS actuel"
	@echo "  make build-linux    Cross-compiler pour Linux amd64"
	@echo "  make build-windows  Cross-compiler pour Windows amd64"
	@echo "  make clean          Supprimer les binaires"
	@echo ""

run:
	go run $(CMD)

build:
	go build $(LDFLAGS) -o $(APP) $(CMD)
	@echo "Binaire : ./$(APP) (v$(VERSION))"

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(APP)-linux $(CMD)
	@echo "Binaire : ./$(APP)-linux (v$(VERSION))"

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(APP).exe $(CMD)
	@echo "Binaire : ./$(APP).exe (v$(VERSION))"

clean:
	rm -f $(APP) $(APP)-linux $(APP).exe
	@echo "Binaires supprimés"