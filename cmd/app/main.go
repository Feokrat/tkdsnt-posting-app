package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"time"

	"github.com/Feokrat/tkdsnt-posting-app/internal/vkPost"

	"github.com/Feokrat/tkdsnt-posting-app/internal/config"
	"github.com/Feokrat/tkdsnt-posting-app/pkg/database"

	"github.com/Feokrat/tkdsnt-posting-app/internal/downloadPost"
)

var configFile = "configs/config"

const SOURCES_FILE = "sources.txt"

func main() {
	postFlag := flag.Bool("p", false, "post downloaded")
	downloadFlag := flag.Bool("d", false, "download from sources")
	flag.Parse()

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

	repo := downloadPost.NewRepository(db, logger)

	if *downloadFlag {
		err = download(repo, cfg, logger)
		if err != nil {
			return
		}
	}

	if *postFlag {
		err = postDownloaded(repo, cfg, logger)
		if err != nil {
			return
		}
	}
}

func download(repo downloadPost.Repository, cfg *config.Config, logger *log.Logger) error {
	service := downloadPost.NewService(repo, cfg, logger)

	err := downloadPosts(service, SOURCES_FILE)
	if err != nil {
		logger.Fatalf("error reading source file %s, err: %s", SOURCES_FILE, err.Error())
		return err
	}

	return nil
}

func postDownloaded(repo downloadPost.Repository, cfg *config.Config, logger *log.Logger) error {
	vkPostService := vkPost.NewVkPostService(cfg.Posting.GroupId, cfg.Posting.AccessToken, logger)

	for {
		downloaded, err := repo.GetLastUnposted()
		if err != nil {
			logger.Fatalf("error getting last not posted, err: %s", err.Error())
			return err
		}

		if downloaded.Filenames == nil {
			break
		}

		err = vkPostService.MakePost(downloaded.Filenames, downloaded.SourceUrl)
		if err != nil {
			logger.Fatalf("error posting, err: %s", err.Error())
			return err
		}

		time.Sleep(1 * time.Second)

		err = repo.SetPosted(downloaded.Id)
		if err != nil {
			logger.Fatalf("error making posted, err: %s", err.Error())
			return err
		}
	}

	return nil
}

func downloadPosts(service downloadPost.Service, filename string) error {
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
