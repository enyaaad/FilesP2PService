package handlers

// Этот файл содержит swagger аннотации для всех handlers
// Аннотации добавлены непосредственно в файлы handlers, этот файл служит справочником

// Swagger аннотации для Device handlers:
/*
// Register godoc
// @Summary Регистрация устройства
// @Description Регистрирует новое устройство для пользователя и возвращает device_token для QR-кода
// @Tags devices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RegisterDeviceRequest true "Данные устройства"
// @Success 201 {object} RegisterDeviceResponse "Устройство успешно зарегистрировано"
// @Failure 400 {object} map[string]string "Неверный формат данных"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /devices [post]

// Get godoc
// @Summary Получение устройства
// @Description Возвращает информацию об устройстве по ID
// @Tags devices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID устройства" format(uuid)
// @Success 200 {object} DeviceResponse "Информация об устройстве"
// @Failure 400 {object} map[string]string "Неверный ID устройства"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Нет доступа к устройству"
// @Failure 404 {object} map[string]string "Устройство не найдено"
// @Router /devices/{id} [get]

// List godoc
// @Summary Список устройств
// @Description Возвращает список всех устройств пользователя
// @Tags devices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Список устройств"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Router /devices [get]

// Update godoc
// @Summary Обновление устройства
// @Description Обновляет информацию об устройстве
// @Tags devices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID устройства" format(uuid)
// @Param request body object true "Данные для обновления" schema(UpdateDeviceRequest)
// @Success 200 {object} DeviceResponse "Устройство обновлено"
// @Failure 400 {object} map[string]string "Неверный формат данных"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Нет доступа к устройству"
// @Failure 404 {object} map[string]string "Устройство не найдено"
// @Router /devices/{id} [put]

// Delete godoc
// @Summary Удаление устройства
// @Description Удаляет устройство пользователя
// @Tags devices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID устройства" format(uuid)
// @Success 200 {object} map[string]bool "Устройство удалено"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Нет доступа к устройству"
// @Failure 404 {object} map[string]string "Устройство не найдено"
// @Router /devices/{id} [delete]

// UpdateLastSeen godoc
// @Summary Обновление времени последней активности
// @Description Обновляет время последней активности устройства (не требует аутентификации)
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "ID устройства" format(uuid)
// @Success 200 {object} map[string]bool "Время обновлено"
// @Failure 400 {object} map[string]string "Неверный ID устройства"
// @Failure 404 {object} map[string]string "Устройство не найдено"
// @Router /devices/{id}/last-seen [post]
*/

// Swagger аннотации для File handlers:
/*
// Upload godoc
// @Summary Загрузка файла
// @Description Загружает файл на сервер через multipart/form-data
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Файл для загрузки"
// @Success 201 {object} map[string]interface{} "Файл успешно загружен"
// @Failure 400 {object} map[string]string "Неверный формат данных"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /files [post]

// Download godoc
// @Summary Скачивание файла
// @Description Скачивает файл с сервера (поддерживает Range requests)
// @Tags files
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param id path string true "ID файла" format(uuid)
// @Param offset query int false "Смещение в байтах" default(0)
// @Param limit query int false "Лимит байт для чтения" default(0)
// @Success 200 {file} file "Файл"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Нет доступа к файлу"
// @Failure 404 {object} map[string]string "Файл не найден"
// @Router /files/{id}/download [get]

// GetMetadata godoc
// @Summary Метаданные файла
// @Description Возвращает метаданные файла
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID файла" format(uuid)
// @Success 200 {object} FileResponse "Метаданные файла"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Нет доступа к файлу"
// @Failure 404 {object} map[string]string "Файл не найден"
// @Router /files/{id} [get]

// List godoc
// @Summary Список файлов
// @Description Возвращает список файлов пользователя с пагинацией
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Лимит файлов" default(50)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} ListFilesResponse "Список файлов"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Router /files [get]

// Delete godoc
// @Summary Удаление файла
// @Description Удаляет файл с сервера
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID файла" format(uuid)
// @Success 200 {object} map[string]bool "Файл удален"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Нет доступа к файлу"
// @Failure 404 {object} map[string]string "Файл не найден"
// @Router /files/{id} [delete]
*/

// Swagger аннотации для WebRTC handlers:
/*
// GetTurnCredentials godoc
// @Summary Получение TURN credentials
// @Description Возвращает TURN серверы и credentials для WebRTC
// @Tags webrtc
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} TurnCredentialsResponse "TURN credentials"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Router /webrtc/turn-credentials [get]
*/
