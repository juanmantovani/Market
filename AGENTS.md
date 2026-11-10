# Market

Reglas para agentes y colaboradores. Este archivo es la fuente de verdad;
`CLAUDE.md` solo apunta acá. Adaptado de las reglas de LoteosAPP. El plan
vigente está en `docs/plans/version-base.md`.

## Flujo de trabajo

- Por defecto, el agente actual analiza, implementa y verifica el trabajo
  directamente, sin delegarlo a subagentes.
- Usar un flujo secuencial planner → programmer → reviewer solo cuando el
  usuario lo pida expresamente. En ese caso el agente principal coordina,
  conserva las decisiones del usuario, entrega la respuesta final y no
  implementa en paralelo; los subagentes no crean otros subagentes.

## Idioma

- Responder al usuario en español.
- Código, identificadores, comentarios y mensajes de commit en inglés.
- También en inglés todo lo que no ve el usuario: rutas de la API, campos
  JSON, nombres de tablas y columnas, códigos de error (`code`).
- En español: textos de la UI y el `Message` de los errores de negocio.
- Usar el mismo vocabulario para una entidad en dominio, API, base de datos y
  frontend. El glosario ES ↔ EN vive en `docs/domain.md`.

## Gestor de paquetes

- Nunca usar `npm`. Usar siempre `pnpm` (`pnpm install`, `pnpm add`,
  `pnpm remove`, `pnpm run`, `pnpm dlx`).
- No crear ni commitear `package-lock.json` ni `npm-shrinkwrap.json`; el
  lockfile del workspace es `pnpm-lock.yaml`.
- Si una documentación muestra un comando `npm`, traducirlo a `pnpm` antes de
  ejecutarlo.

## Estructura del repositorio

- `apps/frontend`: React + TypeScript + Vite + Tailwind CSS + shadcn/ui.
- `apps/backend`: API HTTP en Go.
- `migrations`: migraciones SQL versionadas de PostgreSQL con Goose (a partir
  de la fase de PostgreSQL).
- `docs`: documentación técnica y planes.
- Frontend y backend se construyen de forma independiente.

## Arquitectura del backend

- Organizar por capa técnica (`internal/business`, `internal/infrastructure`),
  no por feature — salvo `internal/business/usecase` e
  `internal/infrastructure/delivery/webapp/dto`, que agrupan casos de uso y
  DTOs en subpaquetes por feature (`usecase/items`, `dto/items`).
- `internal/business` (dominio y casos de uso) no depende de HTTP, JSON,
  PostgreSQL, tipos de `pgx` ni de detalles del adaptador en memoria.
- Entidades y tipos de valor en `internal/business/domain`; contratos que el
  negocio necesita de sus adaptadores (repositorios, `TxManager`, `Clock`) en
  `internal/business/gateway`; casos de uso en `internal/business/usecase`.
- Cada caso de uso es una interfaz de un solo método `Execute` más su
  implementación, en un mismo archivo dentro de su subpaquete
  (`usecase/sales/create_sale.go`). Depende de contratos de `gateway`, nunca
  de un adaptador concreto.
- Las reglas de negocio viven en el dominio (cálculo de totales, política de
  stock, transiciones de estado), no en repositorios ni handlers.
- Una operación que modifica más de un agregado (por ejemplo, venta + stock +
  movimientos) corre dentro de `TxManager.WithinTx(ctx, fn)`. Los repositorios
  no abren transacciones por su cuenta.
- Los errores que un caso de uso devuelve son `*domain.Error` (`Kind`, `Code`,
  `Message`, `Cause` opcional), no `errors.New(...)`. `Kind` es una
  clasificación de negocio (`KindInvalid`, `KindForbidden`, `KindConflict`,
  `KindNotFound`, `KindUnavailable`), nunca un status HTTP. `Cause` lleva el
  error subyacente para loguearlo sin exponerlo. Un error inesperado del
  adaptador se traduce en el borde del caso de uso a
  `domain.ErrStorageUnavailable` con el original en `Cause`.
- `response.WriteError` es el único lugar que mapea `Kind` a status HTTP,
  escribe `code`/`message` y loguea `Cause`. Los handlers nunca escriben su
  propio `switch` de errores. Un error que no es `*domain.Error` se loguea y
  sale como 500 genérico.
- Cada ruta HTTP tiene su propio handler con un único caso de uso como
  dependencia. El handler implementa `handler.HTTPHandler`
  (`Handle(w, r) error`), devuelve el error y deja que `handler.Adapt(h,
  timeout)` lo traduzca. `Adapt` acota `r.Context()` al timeout; ningún
  handler arma su propio `context.WithTimeout`. Un handler sin camino de error
  (por ejemplo `Live`) queda como `func(w, r)` simple.
- Los structs de request/response HTTP van en
  `internal/infrastructure/delivery/webapp/dto/<feature>`, con `package dto`.
  No se exponen entidades de dominio ni modelos de persistencia directamente.
- Decodificar cuerpos con el helper genérico `decodeJSON[T]`; devuelve un
  `*domain.Error` en vez de escribir la respuesta.
- El contenedor de dependencias (IoC) vive en
  `internal/infrastructure/delivery/webapp/dependencies`: solo construye
  objetos y elige el adaptador de persistencia según la configuración que
  recibe. No lee variables de entorno, no registra rutas ni arma el servidor.
- `internal/app` es el composition root: lee la configuración, pide el grafo a
  `dependencies`, registra rutas, construye el servidor, lo ejecuta y lo apaga
  ordenadamente.
- Los casos de uso reciben el `Actor` (id + rol) y piden un permiso con
  `Actor.Authorize(domain.Permission…)`, nunca un rol. Los roles, los
  permisos y la matriz rol → permiso viven solo en
  `internal/business/domain` (`permission.go`); un caso de uso no decide qué
  rol puede qué. Mientras no exista autenticación, un middleware de desarrollo
  inyecta el actor.
- Las operaciones de I/O reciben `context.Context` como primer parámetro.
- Las interfaces se definen cerca de quien las consume y se mantienen chicas;
  solo se crean para un límite real.

## Persistencia

- Hoy la persistencia es en memoria (`infrastructure/repository/memory`),
  seleccionada con `STORAGE_DRIVER=memory`.
- Todo repositorio nuevo se prueba con la suite de contrato compartida
  (`infrastructure/repository/repositorytest`), que debe pasar igual contra
  memoria y contra PostgreSQL.
- Los identificadores los asigna el repositorio (UUIDv7).
- Las bajas de catálogo y terceros son lógicas. Los movimientos de stock y las
  ventas nunca se borran ni se editan: se compensan.
- Cuando llegue PostgreSQL: `pgxpool` para acceso concurrente; migraciones SQL
  en `migrations/` con secciones `-- +goose Up` y `-- +goose Down`; nunca
  editar una migración ya aplicada, agregar una nueva.

## Dinero y cantidades

- Nunca usar punto flotante para dinero ni cantidades, ni en Go, ni en JSON,
  ni en TypeScript.
- `Money` es un entero en centavos; `Quantity` es un entero en milésimas.
- El backend recalcula todo importe; el frontend solo muestra una vista
  previa con las mismas reglas.

## Arquitectura del frontend

- Composición en `src/app`, código de cada funcionalidad en `src/features`,
  código genuinamente reutilizable en `src/shared`.
- `app → features → shared`. Una feature no importa archivos internos de otra;
  la composición entre features ocurre en `app`. Si dos o más features
  necesitan la misma lectura de la API, su cliente vive en `shared/api`.
- Estado del servidor con TanStack Query y claves en `shared/api/query-keys.ts`;
  estado visual cerca del componente que lo usa; Context solo para lo
  transversal (tema). Sin store global hasta que un flujo lo justifique.
- Validar respuestas de la API y formularios con esquemas Zod en el `types.ts`
  de cada feature; los tipos se infieren del esquema.
- Componentes y hooks puros; los efectos solo sincronizan con sistemas
  externos. Composición y variantes explícitas antes que componentes con
  muchos booleanos.
- Preferir imports directos a archivos barril.

## UI y experiencia

- La interfaz tiene que ser amigable, simple y moderna. Seguir los principios
  de `docs/plans/version-base.md` §5.
- Mobile-first: clases de Tailwind sin prefijo para la pantalla más chica y
  luego `sm:`/`md:`/`lg:`. Probar cada pantalla nueva a ~375 px y en
  escritorio.
- Usar los tokens del tema (`bg-primary`, `text-muted-foreground`, …), nunca
  colores crudos de Tailwind (`text-red-500`); lo hace cumplir la regla ESLint
  `local/no-raw-tailwind-colors`. Soportar modo claro y oscuro.
- Reutilizar el patrón de ABM (`PageHeader`, `SearchField`, `ResponsiveList`,
  `FormSheet`, `ConfirmDialog`, `EmptyState`) en vez de armar pantallas ad hoc.
- Accesibilidad WCAG 2.2 AA: labels reales, foco visible, navegación completa
  por teclado, objetivos táctiles de al menos 44 px.
- La pantalla de venta se opera completa con teclado y lector de código de
  barras.
- Textos en español rioplatense, claros y sin jerga técnica.

## Estructura prohibida

- No crear directorios genéricos `utils`, `helpers`, `common`, `controllers`,
  `services`, `repositories` ni `models`.
- No crear directorios vacíos ni abstracciones de reserva: se agregan cuando un
  caso de uso concreto las necesita.
- No introducir CQRS, buses de eventos, microservicios ni abstracciones
  genéricas sin un problema concreto.
- Una dependencia nueva tiene que justificar qué problema resuelve.

## Comentarios en el código

- No agregar comentarios que repitan lo que el código ya dice ni que narren la
  decisión de diseño; ese contexto va en `docs/` o en el PR.
- Reservarlos para lo que no es obvio: un workaround, una regla de negocio o
  fiscal que no se ve en el código, una advertencia.
- Antes de comentar, probar un nombre más claro o una función más chica.

## Tests y cobertura

- Todo comportamiento nuevo y toda corrección incluye tests en el mismo
  cambio.
- Backend: paquete estándar `testing` y `net/http/httptest`; sin librerías de
  aserciones salvo necesidad concreta.
- Casos de uso con fakes de `gateway` ubicados en
  `internal/business/gateway/gatewayfake`; handlers a través de su contrato
  HTTP observable; repositorios con la suite de contrato.
- Frontend: Vitest, React Testing Library, `@testing-library/jest-dom` y
  `@testing-library/user-event`. Consultar y verificar la UI como lo haría una
  persona; no testear estado interno, clases de Tailwind ni funciones
  privadas.
- Tests junto al código: `*_test.go`, `*.test.ts`, `*.test.tsx`.
- Mínimos sobre código de features: 80 % líneas, sentencias y funciones; 75 %
  ramas. Reglas críticas (cálculo de venta, stock, pagos): 90 %.
- La cobertura es una señal, no el objetivo: cubrir éxito, validaciones,
  errores, permisos y bordes.
- No bajar umbrales, borrar aserciones, agregar exclusiones amplias ni saltear
  tests para que una suite pase.
- Al agregar una feature backend, sumar sus paquetes al script raíz
  `test:backend:coverage`.

## Commits

- Conventional Commits simplificados: `<type>: <description>`.
- Tipos: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`.
- Descripción en inglés, imperativa, en minúsculas y sin punto final
  (`feat: add sale void use case`).

## Pull requests

- Ramas: los PR van contra `develop`; a `main` solo llegan desde `develop` o
  `hotfix/*`.
- Desde la CLI: `gh pr create --template .github/pull_request_template.md`.
- Toda descripción incluye: “Qué se hizo”, “Cómo probarlo”, “Cambios en base
  de datos”, “Capturas” e “Issue” (`Closes #NNN`, o `Refs #NNN` si no lo
  termina).
- Antes de pushear, revisar si el cambio desactualiza `README.md` o `docs/` y
  actualizarlos en el mismo PR.

## Verificación

- Frontend: `pnpm --filter @market/frontend typecheck`, `lint`, `test` y `build`.
- Backend: `go test ./...` y `go vet ./...` desde `apps/backend`.
- Todo junto: `pnpm test` y `pnpm test:coverage` desde la raíz.
- Entorno: `docker compose config` y `docker compose up --build`.
- CI corre build, tests, cobertura y auditoría de dependencias en cada PR
  (`docs/ci.md`); correrlos localmente antes de pushear ahorra idas y vueltas.
