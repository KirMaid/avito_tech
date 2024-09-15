package controllers

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strconv"
	"zadanie-6105/internal/models"
)

var validate = validator.New()

type CreateTenderRequest struct {
	Name            string    `json:"name" validate:"required"`
	Description     string    `json:"description"`
	ServiceType     string    `json:"serviceType" validate:"required"`
	OrganizationID  uuid.UUID `json:"organizationId" validate:"required"`
	CreatorUsername string    `json:"creatorUsername" validate:"required"`
}

type UpdateTenderRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func checkOrganizationResponsibility(db *gorm.DB, userID uuid.UUID, organizationID uuid.UUID) error {
	var orgResp models.OrganizationResponsible
	if err := db.Where("user_id = ? AND organization_id = ?", userID, organizationID).First(&orgResp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusForbidden, "Пользователь не имеет доступа")
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
			"error": "Ошибка при получении тендеров",
		})
	}
	return c.Status(200).JSON(tenders)
}

func CreateTender(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var request CreateTenderRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Некорректный запрос",
		})
	}

	if err := validate.Struct(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"reason": "Ошибка валидации данных",
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
		Status:          models.StatusCreated,
		OrganizationID:  request.OrganizationID,
		CreatorUsername: request.CreatorUsername,
	}

	if err := db.Create(&tender).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Не удалось создать тендер",
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
	// Преобразование параметров limit и offset в целые числа
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

	var tenders []models.TenderResponse
	query := db.Model(&models.Tender{})
	query = query.Limit(limit).Offset(offset)
	if err := query.Where("creator_username = ?", username).Find(&tenders).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"reason": "Ошибка при получении тендеров",
		})
	}

	if len(tenders) == 0 {
		return c.Status(200).JSON([]models.TenderResponse{})
	}

	return c.Status(200).JSON(tenders)
}

func UpdateTender(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	tenderID := c.Params("tenderId")
	username := c.Query("username")

	if tenderID == "" || username == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Не указаны обязательные параметры",
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
			"error": "Ошибка при проверке пользователя",
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
			"error": "Ошибка при получении тендера",
		})
	}

	if tender.OrganizationID == uuid.Nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Организация не указана в тендере",
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
			"error": "Некорректные данные запроса",
		})
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
			"error": "Ошибка при сохранении изменений",
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
			"error": "Неверный формат запроса или его параметры.",
		})
	}

	var status models.StatusType
	switch models.StatusType(newStatus) {
	case models.StatusCreated, models.StatusPublished, models.StatusClosed:
		status = models.StatusType(newStatus)
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Некорректное значение статуса.",
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
			"error": "Ошибка при проверке пользователя",
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
			"error": "Ошибка при получении тендера",
		})
	}

	c.Locals("user", user)
	if err := checkOrganizationResponsibility(db, user.ID, tender.OrganizationID); err != nil {
		return err
	}

	tender.Status = status
	if err := db.Save(&tender).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при обновлении статуса тендера",
		})
	}

	return c.Status(200).JSON(tender)
}

func GetTenderStatus(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	tenderID := c.Params("tenderId")
	username := c.Query("username")

	if username == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Параметр username обязателен",
		})
	}

	var user models.Employee
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(400).JSON(fiber.Map{
				"reason": "Пользователь не найден",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при проверке пользователя",
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
			"error": "Ошибка при получении тендера",
		})
	}

	if tender.Status == "CREATED" || tender.Status == "CLOSED" {
		var orgResp models.OrganizationResponsible
		if err := db.Where("user_id = ? AND organization_id = ?", user.ID, tender.OrganizationID).First(&orgResp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(403).JSON(fiber.Map{
					"error": "Недостаточно прав для выполнения действия.",
				})
			}
			return c.Status(500).JSON(fiber.Map{
				"error": "Ошибка при проверке ответственности за организацию",
			})
		}
	} else if tender.Status != "PUBLISHED" {
		return c.Status(403).JSON(fiber.Map{
			"error": "Недостаточно прав для выполнения действия.",
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
			"error": "Не указаны обязательные параметры",
		})
	}

	var tenderUUID uuid.UUID
	if err := tenderUUID.UnmarshalText([]byte(tenderID)); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Некорректный ID тендера",
		})
	}
	var version int
	if _, err := fmt.Sscan(versionStr, &version); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Некорректная версия",
		})
	}

	var user models.Employee
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(400).JSON(fiber.Map{
				"reason": "Пользователь не найден",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при проверке пользователя",
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
			"error": "Ошибка при получении тендера",
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
			"error": "Ошибка при получении версии тендера",
		})
	}

	tender.Name = tenderVersion.Name
	tender.Description = tenderVersion.Description
	tender.ServiceType = tenderVersion.ServiceType
	tender.Status = models.StatusType(tenderVersion.Status)
	if err := db.Save(&tender).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при сохранении отката",
		})
	}

	return c.Status(200).JSON(tender)
}

func saveTenderVersion(db *gorm.DB, tender models.Tender) error {
	latestVersion := models.TenderVersion{
		TenderID:    tender.ID,
		Version:     getNewVersion(db, tender.ID),
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Status:      string(tender.Status),
	}

	return db.Create(&latestVersion).Error
}

func getNewVersion(db *gorm.DB, tenderID uuid.UUID) int {
	var lastVersion models.TenderVersion
	db.Where("tender_id = ?", tenderID).Order("version desc").First(&lastVersion)
	return lastVersion.Version + 1
}

func CreateBid(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type CreateBidInput struct {
		Name            string `json:"name" validate:"required"`
		Description     string `json:"description"`
		Status          string `json:"status" validate:"required"`
		TenderID        string `json:"tenderId" validate:"required"`
		OrganizationID  string `json:"organizationId" validate:"required"`
		CreatorUsername string `json:"creatorUsername" validate:"required"`
	}

	var input CreateBidInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Некорректные данные запроса",
		})
	}

	// Проверка существования тендера
	var tender models.Tender
	if err := db.First(&tender, "id = ?", input.TenderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"error": "Тендер не найден",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при поиске тендера",
		})
	}

	// Проверка существования организации
	var organization models.Organization
	if err := db.First(&organization, "id = ?", input.OrganizationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"error": "Организация не найдена",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при поиске организации",
		})
	}

	bid := models.Bid{
		Name:            input.Name,
		Description:     input.Description,
		Status:          input.Status,
		TenderID:        uuid.MustParse(input.TenderID),
		OrganizationID:  uuid.MustParse(input.OrganizationID),
		CreatorUsername: input.CreatorUsername,
	}

	if err := db.Create(&bid).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при создании предложения",
		})
	}

	// Возвращаем созданное предложение
	return c.Status(200).JSON(bid)
}

func GetUserBids(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	username := c.Query("username")
	if username == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Не указан username",
		})
	}

	var bids []models.Bid
	if err := db.Where("creator_username = ?", username).Find(&bids).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при получении предложений",
		})
	}

	return c.Status(200).JSON(bids)
}

func GetTenderBids(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	tenderID := c.Params("tenderId")
	if tenderID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Не указан tenderId",
		})
	}

	var bids []models.Bid
	if err := db.Where("tender_id = ?", tenderID).Find(&bids).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при получении предложений для тендера",
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
			"error": "Не указан bidId или версия",
		})
	}

	var bidVersion models.BidVersion
	if err := db.Where("bid_id = ? AND version = ?", uuid.MustParse(bidID), version).First(&bidVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"error": "Версия предложения не найдена",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при поиске версии предложения",
		})
	}

	var bid models.Bid
	if err := db.First(&bid, "id = ?", uuid.MustParse(bidID)).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при поиске предложения",
		})
	}

	bid.Name = bidVersion.Name
	bid.Description = bidVersion.Description
	bid.Status = bidVersion.Status
	bid.Version = bidVersion.Version

	if err := db.Save(&bid).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при сохранении отката",
		})
	}

	return c.Status(200).JSON(bid)
}

func EditBid(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	bidID := c.Params("bidId")
	if bidID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Не указан bidId",
		})
	}

	type EditBidInput struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var input EditBidInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Некорректные данные запроса",
		})
	}

	var bid models.Bid
	if err := db.First(&bid, "id = ?", uuid.MustParse(bidID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Предложение не найдено",
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
			"error": "Ошибка при сохранении версии предложения",
		})
	}

	bid.Name = input.Name
	bid.Description = input.Description
	bid.Version += 1

	if err := db.Save(&bid).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при обновлении предложения",
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
			"error": "Не указаны обязательные параметры",
		})
	}

	var reviews []models.Review
	if err := db.Where("bid_id = ? AND author_username = ? AND organization_id = ?", uuid.MustParse(tenderID), authorUsername, uuid.MustParse(organizationID)).Find(&reviews).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка при получении отзывов",
		})
	}

	return c.Status(200).JSON(reviews)
}
