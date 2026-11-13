# Dominio

Las reglas completas de la versión base están en
[plans/version-base.md §4](plans/version-base.md#4-dominio-lógica-core). Este
documento describe lo que **ya está implementado** y crece con cada fase.

## Glosario

El código, la API y la base usan los términos en inglés; la UI, en español.

| Español (UI) | Código / API |
| --- | --- |
| Categoría | `Category` |
| Artículo | `Item` |
| Cliente | `Customer` |
| Proveedor | `Supplier` |
| Vendedor | `Seller` |
| Existencia | `StockLevel` |
| Movimiento de stock | `StockMovement` |
| Ingreso de mercadería | `StockEntry` |
| Ajuste | `StockAdjustment` |
| Venta | `Sale` |
| Línea | `SaleLine` |
| Pago | `Payment` |
| Descuento (de línea) | `Discount` |
| Descuento general (del carrito) | `CartDiscount` |
| Anulación | `Void` |
| Resumen del día | `SalesSummary` |

## Tipos de valor (`internal/business/domain`)

- **`Money`**: `int64` en centavos de ARS. Nunca punto flotante, tampoco en
  JSON ni en TypeScript.
- **`Quantity`**: `int64` en milésimas de la unidad (0,250 kg = `250`; tres
  unidades = `3000`).
- **`Unit`**: `unit`, `kg`, `l`, `m`. Un artículo por unidad (`unit`) solo
  acepta cantidades enteras; toda cantidad vendida es positiva.
- **`Percent`**: centésimas de punto porcentual (12,5 % = `1250`); válido de
  0 a 100 %.
- **Redondeo**: `Money.Times(Quantity)` y `Money.Percent(Percent)` redondean
  *half away from zero* al centavo y calculan sin desbordar; un resultado
  fuera de `int64` es `amount_out_of_range`.
- **`PageRequest` / `Page[T]`**: página desde 1, `limit` por defecto 25 y
  máximo 100; fuera de rango es `invalid_page`.
- **`NormalizeText`**: minúsculas, sin acentos y con espacios colapsados.
  Base de búsquedas y de unicidad de nombres.

## Categorías

| Campo | Regla |
| --- | --- |
| `name` | Obligatorio, hasta 60 caracteres, sin espacios al borde. Único sin distinguir mayúsculas, acentos ni espacios (`NameKey`). |
| `description` | Opcional, hasta 200 caracteres. En blanco se guarda como ausente. |
| `createdAt` / `updatedAt` | UTC, los asigna el caso de uso con la hora del servidor (`time.Now().UTC()`). |

Permisos: listar y consultar piden `catalog:view`; crear, editar y eliminar
piden `catalog:manage` (ver la matriz en
[Actor, roles y permisos](#actor-roles-y-permisos)).

Eliminar borra la categoría y libera su nombre; no se puede deshacer. Si la
categoría tiene artículos, falla con `category_in_use` (`409`): hay que
cambiarles la categoría (o eliminarlos) antes.

Códigos de error: `category_not_found`, `category_name_required`,
`category_name_too_long`, `category_description_too_long`,
`category_name_in_use`, `category_update_empty`, `category_in_use`.

## Artículos

Un artículo es un producto que se vende. Market no vende servicios: todo
artículo controla stock y tiene stock mínimo.

| Campo | Regla |
| --- | --- |
| `code` | Código interno (SKU). Obligatorio, hasta 30 caracteres, único sin distinguir mayúsculas ni acentos. |
| `barcode` | Opcional. Sin espacios, hasta 32 caracteres, único si se carga (dos artículos pueden no tener). |
| `name` | Obligatorio, hasta 100 caracteres. |
| `description` | Opcional, hasta 500 caracteres. |
| `categoryId` | Obligatoria y tiene que existir (`item_category_not_found` si no). |
| `supplierIds` | A quiénes se les compra: ninguno, uno o varios. Cada uno tiene que existir (`item_supplier_not_found` si no); los repetidos se descartan. Siempre viaja como lista (`[]` si no hay). |
| `unit` | `unit` (por unidad), `kg`, `l` o `m`. |
| `salePrice` | Centavos, mayor a 0 y hasta `MaxItemAmount` (mil millones de pesos). **Incluye IVA.** |
| `cost` | Centavos, opcional, entre 0 y `MaxItemAmount`. |
| `taxRate` | Alícuota de IVA en centésimas de punto: `0`, `1050`, `2100` o `2700`. Informativa hasta facturar. |
| `minStock` | Milésimas de la unidad, entre 0 y un millón de unidades; entero si se vende por unidad. Umbral de la alerta de stock bajo (Fase 3). |
| `discontinued` | `false` por defecto. Un artículo con movimientos de stock no se puede eliminar (`item_in_use`): para retirarlo se marca discontinuado. No se vende (`sale_item_discontinued`, `409`) y `GET /api/v1/items` lo deja afuera salvo con `includeDiscontinued=true`, pero conserva su historial; las ventas anteriores y sus anulaciones no cambian. |

Editar reemplaza todos los campos (`PUT`): el formulario siempre manda el
artículo completo, y así un opcional enviado como `null` se borra.

Permisos: los de catálogo (`catalog:view` para listar y consultar,
`catalog:manage` para crear, editar y eliminar).

Eliminar un artículo que tiene movimientos de stock falla con `item_in_use`
(`409`): los movimientos nunca se borran, así que el artículo tampoco.

Códigos de error: `item_not_found`, `item_in_use`, `item_code_required`,
`item_code_too_long`, `item_code_in_use`, `item_barcode_invalid`,
`item_barcode_in_use`, `item_name_required`, `item_name_too_long`,
`item_description_too_long`, `item_category_required`,
`item_category_not_found`, `item_supplier_not_found`, `item_unit_invalid`, `item_price_invalid`,
`item_cost_invalid`, `item_tax_rate_invalid`, `item_min_stock_invalid`.

## Clientes

| Campo | Regla |
| --- | --- |
| `name` | Nombre o razón social. Obligatorio, hasta 120 caracteres. |
| `documentType` | `dni`, `cuit`, `cuil` o `none` (sin documento). |
| `documentNumber` | Se guarda solo con dígitos: `20-12345678-9` y `20123456789` son el mismo CUIT. DNI de 7 u 8 dígitos; CUIT y CUIL de 11 (no se verifica el dígito verificador). Vacío si no hay documento. Único por tipo: dos clientes sin documento nunca chocan. |
| `taxCondition` | Condición frente al IVA: `final_consumer`, `monotax`, `registered` o `exempt`. Las tres últimas exigen CUIT. |
| `email` | Opcional, con forma de email, hasta 254 caracteres. |
| `phone` | Opcional, hasta 30 caracteres. |
| `address` | Opcional, hasta 200 caracteres. |
| `notes` | Opcional, hasta 500 caracteres. |

Editar reemplaza todos los campos (`PUT`), como en artículos. La búsqueda
encuentra por nombre, por los dígitos del documento (`40.111` encuentra el
DNI 40111222), por email y por teléfono.

Permisos: `customers:view` para listar y consultar, `customers:create` para
dar de alta (el cajero también, para cargar un cliente durante una venta) y
`customers:manage` para editar y eliminar.

**Consumidor Final** es un cliente de sistema: existe desde que arranca el
almacenamiento (también sin datos demo), siempre con el mismo ID
(`00000000-0000-7000-8000-000000000001`, `domain.FinalConsumerID`), sin
documento y como consumidor final. Aparece en el listado con
`"system": true` y no se edita ni se elimina (`customer_is_system`, `409`).
Una venta sin cliente se le hace a él.

Eliminar un cliente con ventas falla con `customer_in_use` (`409`): la venta
guarda una copia de sus datos, pero sigue apuntando a él.

Códigos de error: `customer_not_found`, `customer_name_required`,
`customer_name_too_long`, `customer_document_type_invalid`,
`customer_document_invalid`, `customer_document_in_use`,
`customer_tax_condition_invalid`, `customer_tax_condition_requires_cuit`,
`customer_email_invalid`, `customer_phone_too_long`,
`customer_address_too_long`, `customer_notes_too_long`,
`customer_is_system`, `customer_in_use`.

## Proveedores

| Campo | Regla |
| --- | --- |
| `businessName` | Razón social. Obligatoria, hasta 120 caracteres. |
| `cuit` | Opcional. Se guarda solo con dígitos y lleva 11 (no se verifica el dígito verificador). Único si se carga: dos proveedores sin CUIT nunca chocan. |
| `contactName` | Persona de contacto. Opcional, hasta 120 caracteres. |
| `phone` | Opcional, hasta 30 caracteres. |
| `email` | Opcional, con forma de email, hasta 254 caracteres. |
| `address` | Opcional, hasta 200 caracteres. |
| `notes` | Opcional, hasta 500 caracteres. |

Editar reemplaza todos los campos (`PUT`). La búsqueda encuentra por razón
social, persona de contacto, dígitos del CUIT, email y teléfono.

Permisos: `suppliers:view` para listar y consultar y `suppliers:manage` para
crear, editar y eliminar. El cajero no tiene ninguno: los proveedores se
usan al ingresar mercadería (Fase 3), que hace el encargado.

Eliminar un proveedor asignado a algún artículo falla con `supplier_in_use`
(`409`): hay que quitarlo de esos artículos antes. Un proveedor con
ingresos de mercadería registrados tampoco se elimina (mismo código, con un
mensaje que lo explica).

Códigos de error: `supplier_not_found`, `supplier_business_name_required`,
`supplier_business_name_too_long`, `supplier_cuit_invalid`,
`supplier_cuit_in_use`, `supplier_in_use`, `supplier_contact_name_too_long`,
`supplier_email_invalid`, `supplier_phone_too_long`,
`supplier_address_too_long`, `supplier_notes_too_long`.

## Vendedores

Un vendedor es quien atiende la venta, no un usuario de la app (a futuro un
usuario podrá vincularse a un vendedor).

| Campo | Regla |
| --- | --- |
| `name` | Obligatorio, hasta 80 caracteres. |
| `code` | Lo que se tipea para elegirlo en la venta. De 1 a 6 letras o números, sin espacios; se guarda en mayúsculas (`jp` es `JP`) y es único. |
| `commissionRate` | Opcional, en centésimas de punto porcentual (2,5 % = `250`), de 0 a 100 %. Informativa hasta que haya reportes de comisiones. |

Editar reemplaza todos los campos (`PUT`). La búsqueda encuentra por nombre
o código.

Permisos: `sellers:view` para listar y consultar (también el cajero, que
elige el vendedor al cobrar) y `sellers:manage` para crear, editar y
eliminar. Eliminar un vendedor que tenga ventas falla con `seller_in_use`
(`409`).

Códigos de error: `seller_not_found`, `seller_name_required`,
`seller_name_too_long`, `seller_code_required`, `seller_code_invalid`,
`seller_code_in_use`, `seller_commission_invalid`, `seller_in_use`.

## Stock

Todo artículo lleva stock. La existencia (`StockLevel`) de un artículo es
cuánto hay y **solo cambia mediante un movimiento** (`StockMovement`); no hay
“editar stock”. Un artículo sin movimientos tiene existencia cero.

| Campo del movimiento | Regla |
| --- | --- |
| `type` | `initial` (stock inicial), `entry` (ingreso de mercadería), `adjustment` (ajuste), `sale` (venta) y `sale_void` (anulación de una venta). |
| `quantity` | Milésimas de la unidad, con signo: positiva si entra, negativa si sale. Entera si el artículo se vende por unidad. Hasta un millón de unidades por movimiento. |
| `balanceAfter` | Existencia que deja el movimiento. La existencia no pasa de mil millones de unidades (ni baja de menos mil millones). |
| `reason` | Motivo, obligatorio en los ajustes, hasta 200 caracteres. |
| `reference` | Comprobante opcional del ingreso (remito o factura), hasta 60 caracteres. En una venta y en su anulación, `Venta <número>`. |
| `supplierId` / `unitCost` | Opcionales, solo en ingresos. |
| `actorId` / `createdAt` | Quién y cuándo (UTC). |

Los movimientos **nunca se editan ni se borran**: un error se corrige con
otro movimiento. Por eso un artículo o un proveedor con movimientos no se
puede eliminar (`item_in_use`, `supplier_in_use`).

- **Stock inicial** (`initial`): abre el stock de un artículo y solo se
  permite como su primer movimiento (`stock_already_initialized` si ya
  tiene). Se carga con `POST /api/v1/stock/initial` (artículo y cantidad mayor a cero) y, en la pantalla de Artículos, con «Cargar stock inicial» para los artículos sin movimientos; también lo usan los datos demo.
- **Ingreso de mercadería** (`POST /api/v1/stock/entries`): una o varias
  líneas (hasta 100), cada una con artículo, cantidad mayor a cero y costo
  unitario opcional; proveedor y comprobante opcionales. Un artículo no se
  repite en dos líneas. Todo el ingreso se registra o no se registra nada
  (`TxManager.WithinTx`). **Si una línea trae costo, pasa a ser el costo del
  artículo.**
- **Ajuste** (`POST /api/v1/stock/adjustments`): cantidad con signo distinta
  de cero y motivo obligatorio. Un ajuste negativo no puede dejar la
  existencia por debajo de cero (`insufficient_stock`, `409`, nombrando el
  artículo); uno positivo siempre se acepta.
- **Stock bajo**: `minStock > 0 && quantity <= minStock`
  (`domain.IsLowStock`). Un mínimo en cero apaga la alerta.
- **Existencias** (`GET /api/v1/stock?q=&categoryId=&itemId=&lowStock=true`):
  todos los artículos con su existencia, ordenados por nombre; `q` busca por
  nombre, código o código de barras, `itemId` (repetible) deja solo los
  artículos con esos IDs (uno que no existe se ignora) y `lowStock=true`
  deja solo los de stock bajo. Cada fila trae `quantity`, `minStock`, `lowStock` y `updatedAt`
  (`null` si el artículo no tiene movimientos).
- **Movimientos** (`GET /api/v1/stock/movements?itemId=&type=`): del más
  nuevo al más viejo; los registrados en el mismo instante, en orden inverso
  al de registro.

Permisos: `stock:view` para consultar existencias y movimientos (también el
cajero) y `stock:manage` para ingresos y ajustes.

Códigos de error: `stock_item_required`, `stock_item_not_found`,
`stock_supplier_not_found`, `stock_quantity_invalid`,
`stock_adjustment_quantity_invalid`, `stock_quantity_not_whole`,
`stock_balance_out_of_range`, `stock_unit_cost_invalid`,
`stock_entry_lines_required`, `stock_entry_too_many_lines`,
`stock_entry_item_repeated`, `stock_reference_too_long`,
`stock_reason_required`, `stock_reason_too_long`,
`stock_movement_type_invalid`, `stock_already_initialized`,
`insufficient_stock`.

## Ventas

Una venta (`Sale`) es un comprobante confirmado (no fiscal). **Nunca se edita
ni se borra.** El carrito vive en el frontend; el backend recibe la venta
completa y la registra o la rechaza entera.

**Alta** (`POST /api/v1/sales`, header `Idempotency-Key` obligatorio; la clave es única por actor):

| Campo del pedido | Regla |
| --- | --- |
| `customerId` | Opcional. Sin cliente, la venta se le hace a Consumidor Final. Tiene que existir (`sale_customer_not_found`). |
| `sellerId` | Opcional: una venta puede no tener vendedor. Si viene, tiene que existir (`sale_seller_not_found`). |
| `lines` | De 1 a 100 líneas, una por artículo (`sale_item_repeated` si se repite). |
| `lines[].quantity` | Milésimas, mayor a cero y hasta un millón de unidades; entera si el artículo se vende por unidad (`sale_quantity_not_whole`). |
| `lines[].discount` | Opcional, descuento de esa línea: `{ "type": "percent", "value": 1000 }` (centésimas de punto, de 0 a 100 %) o `{ "type": "amount", "value": 50000 }` (centavos, desde cero). `null` es sin descuento. |
| `cartDiscount` | Opcional, el **descuento general** de toda la venta, con la misma forma. Se aplica después de los descuentos de línea. |
| `payments` | Hasta 10 pagos: `method` (`cash`, `debit_card`, `credit_card`, `transfer`, `qr`), `amount` en centavos mayor a cero y `reference` opcional (cupón, número de transferencia), hasta 60 caracteres. |
| `expectedTotal` | El total que vio el cajero, en centavos. |

Reglas (`domain.NewSale`):

- **El backend recalcula todo** con el precio actual de cada artículo. Si el
  total no coincide con `expectedTotal`, falla con `sale_total_mismatch`
  (`409`): el cajero nunca cobra un importe distinto del que vio.
- **Importe de línea**: precio × cantidad ÷ 1000 redondeado al centavo. El
  redondeo es siempre *half away from zero*.
- **Descuento de línea** (`discountAmount`): un porcentaje del importe,
  redondeado al centavo, o un monto fijo, que no puede superar el importe
  (`sale_discount_exceeds_line`).
- **Descuento general** (`cartDiscountAmount`): un porcentaje (redondeado al
  centavo) o un monto fijo de lo que suman las líneas después de sus propios
  descuentos; un monto no puede superarlo (`sale_discount_exceeds_total`).
  Se **reparte entre las líneas** en proporción a lo que vale cada una: cada
  parte es su proporción redondeada hacia abajo y los centavos que sobran
  van, de a uno, a las fracciones más grandes (a igualdad, a la primera
  línea). Así las partes suman exacto el descuento y ninguna línea queda
  negativa. Cada línea guarda su parte en `cartDiscountAmount`.
- **Totales**: `lineTotal` = importe − descuento de línea − parte del
  general; `subtotal` suma los importes sin descuento, `discountTotal` todos
  los descuentos y `total` las líneas, así que el ticket siempre suma exacto
  y nunca es negativo. La venta guarda también los descuentos tal como se
  pidieron (`lines[].discount`, `cartDiscount`).
- **Copia de datos**: cada línea guarda código, nombre, unidad, alícuota y
  precio del artículo; la venta guarda nombre, documento y condición frente
  al IVA del cliente, y nombre y código del vendedor. Editarlos después no
  cambia ventas pasadas.
- **Pagos**: la suma tiene que cubrir el total (`sale_payment_insufficient`);
  los medios que no son efectivo no pueden superarlo
  (`sale_non_cash_payment_exceeds_total`); lo que el efectivo pague de más
  es el vuelto (`change`). Una venta de total cero no necesita pagos.
- **Stock**: cada línea registra un movimiento `sale` con la referencia
  `Venta <número>`. Con `STOCK_ALLOW_NEGATIVE=false` (por defecto), una línea
  que el stock no cubre falla con `insufficient_stock` (`409`) nombrando el
  artículo; con `true`, la existencia queda negativa.
- **Atomicidad**: dentro de un `TxManager.WithinTx` se busca la clave de
  idempotencia, se leen cliente, vendedor y artículos, el dominio calcula la
  venta, el repositorio la guarda (le asigna el ID y el `number`
  correlativo) y se registran los movimientos. Si algo falla, no queda nada.
- **Idempotencia**: la venta guarda su clave y una huella (`SHA-256`) del
  pedido. La misma clave con el mismo pedido devuelve la venta original sin
  registrar nada; con otro pedido, `idempotency_key_conflict` (`409`).

**Listado** (`GET /api/v1/sales?from=&to=&customerId=&sellerId=&status=&page=&limit=`):
de la más nueva a la más vieja (por `number`). `from` y `to` son instantes
RFC 3339 (`2026-11-12T00:00:00-03:00`): el frontend sabe en qué zona horaria
está el comercio y arma el día; `from` se incluye y `to` no. Una fecha mal
escrita es `sale_date_invalid` y un `from` que no es anterior a `to`,
`sale_date_range_invalid`. `status` es `completed` o `voided`.

**Detalle** (`GET /api/v1/sales/{id}`): la venta completa, con sus líneas,
pagos y, si está anulada, `voidedAt`, `voidReason` y `voidedBy` (en una venta
completada son `null`).

**Anulación** (`POST /api/v1/sales/{id}/void`, cuerpo `{ "reason": "…" }`):
solo una venta `completed` (`sale_already_voided`, `409`, si ya está
anulada) y con motivo obligatorio de hasta 200 caracteres. Pasa a `voided`
y, en el mismo `WithinTx`, cada línea registra un movimiento `sale_void` que
devuelve su cantidad al stock (la referencia es la misma, `Venta <número>`).
La venta y sus movimientos originales no se tocan: la anulación los
compensa.

Permisos: `sales:view` para listar y ver el detalle y `sales:create` para
vender (todos los roles); `sales:void` para anular (solo `admin` y
`manager`).

Códigos de error: `sale_not_found`, `idempotency_key_required`,
`idempotency_key_invalid`, `idempotency_key_conflict`, `sale_lines_required`,
`sale_too_many_lines`, `sale_item_required`, `sale_item_repeated`,
`sale_item_not_found`, `sale_item_discontinued`, `sale_quantity_invalid`, `sale_quantity_not_whole`,
`sale_discount_invalid`, `sale_discount_exceeds_line`,
`sale_discount_exceeds_total`, `sale_customer_not_found`, `sale_seller_not_found`,
`sale_total_mismatch`, `sale_too_many_payments`, `payment_method_invalid`,
`payment_amount_invalid`, `payment_reference_too_long`,
`sale_payment_insufficient`, `sale_non_cash_payment_exceeds_total`,
`amount_out_of_range`, `insufficient_stock`, `sale_already_voided`,
`sale_void_reason_required`, `sale_void_reason_too_long`,
`sale_status_invalid`, `sale_date_invalid`, `sale_date_range_invalid`.
Del resumen: `summary_period_invalid`.

## Resumen del día

`GET /api/v1/dashboard/summary?from=&to=` (permiso `dashboard:view`, todos los
roles) suma las ventas de un período y lista el stock bajo. `from` y `to` son
instantes RFC 3339 **obligatorios**, `from` incluido y `to` excluido: el cliente
manda el inicio y el fin de *su* día, así el resumen sigue su zona horaria y
el backend no tiene que configurar una. Un período faltante, al revés o de más
de 31 días falla con `summary_period_invalid` (`400`); una fecha mal escrita,
con `sale_date_invalid`.

Respuesta (importes en centavos, cantidades en milésimas):

| Campo | Qué es |
| --- | --- |
| `sales.count` y `sales.total` | Cantidad y total de las ventas **completadas** del período; las anuladas no cuentan. |
| `sales.byMethod` | Lo cobrado por cada medio de pago, en orden fijo (efectivo, débito, crédito, transferencia, QR); solo aparecen los que cobraron algo. **El efectivo va neto del vuelto**, así los medios suman `sales.total`. |
| `lowStock.total` | Cuántos artículos llegaron a su stock mínimo (`IsLowStock`: mínimo mayor a 0 y existencia menor o igual). |
| `lowStock.items` | Los primeros 20 por nombre, con la misma forma que `GET /api/v1/stock`. |

La cuenta es del dominio (`domain.NewSalesSummary`); el caso de uso
(`usecase/dashboard`) recorre las páginas de `SaleRepository.List` con
`status=completed` y pide a `StockRepository.ListLevels` el stock bajo, sin
agregar métodos a los repositorios.

## Actor, roles y permisos

Toda operación la ejecuta un `Actor` (`ID`, `Name`, `Role`). Roles:
`admin`, `manager` (encargado) y `cashier` (cajero).

Los casos de uso piden un **permiso**, nunca un rol:
`Actor.Authorize(domain.PermissionManageCatalog)` devuelve `forbidden` si el
rol del actor no lo otorga o si el actor no tiene `ID`. Qué rol tiene qué
permiso se decide en un único lugar, `domain/permission.go`:

| Permiso | `admin` | `manager` | `cashier` |
| --- | :---: | :---: | :---: |
| `catalog:view` — ver categorías y artículos | ✅ | ✅ | ✅ |
| `catalog:manage` — crear, editar y eliminar categorías y artículos | ✅ | ✅ | — |
| `customers:view` — ver clientes | ✅ | ✅ | ✅ |
| `customers:create` — dar de alta clientes | ✅ | ✅ | ✅ |
| `customers:manage` — editar y eliminar clientes | ✅ | ✅ | — |
| `suppliers:view` — ver proveedores | ✅ | ✅ | — |
| `suppliers:manage` — crear, editar y eliminar proveedores | ✅ | ✅ | — |
| `sellers:view` — ver vendedores | ✅ | ✅ | ✅ |
| `sellers:manage` — crear, editar y eliminar vendedores | ✅ | ✅ | — |
| `stock:view` — ver existencias y movimientos | ✅ | ✅ | ✅ |
| `stock:manage` — registrar ingresos de mercadería y ajustes | ✅ | ✅ | — |
| `sales:view` — ver ventas | ✅ | ✅ | ✅ |
| `sales:create` — vender | ✅ | ✅ | ✅ |
| `sales:void` — anular ventas | ✅ | ✅ | — |
| `dashboard:view` — ver el resumen del día | ✅ | ✅ | ✅ |

Cada feature nueva suma sus permisos a esa matriz (y a su test,
`TestRolePermissions`), no a sus casos de uso.

Mientras no haya autenticación, el middleware `DevActor` asigna a cada
request un actor `admin` fijo.

## Errores de negocio

`domain.Error{Kind, Code, Message, Cause}`. `Kind` clasifica (`invalid`,
`forbidden`, `conflict`, `not_found`, `unavailable`), `Code` es estable para
el cliente y `Message` está en español para el usuario. `FromStorage(err)`
deja pasar un error de negocio y convierte cualquier otro en
`storage_unavailable` con el original como `Cause`.
