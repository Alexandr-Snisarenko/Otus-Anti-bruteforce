# Anti‑Bruteforce — Архитектура проекта

Документация расположена в каталоге `docs/architecture/`. Все схемы сделаны в PlantUML и могут быть визуализированы как `.svg` или `.png`.

---

##  Структура каталога `docs/architecture/`

```
docs/
├── architecture/
│   ├── anti-bruteforce-components.puml   # Компонентная схема приложения
│   ├── anti-bruteforce-deployment.puml   # Схема развертывания (docker/k8s)
│   ├── anti-bruteforce-classes.puml      # Классы и интерфейсы (core/app)
│   ├── anti-bruteforce-sequence.puml     # Последовательность CheckAttempt
│   └── README.md                         # Текущее описание
```

---

##  1. Component Diagram — структура сервиса

![](./anti-bruteforce-components.svg)

**Описание:**

* `Transport (gRPC/HTTP)` — точка входа API.
* `ApplicationService` — фасад бизнес‑логики.
* `RateLimiter` — ядро, реализующее алгоритм ограничения (Sliding Window).
* `CIDR Matcher` — проверка IP по белому/чёрному спискам.
* `Redis` — хранилище счётчиков (buckets), pub/sub для обновления списков.
* `PostgreSQL` — хранилище белых/чёрных списков.

---

##  2. Deployment Diagram — окружение

![](./anti-bruteforce-deployment.svg)

**Описание:**

* Основной контейнер `anti-bruteforce` — содержит приложение и подписчик pub/sub.
* Взаимодействие с `redis` и `postgres`.
* Клиенты: `Auth Service` (сервер‑сервер API) и `Admin CLI`.

---

##  3. Class/Interface Diagram — основные абстракции

![](./anti-bruteforce-classes.svg)

**Ключевые интерфейсы:**

* `RateLimiter` — предоставляет методы `Allow()` и `Reset()`.
* `SubnetRepo` — работа с Redis.
* `LimiterRepo` — работа с PostgreSQL.
* `SubnetWList`, `SubnetBList` — кеш и проверка IP.
* `ApplicationService` — координирует вызовы, реализует бизнес‑логику.

---

##  4. Sequence Diagram — сценарий CheckAttempt

![](./anti-bruteforce-sequence.svg)

**Порядок взаимодействий:**

1. Клиент вызывает `CheckAttempt`.
2. Сервис проверяет IP в списках (через `SubnetList`).
3. Если IP не в списках — запрашивает `RateLimiter`.
4. Лимитер обновляет счётчики в Redis и возвращает решение.
5. Ответ уходит клиенту.

---

##  Примечания

* Все схемы соответствуют требованиям ТЗ: лимиты N/M/K, списки CIDR, pub/sub обновления.
* Можно генерировать `.svg` из `.puml` через команду:

  ```bash
  make docs
  ```
* Схемы следует хранить в репозитории вместе с кодом, обновляя при изменении архитектуры.
