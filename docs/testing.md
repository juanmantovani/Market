# Pruebas y cobertura

## Herramientas

**Backend (Go)**: la librería estándar. `testing` para unidad e integración,
`net/http/httptest` para handlers y middlewares a través de requests reales,
`go test -coverprofile` + `go tool cover` para cobertura. Sin librerías de
aserciones.

- Los casos de uso se prueban con fakes de los contratos de `gateway`,
  ubicados en `internal/business/gateway/gatewayfake`.
- Los repositorios en memoria se prueban contra la suite de contrato
  `internal/infrastructure/repository/repositorytest`, que después corre igual
  contra PostgreSQL. La primera es `CategoryRepository`.

**Frontend**: Vitest + jsdom, React Testing Library,
`@testing-library/jest-dom` y `@testing-library/user-event`. Cobertura V8.
Las pruebas consultan la UI como una persona: roles, textos, labels, estados
de carga, vacío y error. Nada de estado interno ni clases de Tailwind.

- `renderWithProviders` (`src/test/render.tsx`) monta una pantalla con
  QueryClient (sin reintentos) y toasts.
- Las pantallas de ABM se prueban contra una API falsa en memoria que
  reemplaza `fetch` y respeta las reglas del backend (unicidad,
  eliminación, orden, paginado): los tests recorren flujos completos y verifican
  qué requests se mandaron.
- Como el listado se dibuja como tabla y como tarjetas a la vez (jsdom no
  aplica los breakpoints), las consultas se acotan con `within(table)` o
  `within(list)`.
- Los tests de página tipean en formularios reales y, con cobertura y todos
  los workers ocupados, tardan varios segundos: el límite por test es de
  15 s (`testTimeout`) y los `findBy…` esperan hasta 3 s
  (`asyncUtilTimeout`). Si un test se acerca a ese límite, conviene
  partirlo antes que subirlo.

## Umbrales

| Métrica | Mínimo |
| --- | ---: |
| Líneas / sentencias | 80 % |
| Funciones | 80 % |
| Ramas | 75 % |
| Reglas críticas (venta, stock, pagos) | 90 % |

- **Frontend**: Vitest aplica los umbrales por archivo sobre
  `src/features/**`, `src/shared/api/**` y `src/shared/lib/**`. El reporte de
  texto oculta los archivos al 100 %; `coverage/index.html` los muestra todos.
- **Backend**: `scripts/check-go-coverage.mjs` lee
  `apps/backend/coverage.out` y falla si un paquete o el total queda debajo
  del 80 % de sentencias. Los paquetes medidos se listan en el script
  `test:backend:coverage:profile` de `package.json`: al agregar una feature,
  sumar sus paquetes de `usecase` y `dto` ahí.
- Quedan afuera de la medición `cmd/server`, `internal/app` (composition
  root) y los paquetes sin lógica (`gateway`, `dto`). `internal/app` igual
  tiene tests de flujo completo que levantan el contenedor real con memoria y
  van por HTTP: stock sobre los datos demo, y venta (alta de artículo →
  ingreso → venta → reintento idempotente → stock insuficiente → clientes y
  vendedores con ventas).

## Comandos

```powershell
pnpm test                    # backend + frontend
pnpm test:coverage           # ambos, aplicando umbrales
pnpm test:backend            # go test ./...
pnpm test:frontend           # vitest run
pnpm --filter @market/frontend test:watch
```
