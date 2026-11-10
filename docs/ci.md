# Integración continua

`.github/workflows/ci.yml` corre en cada pull request y en cada push a `main`
o `develop`.

| Job | Qué hace |
| --- | --- |
| `build` | Typecheck, lint y build del frontend; `gofmt`, `go vet` y build del backend. |
| `test` | `pnpm test` (frontend y backend). |
| `coverage` | `pnpm test:coverage`, aplicando los umbrales de [testing.md](testing.md). |
| `dependency-audit` | `pnpm audit --audit-level=high` y `govulncheck` sobre todo el árbol actual. |
| `dependency-review` | Solo en PRs: revisa las dependencias nuevas o cambiadas del diff. |
| `compose-config` | `docker compose config`: valida `compose.yaml` sin levantar nada. |
| `main-source-branch` | Solo en PRs contra `main`: exige que vengan de `develop` o `hotfix/*`. |

La versión de Go sale de la directiva `go` de `apps/backend/go.mod`.
Mantenerla en el último parche evita que `govulncheck` reporte
vulnerabilidades de la librería estándar ya corregidas.

Cuando llegue PostgreSQL, `test` y `coverage` levantan un servicio de base y
aplican las migraciones antes de la suite.

## Ramas

- `main`: lo publicado. Solo recibe PRs desde `develop` o `hotfix/*`.
- `develop`: integración. Toda feature sale de acá y vuelve por PR.
- Ramas de trabajo: `feat/…`, `fix/…`, `chore/…`, `docs/…`.
- Un `hotfix/*` mergeado a `main` se mergea también a `develop`.

### Protección de ramas

Configurada en GitHub (**Settings › Branches**), no en el repositorio.
`main` y `develop` exigen:

- cambios solo por pull request, sin aprobación obligatoria por ahora (con un
  único colaborador bloquearía cada merge: GitHub no deja aprobar el PR
  propio). Sumarla cuando haya más de una persona;
- los checks `build`, `test`, `coverage`, `dependency-audit`,
  `dependency-review` y `compose-config` en verde, más `main-source-branch`
  en `main`;
- la rama del PR actualizada contra la base;
- conversaciones de revisión resueltas;
- sin force push ni borrado de la rama.

La regla no aplica a admins (`enforce_admins` desactivado): el dueño del repo
puede mergear saltándola en una emergencia, y queda registrado en el PR.

`dependency-review` necesita el *Dependency graph* activado
(**Settings › Code security**); está activo junto con las alertas de
Dependabot.
