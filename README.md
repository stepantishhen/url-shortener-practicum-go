# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Профилирование памяти

Профили сохранены в директории `profiles/`:
- `profiles/base.pprof` — профиль до оптимизации
- `profiles/result.pprof` — профиль после оптимизации

### Выполненные оптимизации

1. **`generateID()` в `internal/storage/memory.go`** — добавлен `sync.Pool` для буфера `[6]byte`, что уменьшает нагрузку на GC при интенсивном создании URL.
2. **`FileListener.Notify` в `internal/audit/audit.go`** — устранено создание нового `json.Encoder` при каждом вызове; теперь используется один encoder, созданный в конструкторе.
3. **`MemoryStorage.SaveBatch` в `internal/storage/memory.go`** — устранён промежуточный срез `[]string` для ID; результаты собираются напрямую в `[]BatchOutput`, сокращая число аллокаций на батч.
4. **Хендлеры в `internal/handlers/handler.go`** — `fmt.Sprintf("%s/%s", ...)` заменён конкатенацией строк через `+`; устранено двойное преобразование `string(body)` в `ShortenURL`.

### Результат: `pprof -top -diff_base=profiles/base.pprof profiles/result.pprof`

```
File: storage.test.exe
Type: alloc_space
Showing nodes accounting for -227.85MB, 6.01% of 3793.50MB total
      flat  flat%   sum%        cum   cum%
 -204.67MB  5.40%  5.40%  -204.67MB  5.40%  url-shortener-practicum-go/internal/storage.(*MemoryStorage).GetByUser
  -38.68MB  1.02%  6.42%   -41.68MB  1.10%  url-shortener-practicum-go/internal/storage.(*MemoryStorage).Save
   15.51MB  0.41%  6.01%    19.51MB  0.51%  url-shortener-practicum-go/internal/storage.(*MemoryStorage).SaveBatch
         0     0%  6.01%        2MB 0.053%  sync.(*Pool).Get
```

Отрицательные значения (−227.85 MB суммарно) подтверждают снижение потребления памяти после оптимизации.
