# Desarrollo local

## Requisitos

- Go 1.27+ (`apps/backend/go.mod` fija la versión exacta; con
  `GOTOOLCHAIN=auto`, el default, Go descarga solo el toolchain que falte).
- Node.js 20.19+ y pnpm 10+ (el repo fija `pnpm@11.21.0` en `packageManager`).
- Docker Desktop, solo para levantar todo con Compose.

> **Windows**: cloná el repo en una ruta corta (por ejemplo
> `C:\Users\<usuario>\Documents\Market`). Con rutas largas, las de
> `node_modules\.pnpm\…` superan el límite de 260 caracteres y Node deja de
> resolver módulos (ESLint falla con `ERR_PACKAGE_IMPORT_NOT_DEFINED`).

## Sin Docker (recomendado para el día a día)

En dos terminales, desde la raíz:

```powershell
pnpm install
pnpm dev:backend   # API en http://localhost:8080
```

```powershell
pnpm dev           # frontend en http://localhost:5173
```

El frontend llama a `/api` en su mismo origen; el proxy de Vite
los reenvía al backend (`API_PROXY_TARGET`, por defecto
`http://localhost:8080`). Por eso no hace falta CORS en desarrollo.

## Con Docker Compose

```powershell
docker compose up --build
```

| Servicio | URL | Notas |
| --- | --- | --- |
| `frontend` | http://localhost:5173 | Vite con HMR (polling para volúmenes de Docker). |
| `backend` | http://localhost:8080 | `go run ./cmd/server`; reiniciar el servicio para tomar cambios. |

`BACKEND_PORT` y `FRONTEND_PORT` cambian los puertos publicados.

## Variables de entorno

`.env.example` las lista con sus valores por defecto. La única secreta es
`GS1_PASSWORD`: no se commitea.

| Variable | Dónde | Default | Uso |
| --- | --- | --- | --- |
| `PORT` | backend | `8080` | Puerto HTTP de la API. |
| `STORAGE_DRIVER` | backend | `memory` | Persistencia. Hoy solo `memory`: los datos se pierden al reiniciar. |
| `CORS_ALLOWED_ORIGIN` | backend | vacío | Solo si el frontend se sirve desde otro origen que la API. |
| `SEED_DEMO_DATA` | backend | `true` | Carga datos de ejemplo al iniciar (catálogo, stock, terceros y algunas ventas). `false` arranca vacío, salvo el cliente Consumidor Final. |
| `STOCK_ALLOW_NEGATIVE` | backend | `false` | `true` deja vender más de lo que hay (el stock queda negativo). Con `false`, la venta falla con `insufficient_stock`. |
| `GS1_USERNAME` | backend | vacío | Usuario del portal de socios de GS1 Argentina. Junto con `GS1_PASSWORD`, sugiere el producto al escanear un código que todavía no está cargado. |
| `GS1_PASSWORD` | backend | vacío | Contraseña de ese portal. Secreta. Las dos variables van juntas o vacías. |

El proceso no lee `.env` solo: hay que exportar las variables antes de
`pnpm dev:backend` (Compose las toma del `.env` de la raíz). Para probar el
portal real, desde `apps/backend`:
`GS1_LIVE=1 GS1_USERNAME=… GS1_PASSWORD=… go test -run TestLivePortal ./internal/infrastructure/gs1argentina`.
| `API_PROXY_TARGET` | frontend (Vite) | `http://localhost:8080` | Destino del proxy de `/api`. |
| `VITE_API_URL` | frontend | vacío | Solo si la API vive en otro origen; se suma a `connect-src` de la CSP. |
| `WATCH_POLLING` | frontend (Vite) | vacío | `true` activa polling del watcher (lo usa Compose). |

## Verificación antes de un PR

```powershell
pnpm typecheck
pnpm lint
pnpm test
pnpm test:coverage
pnpm build
```

Y desde `apps/backend`: `go vet ./...` y `gofmt -l .` (sin salida).
