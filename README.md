# MJPEG WebSocket streams

## TL;DR
Проект выполнен по ТЗ, с предоставленной БД и базовым фронтом. Тестовый проект.

## Запуск

### Docker Compose
```bash
docker compose up --build
```

### Генерировать из .proto файлов
```bash
protoc --proto_path=./api \                                               ─╯
      --proto_path=./third_party \
      --go_out=paths=source_relative:./api \
      --validate_out=paths=source_relative,lang=go:./api \
      --go-http_out=paths=source_relative:./api \
      --go-grpc_out=paths=source_relative:./api \
      --openapi_out=fq_schema_naming=true,default_response=false:.
```

### SQLC
```bash
sqlc generate
```

## Описание

По пути
`backend/openapi.yaml`
можно посмотреть сгенерированный swagger файл

Технологии которые использовались:
- go-kratos
- sqlc
- goose для миграций

**CRUD**  
  
Для CRUD были реализованы только методы получения всего списка стримов, конретного по ID и обновление стрима по ID.   
Удаление и создание не реализовано!   
Так как при создании не ясно, как генерировать кадры, а для удаления необходимо рабочее создание. В целом для показа уровня кода на беке и фронте должно быть достаточно 2 доп. методов

Для основных ручек CRUD использовалась связка для удобной кодогенерации kratos+sqlc.

На `:9090/metrics` собираются метрики в prometeus.  
Также у сервиса есть `/health` и `/ready` ручки на основном `:8080` порту

**STREAM**

Для стрима MJPEG через WebSocket связка выше не использовалась, 
так как kratos не поддерживает генерацию под вебсокеты, а для корректной работы с БД нужен pgx.pool

Основная проблема с аллокацией памяти в стриминге решалась 
через переиспользование бакетов с чанками в LRU кеше (на базе sync.pool).  
**Эти области кода хорошо прокомментированы.**

