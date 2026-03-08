# Architecture Decision Records

## ADR-001: Go as primary language

**Status:** Accepted

**Context:** Need a language with strong typing, first-class gRPC support, and minimal runtime for containerized deployment.

**Decision:** Go 1.24 with standard library where possible.

**Consequences:** Fast compilation, easy cross-compilation, but no generics-heavy DDD patterns (compared to Java/Kotlin).

---

## ADR-002: Single bounded context Identity

**Status:** Accepted

**Context:** Auth-модуль покрывает регистрацию, логин и восстановление пароля. Можно выделить `IdentityAccess` и `CredentialRecovery` как два bounded context.

**Decision:** Один bounded context `Identity`. Восстановление пароля — поддомен внутри него, а не отдельный контекст.

**Обоснование:**
- Ubiquitous Language: `User` означает одно и то же в регистрации, логине и recovery — владелец учётных данных.
- Recovery мутирует `User.PasswordHash` — тот же агрегат. Разделение потребует дублирования или shared kernel.
- Общая политика паролей: изменение правил затрагивает оба флоу одновременно.

**Когда стоит разделить:** если recovery вырастет в самостоятельный email-verification сервис с собственной моделью (`Recipient`, `VerificationAttempt`) и асинхронным взаимодействием через события.

**Consequences:** Простая модель, нет искусственных границ. Путь к разделению понятен и документирован.

---

## ADR-003: Pure gRPC without HTTP gateway

**Status:** Accepted

**Context:** Challenge requires gRPC or GraphQL. No frontend in scope.

**Decision:** Single gRPC service with reflection enabled for grpcurl-based testing.

**Consequences:** Simpler architecture, no HTTP translation layer, but requires gRPC client (grpcurl) for manual testing.

---

## ADR-004: Logical CQRS with single database

**Status:** Accepted

**Context:** Auth domain has low read/write asymmetry. Separate stores would add operational complexity without proportional benefit.

**Decision:** Separate command/query handlers and repository interfaces, backed by a single PostgreSQL instance.

**Consequences:** Clean separation of concerns in code, simple deployment, easy to evolve to physical separation if needed.

---

## ADR-005: Argon2id for password hashing

**Status:** Accepted

**Context:** Need memory-hard password hashing resistant to GPU/ASIC attacks.

**Decision:** Argon2id with parameters: time=1, memory=64MB, threads=4, keyLen=32.

**Consequences:** Higher memory usage per hash operation, but significantly better security than bcrypt against modern attack vectors.

---

## ADR-006: Mock outbox for reset token delivery

**Status:** Accepted

**Context:** Challenge needs demonstrable reset flow without external email infrastructure.

**Decision:** Domain events are published to a mock outbox that logs to stdout. Raw reset token is logged but never stored in DB.

**Consequences:** Full flow testable locally, clear path to production outbox pattern with real email integration.

---

## ADR-007: ES256 (ECDSA P-256) для JWT

**Status:** Accepted

**Context:** Планируется multi-service архитектура. При симметричной подписи (HS256) каждый сервис, верифицирующий JWT, должен знать секрет — и может подделать токены.

**Decision:** Асимметричная подпись ES256. Приватный ключ только у auth-сервиса, публичный раздаётся потребителям. Ключ генерируется при старте (для challenge), в production — из Vault/KMS.

**Consequences:** Другие сервисы могут верифицировать JWT без доступа к секрету подписи. Токены на ~32 байта длиннее чем HS256. Путь миграции на ES512 — одна строка.

---

## ADR-008: Redis-backed rate limiting

**Status:** Accepted

**Context:** Rate limiting требуется для login и password reset. Планируется multi-instance deployment.

**Decision:** Sliding Window Counter на Redis (Lua-скрипт, атомарный). Два соседних окна с взвешенным подсчётом устраняют burst на границе fixed window.

**Отклонённые альтернативы:**
- Fixed Window Counter — проще, но 2x burst на стыке окон.
- Sliding Window Log — точен, но O(n) памяти на каждый ключ.
- Token Bucket — допускает burst by design, не подходит для auth-операций.

**Consequences:** Корректно работает при горизонтальном масштабировании. Атомарность через Lua. Добавляет зависимость на Redis. In-memory реализация остаётся для unit-тестов.

---

## ADR-009: Embedded migrations через golang-migrate

**Status:** Accepted

**Context:** Нужна стратегия эволюции схемы данных с версионированием, идемпотентностью и возможностью rollback.

**Decision:** `golang-migrate` с файлами миграций, вшитыми в бинарник через `go:embed`. Отдельный контейнер `migrator` (в k8s — init-container / Job) выполняет миграции до старта сервисов. Auth-service не содержит логики миграций — разделение ответственности.

**Стратегия эволюции схемы:**
- Каждое изменение — пара файлов `up/down` с последовательным номером версии.
- Backward compatibility: expand-contract pattern (сначала добавляем новое, потом убираем старое через отдельную миграцию).
- В таблице `schema_migrations` фиксируется текущая версия и dirty-флаг.
- Rollback через `down`-миграции при необходимости.

**Отклонённые альтернативы:**
- SQL-файлы в Docker entrypoint — нет версионирования, нет идемпотентности, не работает без Docker.
- Atlas — декларативный подход, мощнее, но сложнее для challenge scope.

**Consequences:** Миграции портативны (вшиты в бинарник), версионированы, откатываемы. Сервис автономен — не зависит от внешнего migration runner.

**Production evolution:** мигратор стоит вынести в shared Go-модуль (или заменить на language-agnostic инструмент вроде Atlas), чтобы переиспользовать между сервисами. Каждый сервис поставляет свои SQL-файлы, общий движок — логику версионирования и выполнения. В рамках challenge оставлен в одном репозитории для простоты проверки.

---

## ADR-010: Kubernetes manifests

**Status:** Accepted

**Context:** Challenge даёт бонус за Terraform/Kubernetes/Helm. Docker Compose покрывает локальную разработку, но не production deployment.

**Decision:** Набор K8s манифестов в `k8s/` с Kustomize как entry point. PostgreSQL и Redis — внешние managed-сервисы, не деплоятся в кластер.

**Состав:**
- `Namespace` — изоляция.
- `ConfigMap` / `Secret` — разделение конфигурации и секретов.
- `Job` (migrator) — pre-deploy миграции (аннотации совместимы с Helm hooks).
- `Deployment` (2 реплики) — rolling update (maxUnavailable=0), gRPC readiness/liveness probes, resource requests/limits.
- `Service` (ClusterIP) — внутрикластерный gRPC endpoint.
- `NetworkPolicy` — zero-trust: ingress только на 50051, egress только PG + Redis + DNS.
- `PodDisruptionBudget` — minAvailable=1 при rolling update и node drain.

**Отклонённые альтернативы:**
- Helm chart — мощнее для multi-env, но избыточен для одного сервиса. Kustomize проще валидировать. Путь к Helm описан в README.
- Terraform — требует реальный cloud-аккаунт, невозможно проверить локально.

**Consequences:** Манифесты проверяемы через `kubectl apply --dry-run=client`. Покрывают zero-trust networking, disruption budget, health checks. Естественный переход от Docker Compose к оркестрации.
