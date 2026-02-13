# TURN Server

TURN (Traversal Using Relays around NAT) сервер для ретрансляции WebRTC трафика между устройствами, когда прямое P2P соединение невозможно.

## Назначение

TURN сервер используется как fallback механизм, когда:
- Устройства находятся за строгим NAT
- Файрвол блокирует прямые P2P соединения
- Прямое соединение невозможно по другим причинам

## Конфигурация

Переменные окружения:

```bash
TURN_PORT=3478                    # UDP порт для TURN сервера
TURN_USERNAME=user                 # Username для аутентификации
TURN_PASSWORD=password             # Password для аутентификации
TURN_REALM=local                   # Realm для TURN
TURN_PUBLIC_IP=                   # Публичный IP сервера (опционально)
```

**Важно:** В production обязательно укажите `TURN_PUBLIC_IP` - реальный публичный IP адрес вашего сервера.

## Использование

### Получение TURN credentials через API

```bash
GET /api/v1/webrtc/turn-credentials
Authorization: Bearer {jwt_token}
```

Ответ:
```json
{
  "turn_servers": [
    "turn:localhost:3478?transport=udp",
    "turn:user:password@localhost:3478?transport=tcp"
  ],
  "username": "user",
  "password": "password",
  "realm": "local"
}
```

### Использование в WebRTC клиенте

```javascript
// Получаем TURN credentials
const response = await fetch('/api/v1/webrtc/turn-credentials', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const { turn_servers, username, password } = await response.json();

// Настраиваем WebRTC
const pc = new RTCPeerConnection({
  iceServers: [
    // STUN серверы (из конфигурации)
    { urls: 'stun:stun.l.google.com:19302' },
    // TURN серверы
    {
      urls: turn_servers,
      username: username,
      credential: password
    }
  ]
});
```

## Архитектура

### Компоненты

1. **Server** - основной TURN сервер
   - Управляет UDP listener
   - Обрабатывает TURN запросы
   - Ретранслирует трафик между устройствами

2. **AuthHandler** - обработчик аутентификации
   - Проверяет username/password
   - Генерирует auth key для TURN протокола

3. **RelayAddressGenerator** - генератор relay адресов
   - Определяет IP адрес для ретрансляции
   - Использует публичный IP сервера

### Поток данных

```
Устройство A ──┐
               │ TURN Request
            TURN ────> Relay ────> Устройство B
               │ Server  Address
Устройство C ──┘
```

1. Устройство A подключается к TURN серверу
2. TURN сервер создает relay адрес
3. Устройство B подключается к relay адресу
4. TURN сервер ретранслирует трафик между устройствами

## Производительность

- **Пропускная способность:** Зависит от сервера (обычно 100-1000 Мбит/с)
- **Задержка:** +10-50 мс по сравнению с прямым P2P
- **Нагрузка:** Весь трафик проходит через сервер

## Безопасность

- Аутентификация через username/password
- В production рекомендуется:
  - Использовать временные токены вместо статичных credentials
  - Ограничить доступ по IP
  - Мониторить использование TURN сервера

## Мониторинг

TURN сервер логирует:
- Подключения устройств
- Ошибки аутентификации
- Ошибки ретрансляции

## Troubleshooting

### Устройства не могут подключиться

1. Проверьте, что `TURN_PUBLIC_IP` указан правильно
2. Убедитесь, что порт `TURN_PORT` открыт в файрволе
3. Проверьте логи на ошибки аутентификации

### Медленная передача

- TURN сервер всегда медленнее прямого P2P
- Убедитесь, что сервер имеет достаточную пропускную способность
- Рассмотрите использование нескольких TURN серверов для балансировки

## Дальнейшие улучшения

- [ ] Динамическая генерация временных credentials
- [ ] Интеграция с Redis для кэширования credentials
- [ ] Метрики использования TURN сервера
- [ ] Автоматическое определение публичного IP через STUN
