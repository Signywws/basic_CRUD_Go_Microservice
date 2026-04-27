# Notes Service

## 📌 Description

Notes Service — это backend сервис на Go для управления заметками.

Сервис реализует базовые CRUD-операции:

* создание заметки
* получение списка заметок
* получение заметки по ID
* обновление заметки
* удаление заметки

Сервис построен по многослойной архитектуре:

```text
handler -> service -> repository
```

---

## 🏗 Архитектура

### Handler Layer

Отвечает за:

* обработку HTTP-запросов
* валидацию входных данных
* формирование HTTP-ответов

### Service Layer

Отвечает за:

* бизнес-логику
* правила валидации
* работу с данными через repository

### Repository Layer

Отвечает за:

* работу с базой данных (PostgreSQL)
* выполнение SQL-запросов

---

## 🧱 Стек технологий

* Go (net/http)
* PostgreSQL
* database/sql
* pgx driver

---

## ⚙️ Конфигурация

Сервис настраивается через переменные окружения:

```env
HTTP_PORT=8080
DB_DSN=user=postgres password=admin host=localhost port=5432 dbname=notesbd sslmode=disable
```

---

## 🚀 Запуск

```bash
go run .
```

---

## 📡 API

### Создание заметки

```http
POST /notes
Content-Type: application/json

{
  "text": "My note"
}
```

---

### Получение всех заметок

```http
GET /notes
```

---

### Получение заметки по ID

```http
GET /notes?id=1
```

---

### Обновление заметки

```http
PUT /notes?id=1
Content-Type: application/json

{
  "text": "Updated text"
}
```

---

### Удаление заметки

```http
DELETE /notes?id=1
```

---

## ❤️ Health endpoints

### Проверка доступности сервиса

```http
GET /health
```

Ответ:

```json
{
  "status": "ok"
}
```

---

### Проверка готовности сервиса

```http
GET /health/ready
```

Ответ:

```json
{
  "status": "ready"
}
```

Если БД недоступна:

```json
{
  "status": "not ready"
}
```

---

## 🧠 Особенности

* Чистое разделение слоёв
* Использование интерфейсов
* Dependency Injection
* Проверка состояния сервиса
* Подготовка к масштабированию

---


