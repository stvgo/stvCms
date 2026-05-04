---
allowed-tools: Bash(go test:*), Bash(go tool cover:*)
description: Ejecuta los tests y verifica que el coverage de services + repository sea >= 90%, igual que el CI de GitHub Actions.
---

## CI Coverage — Reglas del proyecto

El CI (`.github/workflows/ci.yml`) mide cobertura **solo** en estas capas:

| Paquete | Umbral |
|---------|--------|
| `internal/services/...` | ≥ 90% |
| `internal/repository/...` | ≥ 90% |
| **Combinado** | ≥ 90% |

Capas **excluidas** del gate de coverage: `internal/handlers/`, `internal/clients/`, `internal/config/`, `internal/mocks/`.

## Coverage actual

!`go test ./internal/services/... ./internal/repository/... -coverprofile=/tmp/stvcms_coverage.out -covermode=atomic 2>&1`

!`go tool cover -func=/tmp/stvcms_coverage.out 2>/dev/null | grep -E "(services/Post|repository/Post|total)"`

## Tu tarea

1. Analiza el output de arriba.
2. Si el coverage combinado está **por debajo del 90%**, identifica las funciones con cobertura baja y escribe los tests faltantes.
3. Si está **en 90% o más**, confirma que el CI pasará y no hagas cambios.
4. Recuerda: los handlers **no** se testean en este proyecto.
