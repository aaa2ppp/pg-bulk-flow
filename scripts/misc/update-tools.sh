#!/bin/sh

# Основные инструменты от Go team
go install golang.org/x/tools/cmd/...@latest

# Популярные сторонние утилиты
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

go install github.com/swaggo/swag/cmd/swag@latest
