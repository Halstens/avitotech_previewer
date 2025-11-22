## Быстрое тестирование API

После запуска сервиса протестируйте основные сценарии:

### Сценарий 1: Создание команды и PR
```bash
# Создать команду
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true},
      {"user_id": "u3", "username": "Charlie", "is_active": true}
    ]
  }'

# Создать PR (автоматически назначит ревьюеров)
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1", 
    "pull_request_name": "Add feature",
    "author_id": "u1"
  }'

# Проверить статистику
curl http://localhost:8080/stats