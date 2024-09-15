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

func GetTenders(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var tenders []models.Tender
	serviceType := c.Query("serviceType")

	query := db.Model(&models.Tender{})
	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}

	if err := query.Find(&tenders).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении тендеров",
		})
	}
	return c.Status(200).JSON(tenders)
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

func GetMyTenders(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	username := c.Query("username")
	limitStr := c.Query("limit", "5")   // Значение по умолчанию 5
	offsetStr := c.Query("offset", "0") // Значение по умолчанию 0

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
	if err := db.Where("creator_username = ?", username).Limit(limit).Offset(offset).Find(&tenders).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении тендеров",
		})
	}

	if len(tenders) == 0 {
		return c.Status(200).JSON([]fiber.Map{})
	}

	// Получение всех версий тендеров
	var tenderVersions []models.TenderVersion
	tenderIDs := make([]uuid.UUID, len(tenders))
	for i, tender := range tenders {
		tenderIDs[i] = tender.ID
	}

	if err := db.Where("tender_id IN ?", tenderIDs).Order("version DESC").Find(&tenderVersions).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении версий тендеров",
		})
	}

	// Создание ответа
	var response []fiber.Map
	for _, tender := range tenders {
		// Фильтруем версии для текущего тендера
		var versions []models.TenderVersion
		for _, version := range tenderVersions {
			if version.TenderID == tender.ID {
				versions = append(versions, version)
			}
		}

		for _, version := range versions {
			tenderData := fiber.Map{
				"id":             tender.ID,
				"name":           tender.Name,
				"description":    tender.Description,
				"serviceType":    tender.ServiceType,
				"status":         tender.Status,
				"organizationId": tender.OrganizationID,
				"createdAt":      tender.CreatedAt,
				"version":        version.Version,
			}

			response = append(response, tenderData)
		}
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

	version := 1
	var lastVersion models.TenderVersion
	if err := db.Where("tender_id = ?", tender.ID).Order("version desc").First(&lastVersion).Error; err == nil {
		version = lastVersion.Version + 1
	}

	tenderVersion := models.TenderVersion{
		ID:          uuid.New(),
		TenderID:    tender.ID,
		Version:     version,
		Name:        request.Name,
		Description: request.Description,
		ServiceType: request.ServiceType,
		Status:      tender.Status,
		CreatedAt:   time.Now(),
	}

	if request.Name != "" {
		tender.Name = request.Name
	}
	if request.Description != "" {
		tender.Description = request.Description
	}
	if request.ServiceType != "" {
		tender.ServiceType = request.ServiceType
	}

	if err := db.Save(&tender).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при сохранении изменений",
		})
	}

	if err := db.Create(&tenderVersion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при сохранении версии тендера",
		})
	}

	return c.Status(200).JSON(tender)
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

	newVersion := models.TenderVersion{
		TenderID:    tender.ID,
		Version:     tenderVersion.Version + 1,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Status:      tender.Status,
		CreatedAt:   time.Now(), // Время создания новой версии
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
		Version:        newVersion.Version, // Добавляем версию новой записи
	}

	return c.Status(200).JSON(response)
}

func CreateBid(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type CreateBidInput struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
		TenderID    string `json:"tenderId" validate:"required"`
		AuthorType  string `json:"authorType" validate:"required"`
		AuthorID    string `json:"authorId" validate:"required"`
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
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	var authorType models.AuthorType
	switch input.AuthorType {
	case "User":
		authorType = models.AuthorTypeUser
	case "Organization":
		authorType = models.AuthorTypeOrganization
	default:
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
		})
	}

	authorID, err := uuid.Parse(input.AuthorID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Неверный формат запроса или его параметры.",
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
			"reason": "Ошибка при поиске тендера",
		})
	}

	var organization models.Organization
	if err := db.First(&organization, "id = ?", authorID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"reason": "Организация не найдена",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при поиске организации",
		})
	}

	if organization.ID != tender.OrganizationID {
		return c.Status(403).JSON(fiber.Map{
			"reason": "Недостаточно прав для выполнения действия.",
		})
	}

	bid := models.Bid{
		Name:        input.Name,
		Description: input.Description,
		Status:      models.BidStatusCreated,
		TenderID:    tenderID,
		AuthorType:  authorType,
		AuthorID:    authorID,
		Version:     1,
	}

	if err := db.Create(&bid).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при создании предложения",
		})
	}

	response := fiber.Map{
		"id":          bid.ID.String(),
		"name":        bid.Name,
		"description": bid.Description,
		"status":      "Created",
		"tenderId":    bid.TenderID.String(),
		"createdAt":   bid.CreatedAt.Format(time.RFC3339),
		"authorType":  input.AuthorType,
		"authorId":    bid.AuthorID.String(),
		"version":     bid.Version,
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

	var bids []models.Bid
	if err := db.Where("creator_username = ?", username).Find(&bids).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении предложений",
		})
	}

	return c.Status(200).JSON(bids)
}

func GetTenderBids(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	tenderID := c.Params("tenderId")
	if tenderID == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Не указан tenderId",
		})
	}

	var bids []models.Bid
	if err := db.Where("tender_id = ?", tenderID).Find(&bids).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении предложений для тендера",
		})
	}

	return c.Status(200).JSON(bids)
}

func RollbackBidVersion(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	bidID := c.Params("bidId")
	version := c.Params("version")

	if bidID == "" || version == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Не указан bidId или версия",
		})
	}

	var bidVersion models.BidVersion
	if err := db.Where("bid_id = ? AND version = ?", uuid.MustParse(bidID), version).First(&bidVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"reason": "Версия предложения не найдена",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при поиске версии предложения",
		})
	}

	var bid models.Bid
	if err := db.First(&bid, "id = ?", uuid.MustParse(bidID)).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при поиске предложения",
		})
	}

	bid.Name = bidVersion.Name
	bid.Description = bidVersion.Description
	bid.Status = bidVersion.Status
	bid.Version = bidVersion.Version

	if err := db.Save(&bid).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при сохранении отката",
		})
	}

	return c.Status(200).JSON(bid)
}

func EditBid(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	bidID := c.Params("bidId")
	if bidID == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Не указан bidId",
		})
	}

	type EditBidInput struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var input EditBidInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Некорректные данные запроса",
		})
	}

	var bid models.Bid
	if err := db.First(&bid, "id = ?", uuid.MustParse(bidID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"reason": "Предложение не найдено",
		})
	}

	bidVersion := models.BidVersion{
		ID:          uuid.New(),
		BidID:       bid.ID,
		Name:        bid.Name,
		Description: bid.Description,
		Status:      bid.Status,
		Version:     bid.Version,
		CreatedAt:   bid.CreatedAt,
	}

	if err := db.Create(&bidVersion).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при сохранении версии предложения",
		})
	}

	bid.Name = input.Name
	bid.Description = input.Description
	bid.Version += 1

	if err := db.Save(&bid).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при обновлении предложения",
		})
	}

	return c.Status(200).JSON(bid)
}

func GetBidReviews(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	tenderID := c.Params("tenderId")
	authorUsername := c.Query("authorUsername")
	organizationID := c.Query("organizationId")

	if tenderID == "" || authorUsername == "" || organizationID == "" {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Не указаны обязательные параметры",
		})
	}

	var reviews []models.Review
	if err := db.Where("bid_id = ? AND author_username = ? AND organization_id = ?", uuid.MustParse(tenderID), authorUsername, uuid.MustParse(organizationID)).Find(&reviews).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении отзывов",
		})
	}

	return c.Status(200).JSON(reviews)
}
