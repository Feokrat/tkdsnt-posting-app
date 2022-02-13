package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/Feokrat/tkdsnt-posting-app/internal/config"

	"github.com/Feokrat/tkdsnt-posting-app/pkg/database"

	"github.com/Feokrat/tkdsnt-posting-app/internal/post"
)

var configFile = "configs/config"

const SOURCES_FILE = "sources.txt"

func main() {
	fmt.Println("Hello world")
	logger := log.New(os.Stdout, "", 0)

	cfg, err := config.Init(configFile, logger)
	if err != nil {
		logger.Fatalf("failed to load application configuration: %s", err)
	}

	db, err := database.NewPostgresDB(cfg.Postgresql, logger)
	if err != nil {
		logger.Fatalf("error connecting to database %s", err.Error())
	}
	defer database.ClosePostgresDB(db)

	repo := post.NewRepository(db, logger)
	service := post.NewService(repo, logger)

	err = downloadPosts(service, SOURCES_FILE)
	if err != nil {
		logger.Fatalf("error reading source file %s, err: %s", SOURCES_FILE, err.Error())
		return
	}

	err =
}

func downloadPosts(service post.Service, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		_, err = service.DownloadPost(scanner.Text())
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
