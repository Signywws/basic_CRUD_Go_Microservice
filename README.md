# 📝 Notes API (Golang + PostgreSQL)

Простой REST API микросервис для управления заметками, написанный на Go с использованием стандартной библиотеки `net/http` и `database/sql`.

Проект реализует полный CRUD (Create, Read, Update, Delete) и демонстрирует базовую backend-архитектуру микросервиса:

```
Handler → Service → Repository → PostgreSQL
```

---

## 🚀 Возможности

* Создание заметки
* Получение всех заметок
* Получение заметки по ID
* Обновление заметки
* Удаление заметки
* Работа с PostgreSQL через `database/sql`
* Обработка ошибок (`sql.ErrNoRows`)
* Разделение логики по слоям

---

## 🧱 Архитектура

Проект разделён на три слоя:

### 1. Handler (HTTP слой)

Отвечает за:

* обработку HTTP-запросов
* парсинг JSON
* возврат HTTP-ответов

### 2. Service (бизнес-логика)

Отвечает за:

* валидацию данных
* бизнес-правила

### 3. Repository (работа с БД)

Отвечает за:

* выполнение SQL-запросов
* взаимодействие с PostgreSQL

---

## 🗄️ База данных

Используется PostgreSQL.

### Таблица `notes`

```sql
CREATE TABLE IF NOT EXISTS notes (
    id SERIAL PRIMARY KEY,
    text TEXT NOT NULL
);
```

---

## ⚙️ Настройка и запуск

### 1. Установить PostgreSQL

Создать базу данных:

```sql
CREATE DATABASE notesdb;
```

---

### 2. Настроить строку подключения

В `main.go`:

```go
dsn := "user=postgres password=YOUR_PASSWORD host=localhost port=5432 dbname=notesdb sslmode=disable"
```

---

### 3. Установить зависимости

```bash
go mod tidy
```

---

### 4. Запустить сервер

```bash
go run main.go
```

Сервер будет доступен на:

```
http://localhost:8080
```

---

## 📡 API

### 🔹 Создать заметку

```
POST /notes
```

Body:

```json
{
  "text": "my note"
}
```

---

### 🔹 Получить все заметки

```
GET /notes
```

---

### 🔹 Получить заметку по ID

```
GET /notes?id=1
```

---

### 🔹 Обновить заметку

```
PUT /notes?id=1
```

Body:

```json
{
  "text": "updated note"
}
```

---

### 🔹 Удалить заметку

```
DELETE /notes?id=1
```

---

## ⚠️ Обработка ошибок

* `400 Bad Request` — неверные данные
* `404 Not Found` — заметка не найдена
* `500 Internal Server Error` — ошибка сервера

---

## 🧠 Ключевые моменты реализации

### 🔹 `QueryRow` vs `Query` vs `Exec`

| Метод      | Использование           |
| ---------- | ----------------------- |
| `QueryRow` | одна строка             |
| `Query`    | несколько строк         |
| `Exec`     | без возвращаемых данных |

---

### 🔹 `UPDATE ... RETURNING`

Используется для:

* обновления записи
* получения результата одним запросом

---

### 🔹 `RowsAffected`

Используется при `DELETE` для проверки:

* была ли реально удалена запись

---

## 📌 Технологии

* Go (Golang)
* PostgreSQL
* `database/sql`
* `pgx` driver

---

## 📈 Возможные улучшения

* Добавить middleware (логирование, recovery)
* Вынести конфигурацию в env
* Добавить Docker
* Добавить JWT аутентификацию
* Добавить тесты
* Использовать router (например `chi`)

---

## 👨‍💻 Автор

Проект создан в учебных целях для изучения backend-разработки на Go.
