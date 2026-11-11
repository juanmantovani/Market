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

Eliminar borra la categoría y libera su nombre; no se puede deshacer. Cuando
existan artículos, eliminar una categoría que tenga artículos va a fallar con
`category_in_use` (`409`).

Códigos de error: `category_not_found`, `category_name_required`,
`category_name_too_long`, `category_description_too_long`,
`category_name_in_use`, `category_update_empty`.

## Actor, roles y permisos

Toda operación la ejecuta un `Actor` (`ID`, `Name`, `Role`). Roles:
`admin`, `manager` (encargado) y `cashier` (cajero).

Los casos de uso piden un **permiso**, nunca un rol:
`Actor.Authorize(domain.PermissionManageCatalog)` devuelve `forbidden` si el
rol del actor no lo otorga o si el actor no tiene `ID`. Qué rol tiene qué
permiso se decide en un único lugar, `domain/permission.go`:

| Permiso | `admin` | `manager` | `cashier` |
| --- | :---: | :---: | :---: |
| `catalog:view` — ver categorías (y luego artículos) | ✅ | ✅ | ✅ |
| `catalog:manage` — crear, editar y eliminar | ✅ | ✅ | — |

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
