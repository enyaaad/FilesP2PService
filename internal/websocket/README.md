# WebRTC Signaling Server

WebSocket сервер для WebRTC signaling между устройствами пользователя.

## Endpoint

```
ws://localhost:8081/ws/signaling?device_id={device_id}&device_token={device_token}
```

## Подключение

Устройство подключается к WebSocket серверу, передавая:
- `device_id` - UUID устройства
- `device_token` - токен устройства (из БД)

Сервер проверяет валидность токена и регистрирует устройство в Hub.

## Формат сообщений

Все сообщения передаются в формате JSON.

### SDP Offer

Отправка SDP offer от устройства A к устройству B:

```json
{
  "type": "offer",
  "to_device_id": "device-b-uuid",
  "sdp": {
    "type": "offer",
    "sdp": "v=0\r\no=- 123456789 123456789 IN IP4 127.0.0.1\r\n..."
  }
}
```

Сервер автоматически добавит `from_device_id` и перешлет сообщение целевому устройству.

### SDP Answer

Отправка SDP answer от устройства B к устройству A:

```json
{
  "type": "answer",
  "to_device_id": "device-a-uuid",
  "sdp": {
    "type": "answer",
    "sdp": "v=0\r\no=- 987654321 987654321 IN IP4 127.0.0.1\r\n..."
  }
}
```

### ICE Candidate

Отправка ICE candidate:

```json
{
  "type": "ice-candidate",
  "to_device_id": "device-b-uuid",
  "candidate": {
    "candidate": "candidate:1 1 UDP 2130706431 192.168.1.100 54321 typ host",
    "sdpMLineIndex": 0,
    "sdpMid": "audio"
  }
}
```

### Ошибка

Сервер отправляет сообщения об ошибках:

```json
{
  "type": "error",
  "error": "error message"
}
```

## Безопасность

- Устройства должны принадлежать одному пользователю для обмена сообщениями
- Проверка `device_token` при подключении
- Валидация всех входящих сообщений

## Пример использования (JavaScript)

```javascript
const deviceId = 'your-device-id';
const deviceToken = 'your-device-token';
const ws = new WebSocket(`ws://localhost:8081/ws/signaling?device_id=${deviceId}&device_token=${deviceToken}`);

ws.onopen = () => {
  console.log('Connected to signaling server');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  switch (message.type) {
    case 'offer':
      // Обработать SDP offer
      handleOffer(message.sdp);
      break;
    case 'answer':
      // Обработать SDP answer
      handleAnswer(message.sdp);
      break;
    case 'ice-candidate':
      // Добавить ICE candidate
      addIceCandidate(message.candidate);
      break;
    case 'error':
      console.error('Signaling error:', message.error);
      break;
  }
};

// Отправить SDP offer
function sendOffer(toDeviceId, sdp) {
  ws.send(JSON.stringify({
    type: 'offer',
    to_device_id: toDeviceId,
    sdp: {
      type: 'offer',
      sdp: sdp
    }
  }));
}
```

## Архитектура

- **Hub** - управляет всеми подключенными клиентами
- **Client** - представляет подключенное устройство
- **SignalingMessage** - структура сообщения для signaling

## Логирование

Сервер логирует:
- Подключения/отключения устройств
- Ошибки WebSocket соединений
- Предупреждения о недоступных устройствах
