package downloadPost

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/Feokrat/tkdsnt-posting-app/internal/model"

	url2 "net/url"

	twitterscraper "github.com/n0madic/twitter-scraper"
)

const DOWNLOAD_PATH = "pics"

type Service interface {
	DownloadPost(sourceUrl string) ([]string, error)
	Post() error
}

type service struct {
	repo   Repository
	logger *log.Logger
}

func NewService(repo Repository, logger *log.Logger) Service {
	return service{repo, logger}
}

func (s service) Post() error {
	return nil
}

func (s service) DownloadPost(sourceUrl string) ([]string, error) {
	filenames, err := downloadTweet(sourceUrl)
	if err != nil {
		s.logger.Printf("error downloading post %s", sourceUrl)
		return nil, err
	}

	post := model.DownloadedPost{
		Filenames: filenames,
		SourceUrl: sourceUrl,
		Posted:    false,
	}

	_, err = s.repo.Create(post)
	if err != nil {
		s.logger.Printf("error creating record in database about post %s", sourceUrl)
		return nil, err
	}

	return filenames, err
}

func downloadTweet(tweetLink string) ([]string, error) {
	url, err := url2.Parse(tweetLink)
	if err != nil {
		return nil, err
	}
	tweetId := path.Base(url.Path)

	scraper := twitterscraper.New()
	tweet, err := scraper.GetTweet(tweetId)
	if err != nil {
		return nil, err
	}
	photos := tweet.Photos

	log.Printf("%d photos", len(photos))

	if _, err := os.Stat(DOWNLOAD_PATH); os.IsNotExist(err) {
		err := os.Mkdir(DOWNLOAD_PATH, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	var filenames []string
	for i, img := range photos {
		fileName := fmt.Sprintf("%s/%s_%s_%d.%s", DOWNLOAD_PATH, tweet.Username, tweet.ID, i, photos[0][len(photos[0])-3:])
		err := downloadFile(img, fileName)
		if err != nil {
			return nil, err
		}
		log.Printf("Downloaded image: %s\n", fileName)
		filenames = append(filenames, fileName)
	}
	return filenames, nil
}

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)

	if response.StatusCode != 200 {
		return errors.New("received non 200 response code")
	}
	//Create empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	//Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
