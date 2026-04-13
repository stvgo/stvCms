# Ideas de Implementación para el Backend

## Prioridad: Críticas

### 1. Arreglar el Stub de `GetPostByFilter`
- Implementar la lógica de filtrado en la capa de servicio
- Soporte para filtrar por: tipo de bloque de contenido, rango de fechas, ID de usuario, búsqueda por palabra clave en título/contenido
- Agregar soporte de paginación (límite/desplazamiento)

### 2. Arreglar `LoginAndRegisterService`
- Crear una implementación propia de `LoginAndRegisterService` en lugar de retornar `postService{}`
- Definir métodos en la interfaz `ILoginAndRegisterService` (ej. `GetUserByEmail`, `CreateUserFromOAuth`)
- Implementar la persistencia del usuario en la base de datos al completar el callback de OAuth

---

## Prioridad: Alta

### 3. Sistema de Gestión de Usuarios
- Crear el modelo `User` (ID, email, nombre, avatar, rol, created_at, updated_at)
- Construir capas de repositorio, servicio y handler para usuarios
- Endpoints a implementar:
  - `GET /user/profile` - Obtener perfil del usuario actual
  - `PUT /user/profile` - Actualizar perfil
  - `GET /user/posts` - Obtener posts del usuario actual
  - `GET /admin/users` - Listar todos los usuarios (solo admin)
  - `PUT /admin/users/:id/role` - Cambiar rol de usuario (solo admin)

### 4. Middleware de Autenticación
- Implementar generación de tokens JWT al completar el login por OAuth
- Crear middlewares para Echo:
  - Validación de JWT (`Protected`)
  - Control de acceso basado en roles (`RequireRole("admin")`)
- Agregar mecanismo de refresh token
- Almacenar lista negra de tokens en Redis para soporte de logout

### 5. Validación y Sanitización de Entradas
- Agregar tags de validación a todos los DTOs de request (`validate:"required,min=3,max=200"`)
- Implementar middleware de Echo para formatear errores de validación
- Sanitizar contenido HTML en bloques de contenido para prevenir XSS
- Validar orden y tipos de bloques de contenido

### 6. Estandarización de Manejo de Errores
- Crear formato unificado de respuesta de error (`ErrorResponse{ Code, Message, Details }`)
- Implementar tipos de error personalizados (NotFound, ValidationError, Unauthorized, Forbidden)
- Agregar middleware global de manejo de errores
- Loguear errores con contexto (request ID, user ID, timestamp)

---

## Prioridad: Media

### 7. Funcionalidades Avanzadas de Posts

#### 7.1. Estado y Borradores
- Agregar campo `Status` al modelo Post (`draft`, `published`, `archived`)
- Filtrar posts por estado
- Endpoint de autoguardado de borradores

#### 7.2. Tags y Categorías
- Crear modelos `Tag` y `Category` con relaciones many-to-many
- Endpoints para gestionar tags/categorías
- Filtrar posts por tag o categoría
- Endpoint de autocompletado para tags

#### 7.3. Generación de Slugs
- Agregar campo `Slug` al modelo Post (versión URL-friendly del título)
- Generar slug automáticamente al crear
- Asegurar unicidad del slug
- Acceder posts por slug: `GET /post/:slug`

#### 7.4. Búsqueda de Texto Completo
- Implementar búsqueda de texto completo de PostgreSQL en título y bloques de contenido
- Endpoint: `GET /post/search?q=query&page=1&limit=10`
- Retornar resultados ordenados por relevancia con coincidencias resaltadas

### 8. Sistema de Comentarios
- Crear modelo `Comment` (ID, post_id, user_id, content, parent_id para comentarios anidados, created_at)
- Endpoints:
  - `POST /post/:id/comments` - Agregar comentario
  - `GET /post/:id/comments` - Obtener todos los comentarios (estructura anidada/árbol)
  - `PUT /comment/:id` - Editar comentario
  - `DELETE /comment/:id` - Eliminar comentario
- Rate limiting por usuario/IP vía Redis

### 9. Gestión de Medios
- Crear modelo `Media` (ID, filename, original_name, size, mime_type, uploaded_by, created_at)
- Endpoints:
  - `POST /media/upload` - Subir archivo
  - `GET /media` - Listar todos los medios (paginado)
  - `GET /media/:id` - Obtener detalles de un medio
  - `DELETE /media/:id` - Eliminar medio
  - `GET /media/:filename` - Servir archivo
- Soporte para múltiples tipos de archivo (imágenes, documentos, videos)
- Generar thumbnails para imágenes
- Integrar con almacenamiento en la nube (AWS S3, Cloudinary)

### 10. Rate Limiting
- Implementar limitador de tasa usando Redis
- Aplicar a:
  - Endpoints de autenticación (prevenir fuerza bruta)
  - Creación de comentarios
  - Subida de imágenes
  - Endpoint de búsqueda
- Límites configurables por tipo de endpoint

### 11. Expansión de Estrategia de Caché
- Cachear posts individuales por ID: `post:{id}`
- Cachear perfiles de usuario: `user:{id}`
- Cachear árboles de comentarios: `comments:{post_id}`
- Implementar calentamiento de caché al iniciar
- Agregar tags de caché para invalidación masiva

### 12. Analíticas y Métricas
- Rastrear vistas de posts (contador por post, visitantes únicos)
- Endpoint: `GET /post/:id/analytics` (solo admin)
- Almacenar eventos de vista en Redis (ligero, procesado por lotes a BD)
- Endpoint de posts populares: `GET /post/popular`
- Agregar middleware de request ID para trazabilidad

---

## Prioridad: Baja

### 13. Permisos de Contenido Basados en Roles
- Permitir que diferentes roles tengan diferentes capacidades:
  - `admin`: acceso total
  - `editor`: editar/eliminar cualquier post
  - `author`: crear/editar propios posts
  - `viewer`: solo lectura
- Implementar verificaciones de permisos en la capa de servicio

### 14. Sistema de Webhooks
- Permitir registrar URLs de webhook para eventos:
  - `post.created`, `post.updated`, `post.deleted`
  - `comment.created`
  - `user.registered`
- Almacenar configuraciones de webhook en BD
- Reintentar entregas fallidas con backoff exponencial
- Endpoint de log de entregas de webhook

### 15. Exportar/Importar
- Exportar posts como JSON o Markdown
- Importar posts desde JSON/Markdown
- Endpoint de operaciones masivas

### 16. Posts Programados
- Agregar campo `PublishAt` al modelo Post
- Worker en segundo plano para publicar posts programados
- Filtro: `GET /post/scheduled` (solo admin)
- Cron job o sorted sets de Redis para programación

### 17. Versionado de API
- Versionar rutas: `/api/v1/post/...`
- Mantener compatibilidad hacia atrás
- Documentar diferencias entre versiones

### 18. Endpoints de Salud y Monitoreo
- `GET /health` - Verificación básica de salud (DB + Redis ping)
- `GET /health/ready` - Verificación de readiness (migraciones listas, conexiones activas)
- `GET /metrics` - Endpoint compatible con Prometheus
- Exponer: conexiones activas, tasa de acierto de caché, tasa de errores, tiempos de respuesta

### 19. Sistema de Notificaciones
- Notificaciones in-app para usuarios
- Notificaciones por email (vía SendGrid, AWS SES, etc.)
- Preferencias de notificación por usuario
- Cola de notificaciones vía Redis

### 20. Infraestructura de Testing
- Tests unitarios para servicios (mock de repositorios)
- Tests de integración para handlers (capa HTTP)
- Tests de repositorio contra base de datos de prueba (docker-compose test DB)
- Helpers/fixtures de prueba
- Apuntar a >70% de cobertura de código

### 21. Inyección de Dependencias
- Refactorizar servicios para aceptar dependencias por constructor
- Elimir acoplamiento fuerte (ej. `PostService` llamando directamente a `NewPostGormRepository()`)
- Usar contenedor DI o librería wire para arquitectura más limpia

### 22. Logging y Observabilidad
- Logging estructurado (formato JSON) con niveles (DEBUG, INFO, WARN, ERROR)
- Middleware de logging de requests (método, path, status, duración, IP)
- IDs de correlación/request a lo largo del ciclo de vida del request
- Opcional: integración de OpenTelemetry para trazabilidad distribuida

### 23. Optimizaciones de Base de Datos
- Agregar índices de BD en campos consultados frecuentemente (user_id, status, created_at)
- Implementar paginación a nivel de BD para `GetPosts`
- Usar paginación basada en cursor para datasets grandes
- Sistema de migraciones de BD (en lugar de AutoMigrate)

### 24. Documentación
- Generar automáticamente spec OpenAPI/Swagger
- Integrar Swagger UI: `GET /docs`
- Documentar todos los esquemas de request/response
- Agregar ejemplos de requests en la colección de Insomnia

---

## Orden Sugerido de Implementación

1. **Fase 1** (Fundación): Arreglar stubs, agregar gestión de usuarios, middleware de autenticación
2. **Fase 2** (Funcionalidades Core): Validación, manejo de errores, estado de posts, tags/categorías, búsqueda
3. **Fase 3** (Interacción): Comentarios, gestión de medios, rate limiting
4. **Fase 4** (Pulido): Expansión de caché, analíticas, endpoints de salud, testing
5. **Fase 5** (Avanzado): Webhooks, notificaciones, posts programados, observabilidad
