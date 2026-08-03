# Blog Platform API

Расширенное REST API для блог-платформы на Go с чистой архитектурой.

## Технологии

- Go 1.21+
- PostgreSQL
- JWT аутентификация
- Docker & Docker Compose
- Чистая архитектура (Handler → Service → Repository → DB)
- Конкурентность (Worker pool для обработки задач)

## Функциональность

- Регистрация и аутентификация пользователей (JWT)
- CRUD операции с постами
- Комментирование постов
- Отложенная публикация постов с автоматическим планировщиком
- Пагинация списка постов
- Конкурентная обработка отложенных публикаций

## Установка и запуск

### Требования

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Через Docker Compose

1. **Клонируйте репозиторий**

   ```bash
   git clone git@github.com:Maltassarus/platform.git
   cd blog-platform
   ```

2. **Настройте переменные окружения**

   Скопируйте пример конфигурации:

   ```bash
   cp .env.example .env
   ```

   При необходимости отредактируйте `.env`

3. **Запустите контейнеры**

   ```bash
   docker-compose up -d
   ```

   Эта команда:
   - поднимет PostgreSQL (контейнер `blog_postgres`);
   - соберёт и запустит приложение (контейнер `blog_app`);
   - автоматически выполнит миграции базы данных;
   - запустит встроенный планировщик для публикации отложенных постов.

4. **Проверьте работоспособность**

   После запуска API будет доступно по адресу `http://localhost:8080`.

   Проверьте статус:

   ```bash
   curl http://localhost:8080/api/health
   ```

   Ожидаемый ответ:

   ```json
   {"status":"ok"}
   ```

5. **Просмотр логов**

   ```bash
   docker-compose logs -f app  
   docker-compose logs -f postgres 
   ```

6. **Остановка и удаление контейнеров**

   ```bash
   docker-compose down
   ```

   Чтобы также удалить том с данными PostgreSQL (сбросить состояние БД):

   ```bash
   docker-compose down -v
   ```

### Переменные окружения

Основные настройки задаются в файле `.env`:

| Переменная               | Описание                                     | Значение по умолчанию |
|--------------------------|----------------------------------------------|-----------------------|
| `DB_HOST`                | Хост PostgreSQL                              | `localhost`           |
| `DB_PORT`                | Порт PostgreSQL                              | `5432`                |
| `DB_USER`                | Пользователь PostgreSQL                      | `postgres`            |
| `DB_PASSWORD`            | Пароль PostgreSQL                            | `postgres`            |
| `DB_NAME`                | Имя базы данных                              | `blog`                |
| `DB_SSLMODE`             | Режим SSL (`disable` / `require`)            | `disable`             |
| `JWT_SECRET`             | Секретный ключ для JWT (обязательно изменить)| `...`                 |
| `SERVER_PORT`            | Порт, на котором слушает сервер              | `8080`                |
| `PUBLISH_CHECK_INTERVAL` | Интервал проверки отложенных постов (сек)    | `10`                  |
| `WORKER_POOL_SIZE`       | Количество воркеров для публикации           | `5`                   |

## API Эндпоинты

Все эндпоинты возвращают JSON.

### Публичные

- `GET /api/health` – проверка состояния сервера и БД.
- `POST /api/register` – регистрация нового пользователя.
  ```json
  {
    "email": "user@example.com",
    "password": "securepass"
  }
  ```
- `POST /api/login` – вход, получение JWT-токена.
  ```json
  {
    "email": "user@example.com",
    "password": "securepass"
  }
  ```
  Ответ содержит `token`, который необходимо передавать в заголовке `Authorization: Bearer <token>` для защищённых запросов.

### Защищённые (требуют JWT)

- `GET /api/posts?limit=10&offset=0` – список опубликованных постов.
- `POST /api/posts` – создание поста.
  ```json
  {
    "title": "Заголовок",
    "content": "Содержание...",
    "publish_at": "2026-08-05T15:00:00Z"
  }
  ```
- `GET /api/posts/{id}` – получение поста по ID.
- `PUT /api/posts/{id}` – обновление поста.
  ```json
  {
    "title": "Новый заголовок",
    "content": "Новое содержание"
  }
  ```
- `DELETE /api/posts/{id}` – удаление поста.
- `GET /api/posts/{id}/comments` – список комментариев к посту.
- `POST /api/posts/{id}/comments` – добавление комментария.
  ```json
  {
    "content": "Текст комментария"
  }
  ```

