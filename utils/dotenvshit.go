package utils

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
)

func Init() {

	envFilePath := filepath.Join("..", ".env") // Go one directory up to find the .env file
	//log.Println("Current directory:", envFilePath)

	if os.Getenv("ENV") != "production" {

		if err := godotenv.Load(); err != nil {
			if err2 := godotenv.Load(envFilePath); err2 != nil {
				log.Println("Error loading .env file:", err2)
				log.Println("No .env file found. Using environment variables.")
			}
		}
		log.Println("Development environment detected. Loading .env file.")
		log.Println(os.Getenv("TEMP_DIR"))
	} else {
		log.Println("Production environment detected. Skipping .env loading.")
	}
}
