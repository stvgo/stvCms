-- Seed data for stvCms (local dev)
-- GORM AutoMigrate crea las tablas; este script solo inserta datos mock.
-- Se ejecuta al primer arranque del contenedor (cuando el volumen está vacío).

-- Esperar a que las tablas existan (AutoMigrate las crea al iniciar la app).
-- Si se necesita correr antes que la app, descomentar el bloque CREATE TABLE.

-- =============================================================================
-- POSTS
-- =============================================================================
INSERT INTO posts (created_at, updated_at, deleted_at, title, user_id) VALUES
  (NOW(), NOW(), NULL, 'Introducción a Go', 'user-mock-001'),
  (NOW(), NOW(), NULL, 'Docker para desarrolladores', 'user-mock-001'),
  (NOW(), NOW(), NULL, 'PostgreSQL: tips y trucos', 'user-mock-002'),
  (NOW(), NOW(), NULL, 'Redis como caché en producción', 'user-mock-002'),
  (NOW(), NOW(), NULL, 'Construyendo una API REST con Echo', 'user-mock-003');

-- =============================================================================
-- CONTENT BLOCKS
-- =============================================================================

-- Post 1: Introducción a Go
INSERT INTO content_blocks (created_at, updated_at, deleted_at, type, "order", content, language, post_id) VALUES
  (NOW(), NOW(), NULL, 'text',  1, 'Go es un lenguaje de programación compilado, tipado estáticamente, diseñado en Google. Es conocido por su simplicidad y eficiencia.', '', 1),
  (NOW(), NOW(), NULL, 'code',  2, 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hola, mundo!")\n}', 'go', 1),
  (NOW(), NOW(), NULL, 'text',  3, 'Go sobresale en aplicaciones de red y sistemas concurrentes gracias a sus goroutines y canales.', '', 1),
  (NOW(), NOW(), NULL, 'url',   4, 'https://go.dev/doc/', '', 1);

-- Post 2: Docker para desarrolladores
INSERT INTO content_blocks (created_at, updated_at, deleted_at, type, "order", content, language, post_id) VALUES
  (NOW(), NOW(), NULL, 'text',  1, 'Docker permite empaquetar aplicaciones con todas sus dependencias en contenedores portables.', '', 2),
  (NOW(), NOW(), NULL, 'code',  2, 'FROM golang:1.22-alpine\nWORKDIR /app\nCOPY . .\nRUN go build -o server ./cmd/server\nCMD ["./server"]', 'dockerfile', 2),
  (NOW(), NOW(), NULL, 'text',  3, 'Con docker-compose podemos orquestar múltiples servicios como la app, la base de datos y el caché.', '', 2),
  (NOW(), NOW(), NULL, 'code',  4, 'docker-compose up -d', 'bash', 2);

-- Post 3: PostgreSQL tips
INSERT INTO content_blocks (created_at, updated_at, deleted_at, type, "order", content, language, post_id) VALUES
  (NOW(), NOW(), NULL, 'text',  1, 'PostgreSQL es uno de los motores de base de datos relacionales más potentes y confiables del mundo open source.', '', 3),
  (NOW(), NOW(), NULL, 'code',  2, 'EXPLAIN ANALYZE SELECT * FROM posts WHERE user_id = ''user-mock-001'';', 'sql', 3),
  (NOW(), NOW(), NULL, 'text',  3, 'Usar índices en columnas de filtrado frecuente mejora drásticamente el rendimiento de las consultas.', '', 3),
  (NOW(), NOW(), NULL, 'code',  4, 'CREATE INDEX idx_posts_user_id ON posts(user_id);', 'sql', 3);

-- Post 4: Redis como caché
INSERT INTO content_blocks (created_at, updated_at, deleted_at, type, "order", content, language, post_id) VALUES
  (NOW(), NOW(), NULL, 'text',  1, 'Redis es una base de datos en memoria de alto rendimiento, ideal para cachear resultados costosos.', '', 4),
  (NOW(), NOW(), NULL, 'code',  2, 'client.Set(ctx, "posts:all", jsonData, 24*time.Hour)', 'go', 4),
  (NOW(), NOW(), NULL, 'text',  3, 'La clave "posts:all" almacena todos los posts durante 24 horas, reduciendo la carga en PostgreSQL.', '', 4);

-- Post 5: API REST con Echo
INSERT INTO content_blocks (created_at, updated_at, deleted_at, type, "order", content, language, post_id) VALUES
  (NOW(), NOW(), NULL, 'text',  1, 'Echo es un framework web de Go de alto rendimiento, minimalista y extensible.', '', 5),
  (NOW(), NOW(), NULL, 'code',  2, 'e := echo.New()\ne.Use(middleware.Logger())\ne.Use(middleware.Recover())\n\ne.GET("/post/getAll", postHandler.GetPosts)\ne.POST("/post/create", postHandler.CreatePost)\n\ne.Start(":8080")', 'go', 5),
  (NOW(), NOW(), NULL, 'text',  3, 'Con pocas líneas tenemos un servidor con logging, recuperación de panics y rutas definidas.', '', 5),
  (NOW(), NOW(), NULL, 'url',   4, 'https://echo.labstack.com/', '', 5);
