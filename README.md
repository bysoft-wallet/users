# Users 
Bysoft users service

## Основные методы:

### http://bysoft.ru/api/v1/users/signIn - логин 

Request
```json
{
  "email": "email@email.ru",
  "password": "testPass123" 
}
```

Response
```json
{
    "access": "eyJhbGciOiJIUzI1NiIsInR...",
    "refresh": "eyJhbGciOiJIUzI1NiIsInR..."
}
```

### http://bysoft.ru/api/v1/users/signUp - регистрация

Request
```json
{
  "email": "email@email.ru",
  "name": "Имя"
  "password": "testPass123" 
}
```

Response
```json
{
    "access": "eyJhbGciOiJIUzI1NiIsInR...",
    "refresh": "eyJhbGciOiJIUzI1NiIsInR..."
}
```

### http://bysoft.ru/api/v1/users/me - получение информации о профиле пользователя

Response

```json
{
  "userId": "be53694e-7b60-4d57-b62f-4acaf5f458a1",
  "email": "win@win.ru",
  "name": "winwin",
  "current_currency": "RUR",
}
```

### http://bysoft.ru/api/v1/users/refresh - получение новых токенов по refresh

Request
```json
{
  "refresh": "eyJhbGciOiJIUzI1NiIsInR..."
}
```

Response
```json
{
    "access": "eyJhbGciOiJIUzI1NiIsInR...",
    "refresh": "eyJhbGciOiJIUzI1NiIsInR..."
}
```

Для авторизации используются JWT токены с полями в payload:
```json
{
  "userId": "be53694e-7b60-4d57-b62f-4acaf5f458a1",
  "exp": 1668001276,
}
```

Ожидается что токен будет передаваться при последующих запросах в заголовке X-API-Token
