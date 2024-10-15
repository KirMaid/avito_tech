package controllers

import (
	"bytes"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"testing"
)

// Helper функция для инициализации мокированной базы данных и GORM
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn: db,
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to initialize gorm: %v", err)
	}

	// Функция очистки после теста
	cleanup := func() {
		db.Close()
	}

	return gormDB, mock, cleanup
}

// Тест успешного создания тендера
func TestCreateTender_Success(t *testing.T) {
	// Инициализация мокированной базы данных
	gormDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	// Создание тестового пользователя
	userID := uuid.New()
	orgID := uuid.New()
	tenderID := uuid.New()
	tenderVersionID := uuid.New()

	// Ожидание запроса на поиск пользователя по username
	mock.ExpectQuery(`SELECT \* FROM "employees" WHERE username = \$1 ORDER BY "employees"\."id" LIMIT 1`).
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
			AddRow(userID, "testuser"))

	// Ожидание запроса на проверку ответственности организации
	mock.ExpectQuery(`SELECT \* FROM "organization_responsibles" WHERE user_id = \$1 AND organization_id = \$2 ORDER BY "organization_responsibles"\."id" LIMIT 1`).
		WithArgs(userID, orgID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "organization_id"}).
			AddRow(uuid.New(), userID, orgID))

	// Ожидание вставки нового тендера
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tenders"`).
		WithArgs(sqlmock.AnyArg(), "Test Tender", "Test Description", "ServiceType1", "CREATED", orgID, "testuser", 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tenderID))
	mock.ExpectCommit()

	// Ожидание вставки версии тендера
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tender_versions"`).
		WithArgs(sqlmock.AnyArg(), tenderID, "Test Tender", "Test Description", "ServiceType1", "CREATED", 1, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tenderVersionID))
	mock.ExpectCommit()

	// Создание тестового приложения Fiber
	app := fiber.New()

	// Создание тестового JSON запроса
	requestBody, err := json.Marshal(map[string]interface{}{
		"name":            "Test Tender",
		"description":     "Test Description",
		"serviceType":     "ServiceType1",
		"organizationId":  orgID,
		"creatorUsername": "testuser",
	})
	assert.NoError(t, err)

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", "/tenders", bytes.NewBuffer(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Создание контекста Fiber
	_, _ = app.Test(req, -1)

	// Создание контекста Fiber с локальной базой данных
	app = fiber.New()
	app.Post("/tenders", func(c *fiber.Ctx) error {
		c.Locals("db", gormDB)
		return CreateTender(c)
	})

	// Отправка запроса
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Проверка статуса ответа
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Проверка ожиданий моков
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// Тест создания тендера с неверными данными (валидация)
func TestCreateTender_InvalidInput(t *testing.T) {
	// Инициализация мокированной базы данных
	gormDB, _, cleanup := setupMockDB(t)
	defer cleanup()

	// Создание тестового приложения Fiber
	app := fiber.New()

	// Создание тестового JSON запроса с отсутствующим обязательным полем
	requestBody, err := json.Marshal(map[string]interface{}{
		"description":     "Test Description",
		"serviceType":     "ServiceType1",
		"organizationId":  uuid.New(),
		"creatorUsername": "testuser",
	})
	assert.NoError(t, err)

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", "/tenders", bytes.NewBuffer(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Создание контекста Fiber с локальной базой данных
	app.Post("/tenders", func(c *fiber.Ctx) error {
		c.Locals("db", gormDB)
		return CreateTender(c)
	})

	// Отправка запроса
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Проверка статуса ответа
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Тест создания тендера, когда пользователь не найден
func TestCreateTender_UserNotFound(t *testing.T) {
	// Инициализация мокированной базы данных
	gormDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	// Ожидание запроса на поиск пользователя по username
	mock.ExpectQuery(`SELECT \* FROM "employees" WHERE username = \$1 ORDER BY "employees"\."id" LIMIT 1`).
		WithArgs("nonexistentuser").
		WillReturnError(gorm.ErrRecordNotFound)

	// Создание тестового приложения Fiber
	app := fiber.New()

	// Создание тестового JSON запроса
	requestBody, err := json.Marshal(map[string]interface{}{
		"name":            "Test Tender",
		"description":     "Test Description",
		"serviceType":     "ServiceType1",
		"organizationId":  uuid.New(),
		"creatorUsername": "nonexistentuser",
	})
	assert.NoError(t, err)

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", "/tenders", bytes.NewBuffer(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Создание контекста Fiber с локальной базой данных
	app.Post("/tenders", func(c *fiber.Ctx) error {
		c.Locals("db", gormDB)
		return CreateTender(c)
	})

	// Отправка запроса
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Проверка статуса ответа
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// Проверка ожиданий моков
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// Тест создания тендера без достаточных прав
func TestCreateTender_Forbidden(t *testing.T) {
	// Инициализация мокированной базы данных
	gormDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	// Создание тестового пользователя
	userID := uuid.New()
	orgID := uuid.New()

	// Ожидание запроса на поиск пользователя по username
	mock.ExpectQuery(`SELECT \* FROM "employees" WHERE username = \$1 ORDER BY "employees"\."id" LIMIT 1`).
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
			AddRow(userID, "testuser"))

	// Ожидание запроса на проверку ответственности организации - возвращает отсутствие записи
	mock.ExpectQuery(`SELECT \* FROM "organization_responsibles" WHERE user_id = \$1 AND organization_id = \$2 ORDER BY "organization_responsibles"\."id" LIMIT 1`).
		WithArgs(userID, orgID).
		WillReturnError(gorm.ErrRecordNotFound)

	// Создание тестового приложения Fiber
	app := fiber.New()

	// Создание тестового JSON запроса
	requestBody, err := json.Marshal(map[string]interface{}{
		"name":            "Test Tender",
		"description":     "Test Description",
		"serviceType":     "ServiceType1",
		"organizationId":  orgID,
		"creatorUsername": "testuser",
	})
	assert.NoError(t, err)

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", "/tenders", bytes.NewBuffer(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Создание контекста Fiber с локальной базой данных
	app.Post("/tenders", func(c *fiber.Ctx) error {
		c.Locals("db", gormDB)
		return CreateTender(c)
	})

	// Отправка запроса
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Проверка статуса ответа
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	// Проверка ожиданий моков
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
