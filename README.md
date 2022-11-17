# Users 
Bysoft users service

### POST http://bysoft.ru/users/api/v1/signIn - логин 

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

### POST http://bysoft.ru/users/api/v1/signUp - регистрация

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

### GET http://bysoft.ru/users/api/v1/me - получение информации о профиле пользователя
Требуется access-token в заголовке X-API-Token

Response

```json
{
  "userId": "be53694e-7b60-4d57-b62f-4acaf5f458a1",
  "email": "win@win.ru",
  "name": "winwin",
  "settings": {
    "currency": "RUR"
  }
}
```

### PUT http://bysoft.ru/users/api/v1/settings - обновление настроек пользователя 
Требуется access-token в заголовке X-API-Token

Request
```json
 {
    "currency": "RUR"
 }
```

Response

```json
{
  "userId": "be53694e-7b60-4d57-b62f-4acaf5f458a1",
  "email": "win@win.ru",
  "name": "winwin",
  "settings": {
    "currency": "RUR"
  }
}
```

### POST http://bysoft.ru/users/api/v1/refresh - получение новых токенов по refresh

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

### POST http://bysoft.ru/users/api/v1/validate_email - валидация email адреса

Response

HTTP 200 - валиден
HTTP 422 - невелиден 

### Ожидается что токен будет передаваться при последующих запросах в заголовке X-API-Token
