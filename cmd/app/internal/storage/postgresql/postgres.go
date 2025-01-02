package postgresql

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	models2 "zadanie-6105/cmd/app/internal/models"
)

const defaultPort = 3306

type Config struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Database string `json:"database" yaml:"database"`
}

func (c *Config) withDefaults() (conf Config) {
	if c != nil {
		conf = *c
	}
	//TODO Посмотреть, что будет, если не инициализировать строки в структуре
	if conf.Port == 0 {
		conf.Port = defaultPort
	}
	return
}

// TODO Этот блок убрать и переписать
type Dbinstance struct {
	Db *gorm.DB
}

var DB Dbinstance

func ConnectDb() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require TimeZone=Europe/Moscow",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USERNAME"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DATABASE"),
		os.Getenv("POSTGRES_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
		os.Exit(1)
	}

	log.Println("connected")
	db.Logger = logger.Default.LogMode(logger.Info)

	log.Println("running migrations")
	db.AutoMigrate(
		&models2.Employee{},
		&models2.Organization{},
		&models2.OrganizationResponsible{},
		&models2.Bid{},
		&models2.BidVersion{},
		&models2.Review{},
		&models2.Tender{},
		&models2.TenderVersion{},
	)

	DB = Dbinstance{
		Db: db,
	}
}
