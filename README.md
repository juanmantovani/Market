# Market

Punto de venta: monorepo con una aplicación React/Vite y una API Go. La
persistencia arranca en memoria y pasa a PostgreSQL en una fase posterior.

## Inicio rápido

Requisitos: Go 1.27+, Node.js 20.19+ y pnpm 10+ (o Docker Desktop).

```powershell
pnpm install
pnpm dev:backend   # API en http://localhost:8080
pnpm dev           # en otra terminal: app en http://localhost:5173
```

O todo junto con Docker:

```powershell
docker compose up --build
```

Los datos viven en memoria: se pierden al reiniciar el backend.

## Comandos frecuentes

```powershell
pnpm test            # tests de backend y frontend
pnpm test:coverage   # tests con umbrales de cobertura
pnpm typecheck
pnpm lint
pnpm build
```

## Estructura

```text
apps/
├── backend/       # API Go (dominio, casos de uso, adaptadores)
└── frontend/      # React + TypeScript + Vite + Tailwind CSS + shadcn/ui
docs/              # Documentación técnica y planes
scripts/           # Utilidades de CI (umbral de cobertura de Go)
compose.yaml       # Entorno completo de desarrollo
```

## Documentación

- [Índice de documentación](docs/README.md)
- [Plan de la versión base](docs/plans/version-base.md)
- [Desarrollo local](docs/development.md)
- [Arquitectura](docs/architecture.md)
- [Dominio](docs/domain.md)
- [Pruebas y cobertura](docs/testing.md)
- [Integración continua](docs/ci.md)
- [Reglas para agentes y colaboradores](AGENTS.md)
