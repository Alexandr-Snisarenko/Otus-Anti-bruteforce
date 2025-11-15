# Anti-Bruteforce — Software Design Description (SDD)

## 1. Введение
**Цель:**  
Документ описывает архитектуру, алгоритмы и технические решения микросервиса **Anti-Bruteforce**, реализующего защиту от подбора паролей при авторизации. 

**Общее описание**
Микросервис **Anti-Bruteforce** выполняет контроль количества повторов параметров авторизации (логин/пароль/IP) за заданный период времени.
Сервис ограничивает частоту попыток авторизации для различных комбинаций параметров, например:
* не более N = 10 попыток в минуту для данного логина.
* не более M = 100 попыток в минуту для данного пароля (защита от обратного brute-force).
* не более K = 1000 попыток в минуту для данного IP.

Для проверки лимитов по параметрам авторизации используется алгоритм Sliding Window, через механизм Redis ZSET.

**Основание:**  
Разработано в соответствии с требованиями, изложенными в документе  
[`docs/requirements.md`](requirements.md).

**Область применения:**  
Документ предназначен для разработчиков, ревьюеров и DevOps-инженеров, поддерживающих сервис.

---

## 2. Назначение и функции
Сервис вызывается перед авторизацией пользователя и возвращает:
- `ok=true` — если попытка разрешена;
- `ok=false` — если превышены лимиты или IP в blacklist.

**Функциональные возможности:**
- Проверка частоты попыток по логину, паролю и IP (gRPS-API)
- Проверка присутствия IP адреса в черном/белом списке подсетей(CIDR)
- Управление whitelist/blacklist подсетей (CIDR).
- Сброс счётчиков (bucket-ов).
- gRPC-API и CLI для администрирования.
- Конфигурирование через YAML/ENV.
- Сборка и запуск через Makefile и Docker Compose.

---

## 3. Архитектура системы
### 3.1 Структура каталогов

```
cmd/
 ├─ anti-bruteforce/    # основной сервис (gRPC)
 └─ abf/                # CLI-клиент для администрирования
internal/
 ├─ app/                # бизнес-координация, use-cases
 ├─ core/
 │   ├─ ratelimit/      # логика rate-limit
 │   └─ subnetlist/     # работа с white/black lists
 ├─ storage/
 │   ├─ redisrl/        # реализация rate-limit в Redis (Sliding Window)
 │   ├─ postgres/       # хранение subnet lists
 │   └─ memory/         # in-memory реализация (тесты)
 ├─ server/grpc/         # реализация сервера
 ├─ ctxmeta/             # работа с контекстом
 ├─ version/             # автоформирование версии
 ├─ config/             # парсинг YAML/env
 └─ logger/             # slog + контекстные поля

api
└─ proto/               # файл .proto для API

deployment/
 ├─ docker-compose.yml
 └─ migrations/
docs/
 ├─ architecture/      # архитектурные схемы (puml)
 ├─ requirements.md
 └─ design.md
```

---

## 4. Алгоритмы
### 4.1 Проверка попытки
1. Проверить IP в **whitelist** → разрешить.  
2. Проверить IP в **blacklist** → отказать.  
3. Проверить лимиты:  
   - `N` попыток/мин по логину  
   - `M` попыток/мин по паролю  
   - `K` попыток/мин по IP  
4. Если все три проверки в норме → `ok=true`.

### 4.2 Sliding Window (Redis ZSET)
Для ведения backets и проверки лимитов используется алгоритм Sliding Window, через Redis ZSET.
Атомарность операций обеспечивается путем работы в redis.Client.TxPipeline()

**Параметры:**  
`window_ms = 60_000`  
`score = timestamp_ms`

**Процедура:**
```text
ZREMRANGEBYSCORE key -inf now-window
ZADD key now member
count = ZCARD key
if count >= limit: deny
PEXPIRE key window
```

**Ключи:**
```
rl:login:{login}
rl:pass:{hash(password)}
rl:ip:{ip}
```

**ResetBucket:** `DEL` соответствующих ключей.

### 4.3 In-memory реализация
- Используется в тестах.
- Структура: `map[string]*bucket` + `[]int64`.
- TTL-очистка при бездействии.

### 4.4 Работа со списками
**Таблица PostgreSQL:**
```sql
CREATE TABLE subnets (
  CIDR TEXT NOT NULL,
  LIST_TYPE TEXT NOT NULL,
  COMMENT TEXT,
  DC TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (CIDR, LIST_TYPE)
);

```
**Предполагаются типы:** whitelist, blacklist.  
**Типы:** IPv4, CIDR-нотация `192.1.1.0/25`.

---

## 5. API (gRPC)
### 5.1 Сервис `AntiBruteforce`
```proto
syntax = "proto3";
package abf.v1;

service AntiBruteforce {
  rpc CheckAttempt(CheckAttemptRequest) returns (CheckAttemptResponse);
  rpc ResetBucket(ResetBucketRequest) returns (ResetBucketResponse);
  rpc AddToWhitelist(ManageCIDRRequest) returns (ManageCIDRResponse);
  rpc RemoveFromWhitelist(ManageCIDRRequest) returns (ManageCIDRResponse);
  rpc AddToBlacklist(ManageCIDRRequest) returns (ManageCIDRResponse);
  rpc RemoveFromBlacklist(ManageCIDRRequest) returns (ManageCIDRResponse);
}
```

### 5.2 Сообщения
```proto
message CheckAttemptRequest {
  string login = 1;
  string password = 2;
  string ip = 3;
}
message CheckAttemptResponse { bool ok = 1; }

message ResetBucketRequest { string login = 1; string ip = 2; }
message ResetBucketResponse {}

message ManageCIDRRequest { string cidr = 1; }
message ManageCIDRResponse {}
```

---

## 6. CLI
Бинарник `cmd/abf`. Работает через gRPC API.

Примеры:
```bash
abf check --login alice --password qwerty --ip 10.1.2.3
abf reset --login alice --ip 10.1.2.3
abf whitelist add 192.0.2.0/24
abf blacklist rm 198.51.100.0/25
```

---

## 7. Безопасность
- Пароль не логируется, в Redis хранится хэш.  
- gRPC может работать поверх TLS.  
- CLI ограничен служебным доступом.

---

