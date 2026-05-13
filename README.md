# Omnibus Backend

Backend para un sistema de gestión de terminales de ómnibus. Permite administrar terminales, andenes, ciudades, usuarios y notificaciones en tiempo real a pasajeros.

## Punto de entrada

```
cmd/api/main.go
```

El arranque sigue este flujo: carga de config → conexión a PostgreSQL → inicialización del contenedor de dependencias (`app.New`) → registro de rutas → inicio del servidor.

## Tecnologías y librerías principales

| Librería | Uso |
|---|---|
| `github.com/labstack/echo/v4` | Framework HTTP |
| `gorm.io/gorm` + `gorm.io/driver/postgres` | ORM y driver de PostgreSQL |
| `github.com/golang-jwt/jwt/v5` | Autenticación JWT |
| `github.com/philippseith/signalr` | Comunicación en tiempo real (SignalR) |
| `golang.org/x/crypto` | Hash de contraseñas (bcrypt) |
| `github.com/google/uuid` | Generación de UUIDs |
| `github.com/joho/godotenv` | Carga de variables de entorno desde `.env` |

## Cómo arrancar el proyecto

### Requisitos previos

- Go 1.25+
- PostgreSQL corriendo localmente (o configurable vía `DATABASE_URL`)
- `psql` en el PATH (para migraciones)
- `make` instalado

### Variables de entorno

Crear un archivo `.env` en la raíz (o exportar las variables al entorno):

```env
DATABASE_URL=postgres://postgres:1234@localhost:5432/omnibus-terminal
JWT_SECRET=tu-secreto-jwt
PASSWORD_RESET_JWT_SECRET=tu-secreto-reset

LISTEN_ADDR=:4989

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=correo@gmail.com
SMTP_PASSWORD=app-password
SMTP_FROM=correo@gmail.com

EXTERNAL_TERMINAL_UPSTREAM_URL=http://localhost:4990
REALTIME_URL=http://localhost:4988/realtime
REALTIME_API_KEY=clave-realtime
CAMERA_NOTIFICATION_API_KEY=clave-camara

FRONT_END_BASE_LINK=http://localhost:4200/
```

### Correr el servidor

```bash
go run ./cmd/api
```

## Migraciones

Las migraciones se gestionan con `psql` plano. El archivo `migrations.mk` define los targets, incluido desde el `Makefile` principal.

```bash
# Aplicar todas las migraciones en orden
make migrate-up

# Revertir todas las migraciones en orden inverso
make migrate-down
```

Las migraciones están versionadas en `migrations/` con el formato `NNN_nombre.up.sql` / `NNN_nombre.down.sql`:

| Versión | Descripción |
|---|---|
| `001` | Tablas base: `city`, `bus_terminal`, `platform` |
| `002` | Auth y roles: `rol`, `users`, `user_refresh_tokens`, `user_terminal` + seed de roles y permisos |
| `003` | CASCADE en FK de `bus_terminal → city` |
| `004` | Campo `external_terminal_id` en `bus_terminal` |
| `005` | Eliminación del campo `dni` en `users` |
| `006` | Tabla `awaited_trip` |
| `007` | Eliminación de `permissions` y `rol_permissions` (RBAC simplificado a roles) |

## Arquitectura

El proyecto sigue una arquitectura en capas estricta:

```
Handler → Service → Repository → Database (PostgreSQL/GORM)
```

```
cmd/api/            Entrypoint
internal/
├── app/            Contenedor de dependencias (DI manual)
├── config/         Carga de configuración (.env + defaults)
├── database/       Conexión a PostgreSQL
├── errors/         Errores de dominio reutilizables
├── handler/        Controladores HTTP (entrada/salida HTTP)
├── mail/           Servicio de email + templates HTML
├── middleware/     CORS, logging, auth JWT, roles, API key
├── models/         Structs de dominio y DTOs
├── realtime/       Cliente SignalR (grupos y métodos del hub)
├── repository/     Acceso a datos vía GORM
├── roles/          Constantes de roles (user, admin, super_admin)
├── router/         Registro de rutas por perfil
├── service/        Lógica de negocio
└── validators/     Validación de input por dominio
pkg/
└── realtime/       Wrapper del cliente SignalR
migrations/         Archivos SQL versionados
```

**Patrones aplicados:**
- Dependency Injection manual en `app.New()`
- Repository Pattern para abstraer el acceso a datos
- Middleware chain: CORS → Logging → Auth JWT → Role check → Handler
- JWT de acceso (corta duración) + Refresh Token en cookie HTTP-only

**Sistema de roles:**

| Rol | Descripción |
|---|---|
| `user` | Pasajero registrado |
| `admin` | Administrador de una terminal |
| `super_admin` | Administrador global del sistema |

## Endpoints

### Públicos / Auth

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/bus_tickets/:ticket_string` | Consulta de pasaje por string de ticket |
| `POST` | `/notify_passengers` | Notifica llegada de bus a pasajeros (API key) |
| `POST` | `/notify_camera_error` | Notifica error de cámara (API key) |
| `GET` | `/api/notifications` | Lista notificaciones del usuario (JWT) |
| `POST` | `/api/auth/register` | Registro de nuevo usuario |
| `POST` | `/api/auth/login` | Login, devuelve access token |
| `POST` | `/api/auth/refresh` | Refresca el access token |
| `POST` | `/api/auth/logout` | Cierra sesión |
| `POST` | `/api/auth/forgot-password` | Solicita email de recuperación |
| `POST` | `/api/auth/validate-recovery-token` | Valida token de recuperación |
| `POST` | `/api/auth/reset-password` | Resetea la contraseña |

### Usuario autenticado (`user`, `admin`, `super_admin`)

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/users/me` | Obtiene perfil propio |
| `PUT` | `/api/users/me` | Actualiza perfil propio |
| `DELETE` | `/api/users/me` | Elimina cuenta propia |
| `GET` | `/api/users/terminals` | Lista terminales asignadas al usuario |
| `POST` | `/api/buses/join` | Se une a un bus (registra un awaited trip) |

### Admin (`admin`, `super_admin`)

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/admin/cities` | Lista ciudades |
| `GET` | `/api/admin/cities/:postal_code` | Obtiene ciudad por código postal |
| `POST` | `/api/admin/cities` | Crea ciudad |
| `PUT` | `/api/admin/cities/:postal_code` | Actualiza ciudad |
| `DELETE` | `/api/admin/cities/:postal_code` | Elimina ciudad |
| `GET` | `/api/admin/platforms` | Lista andenes |
| `GET` | `/api/admin/platforms/:code` | Obtiene andén por código |
| `POST` | `/api/admin/platforms` | Crea andén |
| `PUT` | `/api/admin/platforms/:code` | Actualiza andén |
| `DELETE` | `/api/admin/platforms/:code` | Elimina andén |
| `GET` | `/api/admin/users/by-email` | Busca usuario por email |
| `POST` | `/api/admin/users/promote` | Promueve usuario a admin |
| `POST` | `/api/admin/users/demote` | Degrada admin a usuario |
| `GET` | `/api/admin/notification-types` | Lista tipos de notificación disponibles |
| `POST` | `/api/admin/notifications` | Envía notificación a pasajeros |
| `DELETE` | `/api/admin/notifications` | Elimina una notificación |
| `POST` | `/api/admin/notify-bus-delay` | Notifica demora de bus |

### Super Admin (`super_admin`)

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/super/terminals` | Lista todas las terminales |
| `GET` | `/api/super/terminals/:uuid` | Obtiene terminal por UUID |
| `POST` | `/api/super/terminals` | Crea terminal |
| `PUT` | `/api/super/terminals/:uuid` | Actualiza terminal |
| `DELETE` | `/api/super/terminals/:uuid` | Elimina terminal |
| `POST` | `/api/super/users/promote-super` | Promueve admin a super_admin |
| `POST` | `/api/super/users/demote-super` | Degrada super_admin a admin |
