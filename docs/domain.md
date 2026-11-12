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
| Anulación | `Void` |

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

| Campo | Regla |
| --- | --- |
| `code` | Código interno (SKU). Obligatorio, hasta 30 caracteres, único sin distinguir mayúsculas ni acentos. |
| `barcode` | Opcional. Sin espacios, hasta 32 caracteres, único si se carga (dos artículos pueden no tener). |
| `name` | Obligatorio, hasta 100 caracteres. |
| `description` | Opcional, hasta 500 caracteres. |
| `categoryId` | Obligatoria y tiene que existir (`item_category_not_found` si no). |
| `unit` | `unit` (por unidad), `kg`, `l` o `m`. |
| `salePrice` | Centavos, mayor a 0 y hasta `MaxItemAmount` (mil millones de pesos). **Incluye IVA.** |
| `cost` | Centavos, opcional, entre 0 y `MaxItemAmount`. |
| `taxRate` | Alícuota de IVA en centésimas de punto: `0`, `1050`, `2100` o `2700`. Informativa hasta facturar. |
| `tracksStock` | Si el artículo controla stock (un servicio no). |
| `minStock` | Milésimas de la unidad, entre 0 y un millón de unidades; entero si se vende por unidad. Umbral de la alerta de stock bajo (Fase 3). |

Editar reemplaza todos los campos (`PUT`): el formulario siempre manda el
artículo completo, y así un opcional enviado como `null` se borra.

Permisos: los de catálogo (`catalog:view` para listar y consultar,
`catalog:manage` para crear, editar y eliminar).

Códigos de error: `item_not_found`, `item_code_required`,
`item_code_too_long`, `item_code_in_use`, `item_barcode_invalid`,
`item_barcode_in_use`, `item_name_required`, `item_name_too_long`,
`item_description_too_long`, `item_category_required`,
`item_category_not_found`, `item_unit_invalid`, `item_price_invalid`,
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

El cliente de sistema **Consumidor Final** que propone el plan llega con
ventas (Fase 4), que es quien lo usa.

Códigos de error: `customer_not_found`, `customer_name_required`,
`customer_name_too_long`, `customer_document_type_invalid`,
`customer_document_invalid`, `customer_document_in_use`,
`customer_tax_condition_invalid`, `customer_tax_condition_requires_cuit`,
`customer_email_invalid`, `customer_phone_too_long`,
`customer_address_too_long`, `customer_notes_too_long`.

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

El proveedor preferido de un artículo (`preferredSupplierId` en el plan)
todavía no existe; cuando llegue, eliminar un proveedor que tenga artículos
o ingresos fallará con `supplier_in_use`.

Códigos de error: `supplier_not_found`, `supplier_business_name_required`,
`supplier_business_name_too_long`, `supplier_cuit_invalid`,
`supplier_cuit_in_use`, `supplier_contact_name_too_long`,
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
eliminar. Cuando existan ventas (Fase 4), eliminar un vendedor que tenga
ventas fallará con `seller_in_use`.

Códigos de error: `seller_not_found`, `seller_name_required`,
`seller_name_too_long`, `seller_code_required`, `seller_code_invalid`,
`seller_code_in_use`, `seller_commission_invalid`.

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
