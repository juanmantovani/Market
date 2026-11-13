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

## Probar la demo

Al iniciar, el backend carga datos de ejemplo (`SEED_DEMO_DATA=true`): 8
categorías, 15 artículos con stock, 5 clientes, 4 proveedores, 3 vendedores y
5 ventas del día, una con descuentos, una con pagos mezclados y una anulada.
Con `SEED_DEMO_DATA=false` arranca vacío.

Un recorrido de dos minutos, en <http://localhost:5173>:

1. **Vender**: en el buscador escribir `7790000000059` y Enter (así actúa un
   lector de códigos de barras) para agregar una gaseosa. Con `Pan francés`
   pide el peso en kilos.
2. **Cobrar** con `F9`: dividir el pago entre efectivo y tarjeta, ver el
   vuelto y, al confirmar, **Imprimir ticket**.
3. **Ventas**: abrir una venta para ver el detalle y anularla; el stock se
   repone.
4. **Artículos**: la yerba, la leche y el queso cremoso aparecen con stock
   bajo; desde cada artículo se ajusta el stock, se ve su historial y se
   ingresa mercadería.
5. **Resumen**: ventas del día, total por medio de pago y artículos con stock
   bajo.

Con el teclado: `F2` busca, `F4` elige el vendedor, `F6` el cliente, `F9`
cobra y `?` muestra todos los atajos.

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
