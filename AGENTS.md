# Recommendation_system

Планирование заданий 3D-печати по оборудованию и операторам: Go-бэкенд `recsys-backend/` (API, веб-интерфейс
в `web/`, Swagger), PostgreSQL 16. Подробно — `README.md` (API, модель данных, алгоритм планирования).

- Сборка и тесты: `cd recsys-backend && go build ./cmd/app && go test ./...`.
- Swagger после изменения API: `swag init -g cmd/app/main.go` (в `recsys-backend`).
- Полный стек: `docker compose up -d --build` из корня; порты 5432 и 8080 только на `127.0.0.1`,
  `restart: "no"`, имя проекта `recommendation_system` (том `recommendation_system_pgdata` с данными).
- Тестовые данные — `POST /api/dev/seed`, очистка — `/api/dev/clear` (только admin).
- `recsys-backend/.env` — локальные настройки, не коммитить и не выводить.
- Сессии хранятся в памяти: после перезапуска сервера нужен повторный вход — это ожидаемо.
