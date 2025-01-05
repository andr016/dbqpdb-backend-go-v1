package main

// todo refactor lol

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"dbqpdb-backend-go-v1/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Subject model
type Subject struct {
	ID           int           `json:"subject_id" gorm:"primaryKey;column:subject_id"`
	Name         string        `json:"subject_name" gorm:"column:name"`
	MBTI         string        `json:"mbti" gorm:"column:mbti"`
	SubjectTypes []SubjectType `json:"subject_types" gorm:"foreignKey:SubjectID"`
	ImageURL     string        `json:"image_url" gorm:"column:image_url"`
}

// Typology model
type Typology struct {
	TypologyID  int    `json:"typology_id" gorm:"primaryKey;column:typology_id"`
	Name        string `json:"typology_name" gorm:"column:name"`
	DisplayName string `json:"typology_display_name" gorm:"column:display_name"`
}

var db *gorm.DB

// Type model
type Type struct {
	ID          int    `json:"type_id" gorm:"primaryKey;column:type_id"`
	TypologyID  int    `json:"typology_id" gorm:"primaryKey;column:typology_id"`
	Name        string `json:"type_name" gorm:"column:name"`
	DisplayName string `json:"type_display_name" gorm:"column:display_name"`
	Description string `json:"type_description" gorm:"column:description"`
}

// Type For Subject model
type TypeForSubject struct {
	ID int `json:"type_id" gorm:"primaryKey;column:type_id"`
}

// SubjectType model (associative table)
type SubjectType struct {
	SubjectID  int      `gorm:"primaryKey;column:subject_id"`
	TypologyID int      `gorm:"primaryKey;column:typology_id"`
	TypeID     int      `gorm:"column:type_id"`
	Subject    Subject  `gorm:"foreignKey:SubjectID"`
	Typology   Typology `gorm:"foreignKey:TypologyID"`
	Type       Type     `gorm:"foreignKey:TypeID"`
}

type SubjectResponse struct {
	Subject   string `json:"subject"`
	SubjectID int    `json:"subject_id"`
	Types     []int  `json:"types"`
}

func initDB(dsn string) {
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		//Logger:      logger.Default.LogMode(logger.Silent),
		PrepareStmt: true,
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
}

func getTypesByTypologyID(c *fiber.Ctx) error {
	// Retrieve the typology_id from the URL parameter
	typologyID := c.Params("typology_id")

	// Validate if the typologyID is a valid integer
	parsedTypologyID, err := strconv.Atoi(typologyID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid typology ID"})
	}

	// Query the types associated with the given typology_id
	var types []Type
	if err := db.Where("typology_id = ?", parsedTypologyID).Find(&types).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Types not found for the given typology"})
	}

	// Return the types in the response
	return c.JSON(types)
}

func fetchGetSubjects(c *fiber.Ctx) error {
	var subjects []Subject
	if err := db.Find(&subjects).Error; err != nil {
		// If there's an error, return a 500 Internal Server Error response
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch types",
			"details": err.Error(),
		})
	}

	return c.JSON(subjects)
}

var types []Type

func fetchTypes() error {
	// Fetch the types from the database
	if err := db.Find(&types).Error; err != nil {
		return err
	}
	return nil
}

func getTypes(c *fiber.Ctx) error {
	if err := fetchTypes(); err != nil {
		// If there's an error, return a 500 Internal Server Error response
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch types",
			"details": err.Error(),
		})
	}

	return c.JSON(types)
}

var typologies []Typology

func fetchTypologies() error {
	if err := db.Find(&typologies).Error; err != nil {
		return err
	}
	return nil
}

func getTypologies(c *fiber.Ctx) error {
	// Fetch the typologies from the database using GORM
	if err := fetchTypologies(); err != nil {
		// If there's an error, return a 500 Internal Server Error response
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch typologies",
			"details": err.Error(),
		})
	}

	// Return the typologies as a JSON response
	return c.JSON(typologies)
}

func getTypologyByID(c *fiber.Ctx) error {
	typologyID := c.Params("id")

	var types []Type
	if err := db.Where("id = ?", typologyID).Find(&types).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Typology not found"})
	}
	return c.JSON(types)
}

func getSubjectByID(c *fiber.Ctx) error {
	subjectID := c.Params("id")

	var subject Subject

	// Eager loading related data: subject_types
	if err := db.Preload("SubjectTypes").
		Preload("SubjectTypes.Type").
		First(&subject, subjectID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Subject not found"})
	}

	// Prepare the response structure
	subjectResponse := struct {
		Subject   string `json:"subject"`
		SubjectID int    `json:"subject_id"`
		Types     []int  `json:"types"`
		ImageURL  string `json:"image_url"`
	}{
		Subject:   subject.Name,
		SubjectID: subject.ID,
		ImageURL:  subject.ImageURL,
	}

	// Extract typologies and types from the subject_types association
	// Add related typologies and types
	for _, st := range subject.SubjectTypes {
		subjectResponse.Types = append(subjectResponse.Types, st.Type.ID)
	}
	// Send the response
	return c.JSON(subjectResponse)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func uploadImage(c *fiber.Ctx) error {
	log.Println("fuck")
	// Get the file from the form-data
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("No file uploaded")
	}

	log.Println("got file")
	// Save the file to the local 'uploads' directory
	filePath := filepath.Join("uploads", file.Filename)
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to save file")
	}

	log.Println("saved file")

	// Assuming the subject ID is passed in the URL and the file is saved successfully
	subjectID := c.Params("id")
	// Here, you would update the database with the file path for the subject

	var subject Subject
	if err := db.First(&subject, subjectID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Subject not found")
	}

	// Update the subject record with the image URL
	subject.ImageURL = "/uploads/" + file.Filename
	if err := db.Save(&subject).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update subject with image URL")
	}

	return c.JSON(fiber.Map{
		"message":   "Image uploaded successfully",
		"image_url": subject.ImageURL,
	})
}

func main() {
	const prefix = "/api"

	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	initDB(cfg.Database.DSN)
	fetchTypes()
	fetchTypologies()

	app := fiber.New()
	app.Static("/uploads", "./uploads")

	app.Use(cors.New(cors.Config{
		// Allow all origins to access the resources
		AllowOrigins: "*",
		// Allow specific methods
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		// Allow specific headers
		AllowHeaders: "Content-Type,Authorization",
	}))

	app.Get(prefix+"/subject/:id", getSubjectByID)
	app.Get(prefix+"/typology", getTypologies)
	app.Get(prefix+"/typology/:typology_id", getTypesByTypologyID)
	app.Get(prefix+"/types", getTypes)
	app.Get(prefix+"/subject/", fetchGetSubjects)

	app.Post(prefix+"/upload/subject/:id", uploadImage)

	app.Post(prefix+"/submit/typology/:id", func(c *fiber.Ctx) error {
		typologyID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return err
		}
		var requestData []Type

		// Parse the request body into the requestData slice
		if err := c.BodyParser(&requestData); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Failed to parse request data")
		}

		for _, typeInGlobal := range types {
			var typeFound = false
			for _, typeFromRequest := range requestData {
				if typeInGlobal.ID == typeFromRequest.ID && typeInGlobal.TypologyID == typologyID {
					typeFound = true
					break
				} else {

				}
			}
			if !typeFound {
				if err := db.Where("type_id = ? AND typology_id = ?", typeInGlobal.ID, typologyID).Delete(&Type{}).Error; err != nil {
					log.Println("Error deleting type:", err)
					return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete type")
				}
			}

		}

		// Update, Delete or Add Types based on the logic
		for _, typeFromRequest := range requestData {
			// Check if the type_id is negative (add new type)
			if typeFromRequest.ID < 0 {
				// Add new type (assign typologyID)
				typeFromRequest.TypologyID = typologyID
				if err := db.Save(&Type{
					TypologyID:  typeFromRequest.TypologyID,
					Name:        typeFromRequest.Name,
					DisplayName: typeFromRequest.DisplayName,
					Description: typeFromRequest.Description,
				}).Error; err != nil {
					log.Println("Error adding new type:", err)
					return c.Status(fiber.StatusInternalServerError).SendString("Failed to add new type")
				}
				continue
			}

			// Check if type_id exists in the existing types array
			// var existingType Type
			typeFound := true

			if typeFound {
				// Update the existing type if it's in both the global array and the request data
				// if err := db.Model(&existingType).Updates(Type{
				// 	Name:        typeFromRequest.Name,
				// 	DisplayName: typeFromRequest.DisplayName,
				// 	Description: typeFromRequest.Description,
				// }).Error;
				if err := db.Save(&Type{
					ID:          typeFromRequest.ID,
					TypologyID:  typeFromRequest.TypologyID,
					Name:        typeFromRequest.Name,
					DisplayName: typeFromRequest.DisplayName,
					Description: typeFromRequest.Description,
				}).Error; err != nil {
					log.Println("Error savingtype:", err)
					return c.Status(fiber.StatusInternalServerError).SendString("Failed to add new type")
				}
			} else {
				// this probably isn't needed anymore
				log.Println(typeFromRequest.ID)
				// If type_id is in the global array but not in the request, delete it
				if err := db.Where("type_id = ? AND typology_id = ?", typeFromRequest.ID, typologyID).Delete(&Type{}).Error; err != nil {
					log.Println("Error deleting type:", err)
					return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete type")
				}
			}
		}

		// Return a success response
		return c.Status(fiber.StatusOK).SendString("Types processed successfully")
	})

	app.Post("/api/submitsubject", func(c *fiber.Ctx) error {
		// Create an instance of SubjectResponse
		var data SubjectResponse
		fmt.Println("reaction")

		// Parse the incoming JSON request body into the SubjectResponse struct
		if err := c.BodyParser(&data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
			})
		}

		for _, value := range data.Types {
			fmt.Println(value)
			var typologyID int

			for _, t := range types {
				if t.ID == value {
					typologyID = t.TypologyID
					break
				}
			}

			//db.Where("subject_id = ? AND typology_id = ?", data.SubjectID, typologyID).Delete(&SubjectType{})

			subjectType := SubjectType{
				SubjectID:  data.SubjectID,
				TypologyID: typologyID,
				TypeID:     value,
			}

			if err := db.Save(&subjectType).Error; err != nil {
				log.Fatalf("Error creating or updating SubjectType: %v", err)
			}

			fmt.Println(typologyID)
		}

		fmt.Println(data.Subject)

		return nil
	})

	app.Post(prefix+"/subject/delete/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return errors.New("id is empty")
		}

		var subject Subject
		if err := db.First(&subject, id).Error; err != nil {
			return err
		}
		db.Delete(&subject)
		return nil
	})

	app.Post(prefix+"/delete/typology/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return errors.New("id is empty")
		}

		var typology Typology
		if err := db.First(&typology, id).Error; err != nil {
			return err
		}
		db.Delete(&typology)
		return nil
	})

	app.Post(prefix+"/subject/add", func(c *fiber.Ctx) error {
		subject := new(Subject)
		if err := c.BodyParser(&subject); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
			})
		}
		db.Create(&subject)
		return nil
	})

	app.Post(prefix+"/typology/add", func(c *fiber.Ctx) error {
		typology := new(Typology)
		if err := c.BodyParser(&typology); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
			})
		}
		db.Create(&typology)
		return nil
	})

	app.Listen(":8000")
}
