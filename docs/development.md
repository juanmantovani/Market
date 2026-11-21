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
| `db` | localhost:5432 | PostgreSQL. Arranca siempre, aunque `STORAGE_DRIVER` sea `memory`. |
| `migrate` | — | Aplica las migraciones de `migrations/` con Goose y termina; `backend` espera a que termine antes de arrancar. |

`BACKEND_PORT`, `FRONTEND_PORT` y `DB_PORT` cambian los puertos publicados.

Por defecto `STORAGE_DRIVER` sigue siendo `memory`: `db` y `migrate` quedan
listos pero el backend no los usa. Para probar contra PostgreSQL:

```powershell
STORAGE_DRIVER=postgres SEED_DEMO_DATA=false docker compose up -d backend
```

`SEED_DEMO_DATA=false` es necesario hoy: los datos demo todavía dan por
sentado que **todos** los repositorios son PostgreSQL (hasta ahora solo
`CustomerRepository` lo es), y entrar con `memory` para el resto haría panic.
Por la misma razón, con `STORAGE_DRIVER=postgres` solo responden
`/health/ready` y las rutas de `/api/v1/customers`; el resto da `500` hasta
que su adaptador llegue.

## Migraciones (Goose)

Las migraciones SQL viven en `migrations/`, versionadas con
[Goose](https://github.com/pressly/goose) (`-- +goose Up` / `-- +goose
Down`). `apps/backend/cmd/migrate` es un binario propio que las aplica: no
hace falta instalar la CLI de Goose.

```powershell
cd apps/backend
DATABASE_URL="postgres://market:market@localhost:5432/market?sslmode=disable" go run ./cmd/migrate up
```

El primer argumento es el comando de Goose (`up`, `status`, `down`, …); por
defecto es `up`. Contra el `db` de Compose, el puerto es el de `DB_PORT`
(`5432` por defecto). El servicio `migrate` de Compose corre exactamente este
binario.

## Variables de entorno

Ninguna es secreta todavía. `.env.example` las lista con sus valores por
defecto.

| Variable | Dónde | Default | Uso |
| --- | --- | --- | --- |
| `PORT` | backend | `8080` | Puerto HTTP de la API. |
| `STORAGE_DRIVER` | backend | `memory` | Persistencia: `memory` (se pierde al reiniciar) o `postgres`. |
| `DATABASE_URL` | backend, `migrate` | — | Cadena de conexión a PostgreSQL. Obligatoria si `STORAGE_DRIVER=postgres`, o para correr `cmd/migrate`. |
| `CORS_ALLOWED_ORIGIN` | backend | vacío | Solo si el frontend se sirve desde otro origen que la API. |
| `SEED_DEMO_DATA` | backend | `true` | Carga datos de ejemplo al iniciar (catálogo, stock, terceros y algunas ventas). `false` arranca vacío, salvo el cliente Consumidor Final. |
| `STOCK_ALLOW_NEGATIVE` | backend | `false` | `true` deja vender más de lo que hay (el stock queda negativo). Con `false`, la venta falla con `insufficient_stock`. |
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

`pnpm test` corre igual sin PostgreSQL: la suite de contrato de
`repository/postgres` se salta (`t.Skip`) si no hay `TEST_DATABASE_URL`.
`pnpm test:coverage`, en cambio, sí la necesita para llegar al umbral de
cobertura de ese paquete (CI siempre la tiene). Para tenerla en el día a día:

```powershell
docker compose up -d db migrate
$env:TEST_DATABASE_URL = "postgres://market:market@localhost:5432/market?sslmode=disable"
pnpm test:coverage
```
