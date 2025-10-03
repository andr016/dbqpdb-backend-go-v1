package main

// todo refactor lol

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"dbqpdb-backend-go-v1/auth"
	"dbqpdb-backend-go-v1/config"
	"dbqpdb-backend-go-v1/folder"
	"dbqpdb-backend-go-v1/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Flags
var setupFlag = flag.Bool("setup", false, "Set up the database with migrations")
var noauthFlag = flag.Bool("noauth", false, "Disable authentication")

// DB GORM
var db *gorm.DB

// Config
var conf, err = config.LoadConfig("config/config.json")

func initDB() {
	var err error

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		conf.DB.Host, conf.DB.User, conf.DB.Password, conf.DB.DBName, conf.DB.Port, conf.DB.SSLMode, conf.DB.TimeZone)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		//Logger:      logger.Default.LogMode(logger.Silent),
		PrepareStmt: true,
	})

	if err != nil {
		// Error handling informing user on what to do
		reason := ""
		switch {
		case strings.Contains(err.Error(), "auth"):
			reason = "wrong login information (password?)"
		case strings.Contains(err.Error(), "SQLSTATE 3D000"):
			if *setupFlag {
				fmt.Printf("Database does not exist. Creating the database by name %s...\nNote: this uses exec, make sure there is no sql injection in dbname if that's appropriate", conf.DB.DBName)
				dsnnodb := fmt.Sprintf("host=%s user=%s password=%s port=%d sslmode=%s TimeZone=%s",
					conf.DB.Host, conf.DB.User, conf.DB.Password, conf.DB.Port, conf.DB.SSLMode, conf.DB.TimeZone)
				db, err = gorm.Open(postgres.Open(dsnnodb), &gorm.Config{
					//Logger:      logger.Default.LogMode(logger.Silent),
					PrepareStmt: true,
				})
				if err != nil {
					log.Fatal(err.Error())
				}
				db.Exec(fmt.Sprintf("CREATE DATABASE %s", conf.DB.DBName))
				fmt.Println("Database created successfully!")
				db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
					//Logger:      logger.Default.LogMode(logger.Silent),
					PrepareStmt: true,
				})
				err := db.AutoMigrate(
					&models.Subject{},
					&models.Typology{},
					&models.Type{},
					&models.TypeForSubject{},
					&models.SubjectType{},
					&models.User{})
				if err != nil {
					log.Fatal(err.Error())
				}
				return
			}
			reason = "database doesn't exist. To automatically setup database, run with --setup flag"
		}

		if reason != "" {
			log.Fatalf("Failed to connect to database.\nProbably, %s.\n%v\n", reason, err)
		} else {
			log.Fatalf("Failed to connect to database.\n%v\n", err)
		}
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
	var types []models.Type
	if err := db.Where("typology_id = ?", parsedTypologyID).Find(&types).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Types not found for the given typology"})
	}

	// Return the types in the response
	return c.JSON(types)
}

func fetchGetSubjectsByGroup(c *fiber.Ctx) error {
	groupID := c.Params("id")
	var subjects []models.Subject
	if err := db.Where("group_id = ?", groupID).Find(&subjects).Error; err != nil {
		// If there's an error, return a 500 Internal Server Error response
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch types",
			"details": err.Error(),
		})
	}

	return c.JSON(subjects)
}

func fetchGetSubjects(c *fiber.Ctx) error {
	var subjects []models.Subject
	if err := db.Find(&subjects).Error; err != nil {
		// If there's an error, return a 500 Internal Server Error response
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch types",
			"details": err.Error(),
		})
	}

	return c.JSON(subjects)
}

var types []models.Type

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

var typologies []models.Typology

var groups []models.Group

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

func fetchGroups() error {
	if err := db.Find(&groups).Error; err != nil {
		return err
	}
	return nil
}

func getGroups(c *fiber.Ctx) error {
	if err := fetchGroups(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch typologies",
			"details": err.Error(),
		})
	}
	return c.JSON(groups)
}

func getTypologyByID(c *fiber.Ctx) error {
	typologyID := c.Params("id")

	var types []models.Type
	if err := db.Where("id = ?", typologyID).Find(&types).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Typology not found"})
	}
	return c.JSON(types)
}

func getSubjectByID(c *fiber.Ctx) error {
	subjectID := c.Params("id")

	var subject models.Subject

	// Eager loading related data: subject_types
	if err := db.Preload("SubjectTypes").
		Preload("SubjectTypes.Type").
		First(&subject, subjectID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Subject not found"})
	}

	// Prepare the response structure
	subjectResponse := struct {
		Subject models.Subject `json:"subject"`
		Types   []int          `json:"types"`
	}{
		Subject: subject,
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

func subjectsMostPopularFirstLetter() {
	// fetching
	var subjects []models.Subject
	if err := db.Find(&subjects).Error; err != nil {
		// If there's an error, return a 500 Internal Server Error response
	}

	var wg sync.WaitGroup

	for _, subject := range subjects {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(subject.Name)
		}()
	}
	wg.Wait()
	log.Println("Goroutines finished")
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

	var subject models.Subject
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

func groupAdd(c *fiber.Ctx) error {
	group := new(models.Group)
	if err := c.BodyParser(&group); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}
	db.Create(&group)
	return nil
}

func groupDelete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errors.New("id is empty")
	}

	var group models.Group
	if err := db.First(&group, id).Error; err != nil {
		return err
	}
	db.Delete(&group)
	return nil
}

func updateSubject(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID is required"})
	}

	// 2. Parse request body into a Subject struct
	var updatedSubject models.Subject
	if err := c.BodyParser(&updatedSubject); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 3. Update in database (example using GORM)
	result := db.Model(&models.Subject{}).Where("subject_id = ?", id).Updates(updatedSubject)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update Subject"})
	}

	// 4. Return success
	return c.JSON(fiber.Map{
		"success": true,
		"data":    updatedSubject,
	})
}

func submitTypology(c *fiber.Ctx) error {
	typologyID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	var requestData []models.Type

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
			if err := db.Where("type_id = ? AND typology_id = ?", typeInGlobal.ID, typologyID).Delete(&models.Type{}).Error; err != nil {
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
			if err := db.Save(&models.Type{
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
			if err := db.Save(&models.Type{
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
			if err := db.Where("type_id = ? AND typology_id = ?", typeFromRequest.ID, typologyID).Delete(&models.Type{}).Error; err != nil {
				log.Println("Error deleting type:", err)
				return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete type")
			}
		}
	}

	// Return a success response
	return c.Status(fiber.StatusOK).SendString("Types processed successfully")
}

func submitSubject(c *fiber.Ctx) error {
	// Create an instance of SubjectResponse
	var data models.SubjectResponse
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

		subjectType := models.SubjectType{
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
}

func deleteSubject(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errors.New("id is empty")
	}

	var subject models.Subject
	if err := db.First(&subject, id).Error; err != nil {
		return err
	}
	db.Delete(&subject)
	return nil
}

func createSubject(c *fiber.Ctx) error {
	subject := new(models.Subject)
	if err := c.BodyParser(&subject); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}
	db.Create(&subject)
	return nil
}

func deleteTypology(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errors.New("id is empty")
	}

	var typology models.Typology
	if err := db.First(&typology, id).Error; err != nil {
		return err
	}
	db.Delete(&typology)
	return nil
}

func createTypology(c *fiber.Ctx) error {
	typology := new(models.Typology)
	if err := c.BodyParser(&typology); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}
	db.Create(&typology)
	return nil
}

func main() {
	flag.Parse()
	conf, err = config.LoadConfig("config/config.json")

	// REFACTOR!!!!
	folderPath := "uploads"

	// Call the function to create the folder if it doesn't exist
	err := folder.CreateFolderIfNotExist(folderPath)
	if err != nil {
		fmt.Println("Error:", err)
	}

	const prefix = "/api"

	if err != nil {
		log.Fatal("Error loading config: ", err)
	}
	initDB()
	fetchTypes()
	fetchTypologies()

	if *setupFlag {
		if err != nil {
			log.Fatalf("Failed to set up the database: %v", err)
		}
		fmt.Println("Database setup complete!")
		return
	}

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

	// AUTH TEST
	// Login route
	app.Post("/login", auth.Login)

	// Unauthenticated route
	app.Get("/", auth.Accessible)

	api := app.Group("/api")

	if *noauthFlag {
		fmt.Println("WARNING! Authentication disabled.")
	} else {
		// JWT Middleware
		app.Use(jwtware.New(jwtware.Config{
			SigningKey: jwtware.SigningKey{Key: []byte("secret")},
		}))

		// Restricted Routes
		app.Get("/restricted", auth.Restricted)
	}

	api.Get("/groups", getGroups)
	api.Post("/groups", groupAdd)
	api.Delete("/groups/:id", groupDelete)

	api.Get("/subject/:id", getSubjectByID)
	api.Get("/typology", getTypologies)
	api.Get("/typology/:typology_id", getTypesByTypologyID)
	api.Get("/types", getTypes)
	api.Get("/subject/", fetchGetSubjects)
	api.Get("/subject/group/:id", fetchGetSubjectsByGroup)

	api.Post("/upload/subject/:id", uploadImage)
	api.Put("/subject/:id", updateSubject)

	api.Post("/submit/typology/:id", submitTypology)
	api.Post("/submitsubject", submitSubject)
	api.Post("/delete/subject/:id", deleteSubject)
	api.Post("/delete/typology/:id", deleteTypology)
	api.Post("/subject/add", createSubject)
	api.Post("/typology/add", createTypology)

	go subjectsMostPopularFirstLetter()

	app.Listen(":8000")
}
