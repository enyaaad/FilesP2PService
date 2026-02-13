package docs

import "github.com/backend-app/backend/internal/api/handlers"

// @title Backend API
// @version 1.0
// @description Backend для файлообмена между устройствами с поддержкой P2P через WebRTC
//
// ## Основные возможности:
// - Регистрация и аутентификация пользователей
// - Управление устройствами
// - Загрузка и скачивание файлов с поддержкой Range requests
// - WebRTC Signaling для P2P передачи файлов
// - TURN сервер для ретрансляции трафика
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

// Swagger models
type (
	// RegisterRequest модель для регистрации
	RegisterRequest handlers.RegisterRequest

	// LoginRequest модель для входа
	LoginRequest handlers.LoginRequest

	// RefreshRequest модель для обновления токена
	RefreshRequest handlers.RefreshRequest

	// AuthResponse модель ответа аутентификации
	AuthResponse handlers.AuthResponse

	// UserResponse модель пользователя
	UserResponse handlers.UserResponse

	// RegisterDeviceRequest модель для регистрации устройства
	RegisterDeviceRequest handlers.RegisterDeviceRequest

	// DeviceResponse модель устройства
	DeviceResponse handlers.DeviceResponse

	// RegisterDeviceResponse модель ответа регистрации устройства
	RegisterDeviceResponse handlers.RegisterDeviceResponse

	// UpdateDeviceRequest модель для обновления устройства
	UpdateDeviceRequest handlers.UpdateDeviceRequest

	// ListDevicesResponse модель списка устройств
	ListDevicesResponse handlers.ListDevicesResponse

	// FileResponse модель файла
	FileResponse handlers.FileResponse

	// ListFilesResponse модель списка файлов
	ListFilesResponse handlers.ListFilesResponse

	// UploadFileResponse модель ответа загрузки файла
	UploadFileResponse handlers.UploadFileResponse

	// TurnCredentialsResponse модель TURN credentials
	TurnCredentialsResponse handlers.TurnCredentialsResponse
)
