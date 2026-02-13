package handlers

import (
	"context"
	"net/http"

	"github.com/backend-app/backend/internal/api/middleware"
	devicepb "github.com/backend-app/backend/pkg/proto/device"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeviceHandler struct {
	deviceClient devicepb.DeviceServiceClient
}

func NewDeviceHandler(deviceClient devicepb.DeviceServiceClient) *DeviceHandler {
	return &DeviceHandler{
		deviceClient: deviceClient,
	}
}

type RegisterDeviceRequest struct {
	Name       string `json:"name" binding:"required" example:"My Desktop"`
	DeviceType string `json:"device_type" binding:"required,oneof=desktop mobile" example:"desktop"`
}

type DeviceResponse struct {
	ID          string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID      string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string `json:"name" example:"My Desktop"`
	DeviceType  string `json:"device_type" example:"desktop"`
	DeviceToken string `json:"device_token" example:"550e8400-e29b-41d4-a716-446655440000"`
	LastSeenAt  string `json:"last_seen_at" example:"2024-01-01T00:00:00Z"`
	CreatedAt   string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

type RegisterDeviceResponse struct {
	Device      *DeviceResponse `json:"device"`
	DeviceToken string          `json:"device_token" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type UpdateDeviceRequest struct {
	Name       string `json:"name,omitempty" example:"My Updated Desktop"`
	DeviceType string `json:"device_type,omitempty" binding:"omitempty,oneof=desktop mobile" example:"desktop"`
}

type ListDevicesResponse struct {
	Devices []DeviceResponse `json:"devices"`
	Total   int              `json:"total" example:"5"`
}

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
func (h *DeviceHandler) Register(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.deviceClient.RegisterDevice(context.Background(), &devicepb.RegisterDeviceRequest{
		UserId:     userID.String(),
		Name:       req.Name,
		DeviceType: req.DeviceType,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register device"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register device"})
		return
	}

	c.JSON(http.StatusCreated, RegisterDeviceResponse{
		Device: &DeviceResponse{
			ID:          resp.Device.Id,
			UserID:      resp.Device.UserId,
			Name:        resp.Device.Name,
			DeviceType:  resp.Device.DeviceType,
			DeviceToken: resp.Device.DeviceToken,
			LastSeenAt:  resp.Device.LastSeenAt,
			CreatedAt:   resp.Device.CreatedAt,
			UpdatedAt:   resp.Device.UpdatedAt,
		},
		DeviceToken: resp.DeviceToken,
	})
}

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
func (h *DeviceHandler) Get(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	resp, err := h.deviceClient.GetDevice(context.Background(), &devicepb.GetDeviceRequest{
		DeviceId: deviceID,
		UserId:   userID.String(),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
			case codes.PermissionDenied:
				c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get device"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get device"})
		return
	}

	c.JSON(http.StatusOK, DeviceResponse{
		ID:          resp.Device.Id,
		UserID:      resp.Device.UserId,
		Name:        resp.Device.Name,
		DeviceType:  resp.Device.DeviceType,
		DeviceToken: resp.Device.DeviceToken,
		LastSeenAt:  resp.Device.LastSeenAt,
		CreatedAt:   resp.Device.CreatedAt,
		UpdatedAt:   resp.Device.UpdatedAt,
	})
}

// List godoc
// @Summary Список устройств
// @Description Возвращает список всех устройств пользователя
// @Tags devices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ListDevicesResponse "Список устройств"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /devices [get]
func (h *DeviceHandler) List(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.deviceClient.ListDevices(context.Background(), &devicepb.ListDevicesRequest{
		UserId: userID.String(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list devices"})
		return
	}

	devices := make([]DeviceResponse, len(resp.Devices))
	for i, device := range resp.Devices {
		devices[i] = DeviceResponse{
			ID:          device.Id,
			UserID:      device.UserId,
			Name:        device.Name,
			DeviceType:  device.DeviceType,
			DeviceToken: device.DeviceToken,
			LastSeenAt:  device.LastSeenAt,
			CreatedAt:   device.CreatedAt,
			UpdatedAt:   device.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, ListDevicesResponse{
		Devices: devices,
		Total:   len(devices),
	})
}

// Update godoc
// @Summary Обновление устройства
// @Description Обновляет информацию об устройстве
// @Tags devices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID устройства" format(uuid)
// @Param request body UpdateDeviceRequest true "Данные для обновления"
// @Success 200 {object} DeviceResponse "Устройство обновлено"
// @Failure 400 {object} map[string]string "Неверный формат данных"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Нет доступа к устройству"
// @Failure 404 {object} map[string]string "Устройство не найдено"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /devices/{id} [put]
func (h *DeviceHandler) Update(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	var req UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.deviceClient.UpdateDevice(context.Background(), &devicepb.UpdateDeviceRequest{
		DeviceId:   deviceID,
		UserId:     userID.String(),
		Name:       req.Name,
		DeviceType: req.DeviceType,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
			case codes.PermissionDenied:
				c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update device"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update device"})
		return
	}

	c.JSON(http.StatusOK, DeviceResponse{
		ID:          resp.Device.Id,
		UserID:      resp.Device.UserId,
		Name:        resp.Device.Name,
		DeviceType:  resp.Device.DeviceType,
		DeviceToken: resp.Device.DeviceToken,
		LastSeenAt:  resp.Device.LastSeenAt,
		CreatedAt:   resp.Device.CreatedAt,
		UpdatedAt:   resp.Device.UpdatedAt,
	})
}

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
func (h *DeviceHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	resp, err := h.deviceClient.DeleteDevice(context.Background(), &devicepb.DeleteDeviceRequest{
		DeviceId: deviceID,
		UserId:   userID.String(),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
			case codes.PermissionDenied:
				c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete device"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
	})
}

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
func (h *DeviceHandler) UpdateLastSeen(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	resp, err := h.deviceClient.UpdateLastSeen(context.Background(), &devicepb.UpdateLastSeenRequest{
		DeviceId: deviceID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update last seen"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update last seen"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
	})
}
