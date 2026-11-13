# Arquitectura y estructura

Market aplica los principios útiles de Clean Architecture —límites claros,
dependencias hacia el negocio, adaptadores reemplazables— sin capas que
todavía no aporten. Las reglas obligatorias están en
[AGENTS.md](../AGENTS.md); la estructura objetivo completa, en
[plans/version-base.md §3](plans/version-base.md#3-arquitectura). Acá se
documenta lo que **existe hoy** y las decisiones vigentes.

## Vista general

```mermaid
flowchart LR
    browser["Navegador"] -->|"/ (Vite)"| frontend["Frontend React/Vite"]
    frontend -->|"/api (mismo origen)"| backend["Backend Go"]
    backend --> usecases["Casos de uso"]
    usecases --> gateway["Contratos (gateway)"]
    memory["repository/memory"] -.implementa.-> gateway
```

## Backend

```text
apps/backend/
├── cmd/server/main.go                     # señales + app.New(ctx).Run(ctx)
└── internal/
    ├── app/app.go                         # composition root: config, contenedor, datos demo, rutas, servidor
    ├── business/
    │   ├── domain/                        # Error, Actor, Permission, Money, Quantity, Percent, Page, Category…
    │   ├── gateway/                       # TxManager y un contrato de repositorio por agregado
    │   │   └── gatewayfake/               # fakes de los contratos para tests de casos de uso
    │   └── usecase/                       # categories, items, customers, suppliers, sellers: create, list, get, update, delete; stock: existencias, movimientos, ingreso, ajuste y stock inicial; sales: alta, listado, detalle y anulación; dashboard: resumen del día
    └── infrastructure/
        ├── environments/config.go         # PORT, STORAGE_DRIVER, CORS_ALLOWED_ORIGIN, SEED_DEMO_DATA, STOCK_ALLOW_NEGATIVE
        ├── demo/seed.go                   # datos demo, cargados a través de los casos de uso
        ├── repository/
        │   ├── memory/                    # Store + TxManager + un repositorio por agregado
        │   └── repositorytest/            # suite de contrato compartida por todos los adaptadores
        └── delivery/webapp/
            ├── dependencies/              # IoC: repositorios → casos de uso → handlers
            ├── dto/                       # health, pagination (PageResponse[T]), categories, items, customers, suppliers, sellers, stock, sales, dashboard
            ├── handler/                   # un handler por ruta + Adapt, decodeJSON, pageQuery
            ├── middleware/                # DevActor, Recover, RequestLog
            ├── response/                  # WriteJSON, WriteError (único mapeo Kind → status)
            ├── route/                     # rutas, timeouts y route.Handlers
            └── server/                    # http.Server y CORS opcional
```

Cadena de middlewares (de afuera hacia adentro): `RequestLog` → `Recover` →
`WithCORS` (solo si hay `CORS_ALLOWED_ORIGIN`) → `DevActor` → mux.

### Dirección de dependencias

```mermaid
flowchart LR
    main["cmd/server"] --> app["internal/app"]
    app --> environments["infrastructure/environments"]
    app --> deps["delivery/webapp/dependencies"]
    app --> demo["infrastructure/demo"]
    app --> route["delivery/webapp/route"]
    app --> server["delivery/webapp/server"]
    app --> middleware["delivery/webapp/middleware"]
    deps --> memory
    deps --> route
    demo --> usecase
    route --> handler["delivery/webapp/handler"]
    handler --> response["delivery/webapp/response"]
    handler --> usecase["business/usecase"]
    usecase --> gateway["business/gateway"]
    usecase --> domain["business/domain"]
    memory["repository/memory"] -.implementa.-> gateway
    response --> domain
    middleware --> domain
```

- Un handler no ejecuta lógica de negocio ni accede a la persistencia.
- Un repositorio no toma decisiones de negocio.
- Un caso de uso no conoce `http.Request`, JSON, `sync.Mutex` ni `pgx`.
- `dependencies` arma el grafo y elige el adaptador según `STORAGE_DRIVER`.
  `Container.Handlers()` devuelve el `route.Handlers` ya armado: construye
  objetos, pero no registra rutas (eso lo hace `internal/app`).
- Los datos demo (`SEED_DEMO_DATA`, activo por defecto) pasan por los mismos
  casos de uso que la app, así que respetan sus reglas.

### Convenciones HTTP

- Endpoints de negocio bajo `/api/v1`; JSON en `camelCase` e inglés.
- Errores: `{ "code": "…", "message": "…" }`. Cualquier ruta o método bajo
  `/api/` que no exista responde `404 route_not_found` con ese mismo formato
  (también un método no soportado: se prefiere un único formato de error al
  405 en texto plano del mux).
- Listados: `?q=&page=&limit=` →
  `{ "items": [], "total": 0, "page": 1, "limit": 25 }` (`PageResponse[T]`).
  `items` nunca es `null`.
- Eliminación: `DELETE /{id}` borra el registro y responde `204` sin cuerpo;
  un ID inexistente es `404`. Si otro registro lo referencia, responde `409`
  (`<entidad>_in_use`).
- Edición: categorías usan `PATCH` parcial (un campo ausente o `null` no
  cambia; un texto opcional vacío lo borra). Artículos, clientes, proveedores y vendedores usan
  `PUT`, que reemplaza todos los campos: su formulario siempre manda el
  registro completo y así un opcional en `null` se borra.
- Montos y cantidades viajan como enteros: centavos, centésimas de punto
  porcentual y milésimas de unidad.
- `POST /api/v1/sales` exige el header `Idempotency-Key`: el frontend genera
  la clave, la reutiliza mientras el alta falle y la rota al confirmar. Un
  reintento con la misma clave y el mismo cuerpo responde la venta original
  (`201`, sin registrarla de nuevo); con otro cuerpo, `409
  idempotency_key_conflict`. La clave vale por actor: otro actor que mande la
  misma clave registra su propia venta y nunca recibe la ajena.
- `Adapt(h, timeout)` acota el contexto al timeout de la ruta y el cuerpo a
  1 MiB (`handler.MaxBodyBytes`). Los deadlines del `http.Server` derivan del
  mayor timeout registrado (`route.MaxHandlerTimeout`).
- `GET /health/live` responde `{ "status": "ok" }` para chequeos de
  infraestructura (Docker, un balanceador); la app no lo consulta. `/health/ready` (con
  chequeo de la base) llega con PostgreSQL.

### Persistencia en memoria

`memory.Store` guarda todas las tablas detrás de un único `sync.Mutex`:

- `WithinTx(ctx, fn)` toma el lock durante toda la unidad de trabajo, así que
  las transacciones son serializables. Un `ctx` que ya está dentro de una
  transacción del mismo store se une a ella en vez de bloquearse.
- Cada escritura (`put`, `remove`, `appendRow`, `set`) registra cómo
  deshacerse. Si `fn` devuelve error o entra en pánico, los pasos se deshacen
  en orden inverso y las tablas quedan exactamente como estaban.
- `NewStore` arranca vacío salvo por el cliente de sistema **Consumidor
  Final** (`domain.FinalConsumer`, con ID fijo `domain.FinalConsumerID`); en
  PostgreSQL lo insertará una migración.
- El número de venta sale de un contador (`lastSaleNumber`, escrito con
  `set`), que hace de secuencia. A diferencia de una secuencia, una unidad de
  trabajo que falla devuelve su número; la suite de contrato no depende de
  eso.
- Los movimientos de stock viven en una tabla que solo crece
  (`appendRow`): su orden es el de registro y ningún repositorio los edita
  ni los borra.
- Las operaciones de stock leen la existencia, dejan que el dominio calcule
  el movimiento y lo registran dentro del mismo `WithinTx`; con el lock del
  store, dos ingresos simultáneos del mismo artículo nunca pisan su saldo
  (la suite de contrato lo prueba con 20 unidades de trabajo concurrentes).
  En PostgreSQL, `GetLevel` tendrá que bloquear la fila (`FOR UPDATE`).
- Una venta corre entera en un `WithinTx`: busca la clave de idempotencia,
  lee cliente, vendedor y artículos, deja que el dominio calcule la venta,
  la guarda (el repositorio le asigna ID y número) y registra un movimiento
  `sale` por línea. Si una línea no tiene stock, no queda ni la venta ni
  ningún movimiento. La anulación también: guarda el estado `voided` y un
  movimiento `sale_void` por línea en la misma unidad de trabajo.
- Las tablas guardan valores, y los repositorios copian los punteros y
  slices que guardan o devuelven (por ejemplo `cloneCategory`), para que
  nadie mute una tabla fuera del lock.
- Las restricciones que en PostgreSQL serían índices o claves foráneas
  (nombre único de categoría, código y código de barras únicos, la
  categoría de un artículo tiene que existir y no se puede borrar mientras
  tenga artículos, los proveedores de un artículo tienen que existir y no
  se pueden borrar mientras un artículo los tenga, un documento por cliente,
  un CUIT por proveedor, un código por vendedor, el artículo y el proveedor
  de un movimiento de stock tienen que existir y no se pueden borrar después,
  una venta por actor y clave de idempotencia, y el cliente, el vendedor y los
  artículos de una venta tienen que existir y el cliente y el vendedor no se
  pueden borrar después)
  las verifica el repositorio dentro del lock, y la suite de contrato prueba
  que se cumplan también bajo concurrencia.
- Los IDs son UUIDv7, ordenados por tiempo como los de `uuidv7()` en
  PostgreSQL 18.

Se eligió un registro de deshacer en lugar de copiar todo el estado en cada
transacción (como proponía el plan): cuesta lo mismo que las escrituras
hechas, no lo que mide la base completa.

## Frontend

```text
apps/frontend/src/
├── app/                      # composición: router, layout, navegación, tema, providers
│   ├── router.tsx            # / (punto de venta), /articulos, /categorias, /clientes, /proveedores, /vendedores, /resumen y página 404
│   ├── lazy-pages.ts         # cada página con React.lazy: un chunk por ruta
│   ├── ItemsRoute.tsx        # compone artículos con las categorías, los proveedores y el stock (ver abajo)
│   ├── PosRoute.tsx          # compone el punto de venta con las categorías, el catálogo, los clientes y los vendedores
│   ├── AppLayout.tsx         # sidebar (escritorio) + header + barra inferior (celular)
│   ├── PageSkeleton.tsx      # skeleton del Suspense del layout mientras llega una página
│   ├── navigation.ts         # secciones; las que no entran en la barra del celular van a "Más"
│   ├── ThemeProvider.tsx     # claro / oscuro / automático, guardado en localStorage
│   └── Providers.tsx         # QueryClient + tema + toasts
├── features/
│   ├── categories/           # api/, hooks/, components/, pages/, types.ts
│   ├── items/                # ídem; recibe las categorías por props
│   ├── customers/            # ídem
│   ├── suppliers/            # ídem
│   ├── sellers/              # ídem
│   ├── stock/                # ajuste, stock inicial, ingreso de mercadería (con buscador de artículos) e historial; se usan desde Artículos
│   ├── sales/                # punto de venta: catálogo con buscador y lector de barras, carrito guardado en localStorage y cobro con varios medios de pago
│   └── dashboard/            # resumen del día (/resumen): ventas, total por medio de pago y stock bajo; pide el día local de la persona
├── shared/
│   ├── api/                  # apiFetch + ApiError (Zod), pagination, query-keys
│   ├── config/env.ts         # VITE_API_URL opcional
│   ├── hooks/                # useDebouncedValue
│   ├── lib/                  # cn, money (centavos ↔ texto), quantity (milésimas ↔ texto, unidades), date (fecha y hora local), percent (centésimas ↔ texto), document (DNI y CUIT)
│   └── ui/                   # primitivos shadcn/Base UI + kit de ABM (ver abajo)
├── test/                     # setup de Vitest y renderWithProviders
├── index.css                 # tokens de tema claro/oscuro (oklch)
└── main.tsx
```

- `app → features → shared`; una feature no importa a otra. Ejemplo:
  artículos necesita las categorías para el filtro y el formulario, y los
  proveedores para el formulario; `app/ItemsRoute` los obtiene con
  `useCategoryOptions` y `useSupplierOptions` y se los pasa a `ItemsPage`
  por props. Esos hooks piden todas las páginas de la lista (`listAllPages`,
  de a 100), así que ninguna opción queda afuera por pasar de la primera página.
- La grilla de artículos muestra el stock de cada uno (columna **Stock** y
  tarjetas del celular, con la marca «Stock bajo» en texto, no solo color).
  La grilla está paginada y ordenada por nombre, así que no se cruza con la
  lista de existencias página por página: `ItemsPage` pide las existencias
  de los artículos que está mostrando (`GET /api/v1/stock?itemId=…`) con el
  hook `useStockQuantities` que le pasa `app/ItemsRoute` (envuelve
  `useStockLevelsOf` de stock; artículos no importa la feature de stock).
  Si esa lectura falla, la grilla sigue mostrando los artículos y el stock
  aparece como «—».
- El stock se maneja desde Artículos, no hay pantalla propia: cada fila
  ofrece **Ajustar stock** (con motivo) o, si el artículo no tiene
  movimientos, **Cargar stock inicial**, y **Movimientos**; el encabezado
  ofrece **Ingresar mercadería** y la grilla se filtra con **Solo stock bajo**
  (`GET /api/v1/items?lowStock=true`). Artículos no importa la feature de
  stock: `app/ItemsRoute` le pasa `renderStockActions` (que dibuja
  `StockItemActions`) y `headerActions` (`ReceiveStock`). El ítem sin
  movimientos se reconoce porque su existencia trae `updatedAt: null`.
- El ingreso de mercadería busca los artículos en el endpoint de stock
  (`GET /api/v1/stock?q=`), que ya trae código, nombre, unidad y
  existencia: stock no depende de la feature de artículos. Enter elige el
  primer resultado, así un lector de código de barras completa la línea.
- El inicio (`/`) es el **punto de venta** (`features/sales`).
  `app/PosRoute` le pasa las categorías y `listItems` de la feature de
  artículos como `searchCatalog`: el catálogo comparte la caché
  `queryKeys.items`, así que editar un artículo también refresca la
  pantalla de venta. Enter en el buscador busca enseguida (sin esperar el
  debounce) y agrega el artículo con ese código o código de barras exacto,
  o el único que coincide: así funciona un lector de código de barras. Los
  artículos por peso piden la cantidad en un modal.
- El carrito vive en el navegador (`localStorage`, clave `market.pos.cart`,
  validado con Zod al leerlo): recargar la página no lo pierde. Su total es
  una vista previa con las mismas reglas de redondeo que el backend
  (`features/sales/lib/cart.ts`, con `BigInt`); el backend recalcula todo al
  cobrar. Quitar una línea o vaciar el carrito se deshace desde el aviso
  (`useToast().undoable`), sin pedir confirmación.
- **Cliente y vendedor**: `app/PosRoute` también le pasa `listCustomers` y
  `listSellers`. El cliente va con el carrito (sin elegir es Consumidor
  Final, el que usa el backend por defecto, y vuelve a serlo al vaciar la
  venta o al registrarla); el vendedor se recuerda entre ventas
  (`market.pos.seller`) y puede no haber. Los dos se eligen en el mismo
  modal con búsqueda (`PartyPicker`): Enter elige el primero, o el vendedor
  del código tipeado. La pantalla de clientes muestra a Consumidor Final
  como “Del sistema”, sin editar ni eliminar.
- Cobrar (`F9` o el botón del carrito) abre `CheckoutDialog`: uno o varios
  medios de pago con su importe (el efectivo viene con el total y tiene
  atajos «Exacto», $10.000 y $20.000), cuánto falta o el vuelto, y una
  referencia opcional para los medios que no son efectivo. Las reglas de
  pago (`features/sales/lib/payments.ts`) son las del backend: la suma tiene
  que cubrir el total y solo el efectivo puede pasarlo. Al confirmar envía
  `POST /api/v1/sales` con `expectedTotal` y un `Idempotency-Key`; la clave se
  reutiliza mientras se reintenta la misma venta (mismo cuerpo) y se renueva
  si el cuerpo cambia o la venta se registra. Con la venta registrada, el
  carrito se vacía, se muestran el número y el vuelto, y se vuelven a leer
  artículos y stock. El diálogo no se cierra mientras la solicitud está en
  curso. Envía el cliente y el vendedor elegidos (`customerId` en `null` es
  Consumidor Final; `sellerId` en `null`, sin vendedor). Todavía sin
  descuentos en la pantalla.
- El resumen (`/resumen`, `features/dashboard`) pide
  `GET /api/v1/dashboard/summary` con el inicio y el fin del día local
  (`lib/today.ts`). Su clave (`queryKeys.dashboard`) se invalida al vender y
  al mover stock (ingreso, ajuste o stock inicial); anular una venta tiene que
  invalidarla también cuando exista esa pantalla. El stock bajo se marca con
  texto y ícono, no solo con color.
- Atajos del punto de venta (`useShortcuts`, solo en esa pantalla): `F2`
  vuelve al buscador, `F4` elige el vendedor, `F6` el cliente y `F9` cobra;
  las flechas recorren el catálogo y los resultados, y `Escape` cierra el
  modal abierto o vuelve al buscador y lo limpia.
- `formatMoney`/`parseMoney` y `formatQuantity`/`parseQuantity` son los
  únicos que convierten montos y cantidades entre enteros y texto. Leen lo
  que se tipea en Argentina (`1.234,50`).
- `apiFetch(path, schema, options)` valida cada respuesta con un esquema Zod:
  si el contrato con el backend se desalinea, falla con `invalid_response`
  en vez de renderizar valores indefinidos.
- Estado del servidor con TanStack Query; claves en `shared/api/query-keys.ts`.
  Los listados usan `keepPreviousData` para no parpadear al buscar o
  paginar, y cada mutación invalida la clave raíz de su entidad.

### Kit de ABM (`shared/ui`)

| Componente | Para qué |
| --- | --- |
| `PageHeader` | Título y acción principal de la pantalla. |
| `SearchField` | Búsqueda con ícono y botón para limpiar; el placeholder es el propio label (corto, para que entre en el celular) y debajo muestra por qué campos busca (`hint`). Se combina con `useDebouncedValue`. |
| `ResponsiveList` | Tabla desde `md:` y tarjetas en el celular; skeleton mientras carga. |
| `Pagination` | “1–25 de 40” + Anterior / Siguiente. |
| `FormDialog` | Modal centrado para crear y editar cualquier entidad; el contenido largo se desplaza adentro. `size="wide"` para formularios en dos columnas. Al cerrar, devuelve el foco al botón que lo abrió (`useReturnFocus`). Si el formulario tiene cambios sin guardar (`useUnsavedChanges(values)`), cerrarlo con Escape, la X, tocando afuera o «Cancelar» (`useFormDialog().requestClose`) pregunta «¿Descartar los cambios?»; cerrarlo desde la página después de guardar no pregunta. |
| `NativeSelect` | `<select>` nativo con estilo: teclado, lector de pantalla y selector del celular funcionan sin más. |
| `ConfirmDialog` | Confirmación antes de eliminar (`role="alertdialog"`); también devuelve el foco al botón que la abrió. |
| `EmptyState` | Vacío, sin resultados o error, siempre con una salida. |
| `ToastProvider` / `useToast` | Avisos de éxito o error (Toast de Base UI, sin dependencias extra). |
| `Field`, `Input`, `Textarea` | Campos con label, ayuda y error; 44 px de alto en el celular. |

## Receta de ABM

Los pasos que siguió categorías y que repite cada entidad nueva (artículos,
clientes, proveedores, vendedores).

**Backend**

1. `domain/<entidad>.go`: entidad, constructor `New…` que valida, cambios
   parciales (`…Changes` + `Apply`, para `PATCH`) o reemplazo completo
   (`Replace`, para `PUT`), errores `*domain.Error`
   con mensajes en español y tests de tabla. Si la entidad necesita permisos
   nuevos, se agregan a la matriz de `domain/permission.go` y a su test.
2. `gateway/<entidad>_repository.go`: contrato (`Create`, `Get`, `Update`,
   `Delete`, `List` con búsqueda y página) documentando qué errores devuelve.
3. `gateway/gatewayfake/`: fake que registra llamadas y devuelve lo
   configurado.
4. `usecase/<feature>/`: un archivo por caso de uso (`Execute`) que pide su
   permiso con `Actor.Authorize(domain.Permission…)` (nunca un rol) y
   traduce los errores del adaptador con `domain.FromStorage`.
5. `repository/repositorytest/<entidad>.go`: la suite de contrato (unicidad,
   no encontrado, eliminación, búsqueda, orden, paginado, copias,
   concurrencia).
6. `repository/memory/<entidad>.go`: el adaptador, con su test que corre la
   suite.
7. `dto/<feature>/`: requests, response y conversión desde el dominio.
8. `handler/`: un handler por ruta, con tests de su contrato HTTP.
9. `route.Handlers` + `RegisterRoutes`, `dependencies`, datos demo y el
   paquete nuevo en `test:backend:coverage:profile`.

**Frontend**

1. `features/<feature>/types.ts`: esquemas Zod de la respuesta y del
   formulario, y el mapa de códigos de error del backend a campos.
2. `api/`: una función por endpoint con `apiFetch` y su esquema.
3. `hooks/`: `use<Entidades>` (listado) y las mutaciones que invalidan la
   clave de la entidad.
4. `components/`: formulario (validación con Zod + errores del backend en el
   campo correspondiente; llama a `useUnsavedChanges(values)` y su botón
   «Cancelar» a `useFormDialog().requestClose`), listado sobre `ResponsiveList`.
5. `pages/<Entidades>Page.tsx`: compone el kit (búsqueda, listado,
   paginación, `FormDialog` para crear y editar, `ConfirmDialog` para
   eliminar, toasts).
6. Página diferida en `app/lazy-pages.ts`, ruta en `app/router.tsx` con su
   título de pestaña (`handle: title('…')`, que `usePageTitle` aplica como
   "<título> · Market") y sección en `app/navigation.ts`.
7. Tests de la página contra una API falsa en memoria que respeta las mismas
   reglas que el backend (ver `CategoriesPage.test.tsx`).

### Tema y diseño

- Tokens en `index.css` (`--primary`, `--success`, `--warning`, …) con
  variantes clara y oscura; la regla ESLint `local/no-raw-tailwind-colors`
  prohíbe colores crudos de Tailwind.
- Tipografía Inter (variable, servida desde el propio bundle: la CSP no
  permite fuentes externas).
- El tema se aplica antes del primer render para evitar el destello claro.

### Seguridad

La CSP se inyecta en build con un `<meta>` desde
`vite-plugins/content-security-policy.ts`: `connect-src 'self'` (más el
origen exacto de `VITE_API_URL` si se define), scripts solo propios y
`'unsafe-inline'` en estilos únicamente en desarrollo.

## Decisiones vigentes

- **Mismo origen para frontend y API**: proxy de Vite en desarrollo; en
  producción, reverse proxy o el binario Go sirviendo el build. CORS solo con
  `CORS_ALLOWED_ORIGIN`.
- **Unidad de trabajo explícita** (`gateway.TxManager`): la atomicidad no la
  decide cada repositorio.
- **Registro de deshacer** en el store en memoria (ver arriba).
- **Sin autenticación todavía**; los casos de uso igual reciben un `Actor`.
- **Go 1.27** desde el inicio (LoteosAPP sigue en 1.26).
