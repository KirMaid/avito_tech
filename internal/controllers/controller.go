package controllers

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
	"zadanie-6105/internal/models"
)

var validate = validator.New()

func checkOrganizationResponsibility(db *gorm.DB, userID uuid.UUID, organizationID uuid.UUID) error {
	var orgResp models.OrganizationResponsible
	if err := db.Where("user_id = ? AND organization_id = ?", userID, organizationID).First(&orgResp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusForbidden, "Недостаточно прав для выполнения действия.")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Ошибка при проверке ответственности за организацию")
	}

	return nil
}

func CreateTender(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type CreateTenderRequest struct {
		Name            string    `json:"name" validate:"required"`
		Description     string    `json:"description"`
		ServiceType     string    `json:"serviceType" validate:"required"`
		OrganizationID  uuid.UUID `json:"organizationId" validate:"required"`
		CreatorUsername string    `json:"creatorUsername" validate:"required"`
	}

	var request CreateTenderRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Данные неправильно сформированы или не соответствуют требованиям.",
		})
	}

	if err := validate.Struct(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Данные неправильно сформированы или не соответствуют требованиям.",
		})
	}

	var user models.Employee
	if err := db.Where("username = ?", request.CreatorUsername).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(401).JSON(fiber.Map{
				"reason": "Пользователь не существует или некорректен.",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при проверке пользователя",
		})
	}

	var orgResp models.OrganizationResponsible
	if err := db.Where("user_id = ? AND organization_id = ?", user.ID, request.OrganizationID).First(&orgResp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(403).JSON(fiber.Map{
				"reason": "Недостаточно прав для выполнения действия.",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при проверке прав пользователя",
		})
	}

	tender := models.Tender{
		ID:              uuid.New(),
		Name:            request.Name,
		Description:     request.Description,
		ServiceType:     request.ServiceType,
		Status:          models.TenderStatusCreated,
		OrganizationID:  request.OrganizationID,
		CreatorUsername: request.CreatorUsername,
		Version:         1,
	}

	if err := db.Create(&tender).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Не удалось создать тендер",
		})
	}

	// Запись в таблицу версий
	tenderVersion := models.TenderVersion{
		ID:          uuid.New(),
		TenderID:    tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		Version:     1,
		CreatedAt:   tender.CreatedAt,
	}

	if err := db.Create(&tenderVersion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Не удалось создать запись о версии тендера",
		})
	}

	response := models.TenderResponse{
		ID:             tender.ID,
		Name:           tender.Name,
		Description:    tender.Description,
		ServiceType:    tender.ServiceType,
		Status:         tender.Status,
		OrganizationID: tender.OrganizationID,
		Version:        1,
		CreatedAt:      tender.CreatedAt,
	}

	return c.Status(200).JSON(response)
}

func GetUserTenders(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	username := c.Query("username")
	limitStr := c.Query("limit", "5")
	offsetStr := c.Query("offset", "0")

	if username == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Параметр username обязателен",
		})
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Некорректное значение параметра limit",
		})
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Некорректное значение параметра offset",
		})
	}

	var user models.Employee
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(401).JSON(fiber.Map{
				"reason": "Пользователя не существует",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при проверке пользователя",
		})
	}

	var tenders []models.Tender
	if err := db.Where("creator_username = ?", username).
		Order("name ASC").
		Limit(limit).
		Offset(offset).
		Find(&tenders).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении тендеров",
		})
	}

	if len(tenders) == 0 {
		return c.Status(200).JSON([]fiber.Map{})
	}

	var response []fiber.Map

	for _, tender := range tenders {
		var tenderVersions []models.TenderVersion
		if err := db.Where("tender_id = ?", tender.ID).
			Order("version DESC").
			Find(&tenderVersions).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"reason": "Ошибка при получении версий тендера",
			})
		}

		for _, version := range tenderVersions {
			versionData := fiber.Map{
				"id":             tender.ID,
				"name":           version.Name,
				"description":    version.Description,
				"serviceType":    version.ServiceType,
				"status":         version.Status,
				"organizationId": tender.OrganizationID,
				"createdAt":      version.CreatedAt,
				"version":        version.Version,
			}
			response = append(response, versionData)
		}
	}

	return c.Status(200).JSON(response)
}

func GetTenders(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	limitStr := c.Query("limit", "5")
	offsetStr := c.Query("offset", "0")
	serviceType := c.Query("serviceType")
	statusFilter := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Некорректное значение параметра limit",
		})
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Некорректное значение параметра offset",
		})
	}

	// Запрос тендеров с фильтрами
	var tenders []models.Tender
	query := db.Model(&models.Tender{}).Limit(limit).Offset(offset)

	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}

	if statusFilter == "PUBLISHED" {
		query = query.Where("status = ?", "PUBLISHED")
	}

	if err := query.Find(&tenders).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении тендеров",
		})
	}

	if len(tenders) == 0 {
		return c.Status(200).JSON([]fiber.Map{})
	}

	var response []fiber.Map

	for _, tender := range tenders {
		var versions []models.TenderVersion
		if err := db.Where("tender_id = ?", tender.ID).
			Order("name ASC").
			Find(&versions).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"reason": "Ошибка при получении версий тендера",
			})
		}

		for _, version := range versions {
			tenderData := fiber.Map{
				"id":             tender.ID,
				"name":           version.Name,
				"description":    version.Description,
				"serviceType":    version.ServiceType,
				"status":         version.Status,
				"organizationId": tender.OrganizationID,
				"createdAt":      version.CreatedAt,
				"version":        version.Version,
			}
			response = append(response, tenderData)
		}
	}

	return c.Status(200).JSON(response)
}

func GetTenderStatus(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	tenderID := c.Params("tenderId")
	username := c.Query("username")

	if username == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Данные неправильно сформированы или не соответствуют требованиям.",
		})
	}

	parsedTenderID, err := uuid.Parse(tenderID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Данные неправильно сформированы или не соответствуют требованиям.",
		})
	}

	var user models.Employee
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(401).JSON(fiber.Map{
				"reason": "Пользователь не существует или некорректен.",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при проверке пользователя",
		})
	}

	var tender models.Tender
	if err := db.First(&tender, "id = ?", parsedTenderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"reason": "Тендер не найден",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении тендера",
		})
	}

	if tender.Status == "CREATED" || tender.Status == "CLOSED" {
		var orgResp models.OrganizationResponsible
		if err := db.Where("user_id = ? AND organization_id = ?", user.ID, tender.OrganizationID).First(&orgResp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(403).JSON(fiber.Map{
					"reason": "Недостаточно прав для выполнения действия.",
				})
			}
			return c.Status(500).JSON(fiber.Map{
				"reason": "Ошибка при проверке ответственности за организацию",
			})
		}
	} else if tender.Status != "PUBLISHED" {
		return c.Status(403).JSON(fiber.Map{
			"reason": "Недостаточно прав для выполнения действия.",
		})
	}

	return c.SendString(string(tender.Status))
}

func UpdateTenderStatus(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	tenderID := c.Params("tenderId")
	newStatus := c.Query("status")
	username := c.Query("username")

	if tenderID == "" || newStatus == "" || username == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	parsedTenderID, err := uuid.Parse(tenderID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	newStatus = strings.ToUpper(newStatus)

	var status models.TenderStatusType
	switch models.TenderStatusType(newStatus) {
	case models.TenderStatusCreated, models.TenderStatusPublished, models.TenderStatusClosed:
		status = models.TenderStatusType(newStatus)
	default:
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	var user models.Employee
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(401).JSON(fiber.Map{
				"reason": "Пользователь не существует или некорректен.",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при проверке пользователя",
		})
	}

	var tender models.Tender
	if err := db.First(&tender, "id = ?", parsedTenderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"reason": "Тендер не найден",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении тендера",
		})
	}

	if err := checkOrganizationResponsibility(db, user.ID, tender.OrganizationID); err != nil {
		return err
	}

	var latestVersion models.TenderVersion
	if err := db.Where("tender_id = ?", tender.ID).
		Order("version DESC").
		First(&latestVersion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении последней версии тендера",
		})
	}

	latestVersion.Status = status
	if err := db.Save(&latestVersion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при обновлении статуса последней версии тендера",
		})
	}

	tender.Status = status
	if err := db.Save(&tender).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при обновлении статуса тендера",
		})
	}

	response := models.TenderResponse{
		ID:             tender.ID,
		Name:           tender.Name,
		Description:    tender.Description,
		ServiceType:    tender.ServiceType,
		Status:         tender.Status,
		CreatedAt:      tender.CreatedAt,
		OrganizationID: tender.OrganizationID,
	}

	return c.Status(200).JSON(response)
}

func UpdateTender(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	tenderID := c.Params("tenderId")
	username := c.Query("username")

	if tenderID == "" || username == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Данные неправильно сформированы или не соответствуют требованиям.",
		})
	}

	var user models.Employee
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(401).JSON(fiber.Map{
				"reason": "Пользователь не существует или некорректен.",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при проверке пользователя",
		})
	}

	var tender models.Tender
	if err := db.First(&tender, "id = ?", tenderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"reason": "Тендер не найден",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении тендера",
		})
	}

	if err := checkOrganizationResponsibility(db, user.ID, tender.OrganizationID); err != nil {
		return err
	}

	// Обработка данных запроса
	var request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ServiceType string `json:"serviceType"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Некорректные данные запроса",
		})
	}

	isUpdated := false
	if request.Name != "" && request.Name != tender.Name {
		tender.Name = request.Name
		isUpdated = true
	}
	if request.Description != "" && request.Description != tender.Description {
		tender.Description = request.Description
		isUpdated = true
	}
	if request.ServiceType != "" && request.ServiceType != tender.ServiceType {
		tender.ServiceType = request.ServiceType
		isUpdated = true
	}

	var version int
	if isUpdated {
		var lastVersion models.TenderVersion
		if err := db.Where("tender_id = ?", tender.ID).Order("version desc").First(&lastVersion).Error; err == nil {
			version = lastVersion.Version + 1
		} else {
			version = 1
		}

		tenderVersion := models.TenderVersion{
			ID:          uuid.New(),
			TenderID:    tender.ID,
			Version:     version,
			Name:        tender.Name,
			Description: tender.Description,
			ServiceType: tender.ServiceType,
			Status:      tender.Status,
			CreatedAt:   time.Now(),
		}

		if err := db.Create(&tenderVersion).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"reason": "Ошибка при сохранении версии тендера",
			})
		}
	}

	if err := db.Save(&tender).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при сохранении изменений",
		})
	}

	return c.Status(200).JSON(tender)
}

func RollbackTender(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	tenderID := c.Params("tenderId")
	versionStr := c.Params("version")
	username := c.Query("username")

	if tenderID == "" || versionStr == "" || username == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	var tenderUUID uuid.UUID
	if err := tenderUUID.UnmarshalText([]byte(tenderID)); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	var version int
	if _, err := fmt.Sscan(versionStr, &version); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	var user models.Employee
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(401).JSON(fiber.Map{
				"reason": "Пользователь не существует или некорректен.",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при проверке пользователя",
		})
	}

	var tender models.Tender
	if err := db.First(&tender, "id = ?", tenderUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"reason": "Тендер не найден",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении тендера",
		})
	}

	if err := checkOrganizationResponsibility(db, user.ID, tender.OrganizationID); err != nil {
		return err
	}

	var tenderVersion models.TenderVersion
	if err := db.Where("tender_id = ? AND version = ?", tenderUUID, version).First(&tenderVersion).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"reason": "Версия тендера не найдена",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении версии тендера",
		})
	}

	tender.Name = tenderVersion.Name
	tender.Description = tenderVersion.Description
	tender.ServiceType = tenderVersion.ServiceType
	tender.Status = tenderVersion.Status

	if err := db.Save(&tender).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при сохранении изменений тендера",
		})
	}

	var maxVersion int
	if err := db.Model(&models.TenderVersion{}).Where("tender_id = ?", tender.ID).Select("COALESCE(MAX(version), 0)").Scan(&maxVersion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при определении максимальной версии тендера",
		})
	}

	newVersion := models.TenderVersion{
		TenderID:    tender.ID,
		Version:     maxVersion + 1,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Status:      tender.Status,
		CreatedAt:   time.Now(),
	}

	if err := db.Create(&newVersion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при создании новой версии тендера",
		})
	}

	response := models.TenderResponse{
		ID:             tender.ID,
		Name:           tender.Name,
		Description:    tender.Description,
		ServiceType:    tender.ServiceType,
		Status:         tender.Status,
		OrganizationID: tender.OrganizationID,
		CreatedAt:      tender.CreatedAt,
		Version:        newVersion.Version,
	}

	return c.Status(200).JSON(response)
}

func CreateBid(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type CreateBidInput struct {
		Name            string `json:"name" validate:"required"`
		Description     string `json:"description"`
		TenderID        string `json:"tenderId" validate:"required"`
		OrganizationID  string `json:"organizationId" validate:"required"`
		CreatorUsername string `json:"creatorUsername" validate:"required"`
	}

	var input CreateBidInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	tenderID, err := uuid.Parse(input.TenderID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат идентификатора тендера.",
		})
	}

	organizationID, err := uuid.Parse(input.OrganizationID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат идентификатора организации.",
		})
	}

	var tender models.Tender
	if err := db.First(&tender, "id = ?", tenderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"reason": "Тендер не найден.",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при поиске тендера.",
		})
	}

	if tender.OrganizationID != organizationID {
		return c.Status(403).JSON(fiber.Map{
			"reason": "Организация не имеет права делать предложение на этот тендер.",
		})
	}

	bid := models.Bid{
		Name:            input.Name,
		Description:     input.Description,
		Status:          models.BidStatusCreated,
		TenderID:        tenderID,
		OrganizationID:  organizationID,
		Version:         1, // Начальная версия
		CreatorUsername: input.CreatorUsername,
	}

	if err := db.Create(&bid).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при создании предложения.",
		})
	}

	bidVersion := models.BidVersion{
		ID:          uuid.New(),
		BidID:       bid.ID,
		Version:     bid.Version,
		Name:        bid.Name,
		Description: bid.Description,
		Status:      bid.Status,
		CreatedAt:   time.Now(),
	}

	if err := db.Create(&bidVersion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при создании версии предложения.",
		})
	}

	response := fiber.Map{
		"id":              bid.ID.String(),
		"name":            bid.Name,
		"description":     bid.Description,
		"status":          string(bid.Status),
		"tenderId":        bid.TenderID.String(),
		"organizationId":  bid.OrganizationID.String(),
		"creatorUsername": bid.CreatorUsername,
		"createdAt":       bid.CreatedAt.Format(time.RFC3339),
		"version":         bid.Version,
	}

	return c.Status(200).JSON(response)
}

func GetUserBids(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	username := c.Query("username")
	if username == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Не указан username",
		})
	}

	limit, offset := 10, 0
	if l := c.Query("limit"); l != "" {
		var err error
		if limit, err = strconv.Atoi(l); err != nil {
			return c.Status(400).JSON(fiber.Map{"reason": "Некорректное значение limit"})
		}
	}
	if o := c.Query("offset"); o != "" {
		var err error
		if offset, err = strconv.Atoi(o); err != nil {
			return c.Status(400).JSON(fiber.Map{"reason": "Некорректное значение offset"})
		}
	}

	var bids []models.Bid
	if err := db.Where("creator_username = ?", username).
		Limit(limit).Offset(offset).
		Find(&bids).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении предложений",
		})
	}

	// Get the latest version for each bid
	var bidVersions []models.BidVersion
	for _, bid := range bids {
		var latestVersion models.BidVersion
		if err := db.Where("bid_id = ?", bid.ID).
			Order("version DESC").
			First(&latestVersion).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(500).JSON(fiber.Map{
					"reason": "Ошибка при получении версии предложения",
				})
			}
		} else {
			bidVersions = append(bidVersions, latestVersion)
		}
	}

	bidResponses := make([]fiber.Map, len(bids))
	for i, bid := range bids {
		bidResponses[i] = fiber.Map{
			"id":              bid.ID.String(),
			"name":            bid.Name,
			"description":     bid.Description,
			"status":          bid.Status,
			"tenderId":        bid.TenderID.String(),
			"organizationId":  bid.OrganizationID.String(),
			"creatorUsername": bid.CreatorUsername,
			"createdAt":       bid.CreatedAt.Format(time.RFC3339),
			"version":         bid.Version,
		}
	}

	return c.Status(200).JSON(bidResponses)
}

func UpdateBidStatus(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	bidID := c.Params("bidId")
	newStatus := c.Query("status")
	username := c.Query("username")

	if bidID == "" || newStatus == "" || username == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	parsedBidID, err := uuid.Parse(bidID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	newStatus = strings.ToUpper(newStatus)

	var status models.BidStatusType
	switch models.BidStatusType(newStatus) {
	case models.BidStatusCreated, models.BidStatusPublished, models.BidStatusCanceled:
		status = models.BidStatusType(newStatus)
	default:
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	var user models.Employee
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(401).JSON(fiber.Map{
				"reason": "Пользователь не существует или некорректен.",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при проверке пользователя",
		})
	}

	var bid models.Bid
	if err := db.First(&bid, "id = ?", parsedBidID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"reason": "Предложение не найдено",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении предложения",
		})
	}

	if err := checkOrganizationResponsibility(db, user.ID, bid.OrganizationID); err != nil {
		return err
	}

	var latestVersion models.BidVersion
	if err := db.Where("bid_id = ?", bid.ID).
		Order("version DESC").
		First(&latestVersion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении последней версии предложения",
		})
	}

	latestVersion.Status = status
	if err := db.Save(&latestVersion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при обновлении статуса последней версии предложения",
		})
	}

	bid.Status = status
	if err := db.Save(&bid).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при обновлении статуса предложения",
		})
	}

	bidResponse := fiber.Map{
		"id":              bid.ID.String(),
		"name":            bid.Name,
		"description":     bid.Description,
		"status":          bid.Status,
		"tenderId":        bid.TenderID.String(),
		"organizationId":  bid.OrganizationID.String(),
		"creatorUsername": bid.CreatorUsername,
		"createdAt":       bid.CreatedAt.Format(time.RFC3339),
		"version":         bid.Version,
	}

	return c.Status(200).JSON(bidResponse)
}

func GetBidStatus(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	bidID := c.Params("bidId")
	username := c.Query("username")

	if username == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Данные неправильно сформированы или не соответствуют требованиям.",
		})
	}

	parsedBidID, err := uuid.Parse(bidID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Данные неправильно сформированы или не соответствуют требованиям.",
		})
	}

	var user models.Employee
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(401).JSON(fiber.Map{
				"reason": "Пользователь не существует или некорректен.",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при проверке пользователя",
		})
	}

	var bid models.Bid
	if err := db.First(&bid, "id = ?", parsedBidID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"reason": "Предложение не найдено",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении предложения",
		})
	}

	var organization models.Organization
	if err := db.First(&organization, "id = ?", bid.OrganizationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"reason": "Организация не найдена",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении организации",
		})
	}

	var orgResp models.OrganizationResponsible
	if err := db.Where("user_id = ? AND organization_id = ?", user.ID, bid.OrganizationID).First(&orgResp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(403).JSON(fiber.Map{
				"reason": "Недостаточно прав для выполнения действия.",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при проверке ответственности за организацию",
		})
	}

	return c.SendString(string(bid.Status))
}
