package vkPost

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

const (
	BASE_VK_API_URL = "https://api.vk.com/method"
	VK_API_VERSION  = "5.131"
)

var MAX_UPLOAD_SIZE int64 = 10 << 20

type VkSaveWallPhotoResponseType struct {
	Response []VkSaveWallPhotoResponse `json:"response"`
}

type VkSaveWallPhotoResponse struct {
	Id      int `json:"id"`
	OwnerId int `json:"owner_id"`
}

type VkGetWallUploadServerResponseType struct {
	Response VkGetWallUploadServerResponse `json:"response"`
}

type VkGetWallUploadServerResponse struct {
	AlbumId   int    `json:"album_id"`
	UploadUrl string `json:"upload_url"`
	UserId    int    `json:"user_id"`
}

type VkPhotoUploadResponse struct {
	Server int    `json:"server"`
	Photo  string `json:"photo"`
	Hash   string `json:"hash"`
}

type VkPostService struct {
	logger      *log.Logger
	groupId     int
	accessToken string
}

func NewVkPostService(groupId int, accessToken string, logger *log.Logger) *VkPostService {
	return &VkPostService{
		logger:      logger,
		groupId:     groupId,
		accessToken: accessToken,
	}
}

func (s VkPostService) MakePost(filenames []string, sourceUrl string) error {
	var attachments []string
	for _, filename := range filenames {
		uploadURL, err := s.getPostingUrl()
		if err != nil {
			s.logger.Printf("error getting upload url, err: %s", err)
			return err
		}

		uploaded, err := s.uploadPhoto(uploadURL, filename)
		if err != nil {
			s.logger.Printf("error uploading photo, err: %s", err)
			return err
		}

		mediaId, err := s.saveWallPhoto(uploaded)
		if err != nil {
			s.logger.Printf("error saving wall photo, err: %s", err)
			return err
		}
		attachments = append(attachments, mediaId)
	}

	err := s.wallPost(attachments, sourceUrl)
	if err != nil {
		s.logger.Printf("error posting to the wall, err: %s", err)
		return err
	}

	s.logger.Printf("posted %s\n", sourceUrl)
	return nil
}

func (s VkPostService) wallPost(attachments []string, sourceUrl string) error {
	methodName := "wall.post"

	requestString := fmt.Sprintf("%s/%s?access_token=%s&owner_id=-%d&from_group=1&attachments=%s&copyright=%s&v=%s",
		BASE_VK_API_URL, methodName, s.accessToken, s.groupId, strings.Join(attachments, ","), sourceUrl, VK_API_VERSION)

	r, err := http.Get(requestString)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	return nil
}

func (s VkPostService) saveWallPhoto(uploaded VkPhotoUploadResponse) (string, error) {
	var response VkSaveWallPhotoResponseType
	methodName := "photos.saveWallPhoto"

	requestString := fmt.Sprintf("%s/%s?access_token=%s&&group_id=%d&server=%d&photo=%s&hash=%s&v=%s",
		BASE_VK_API_URL, methodName, s.accessToken, s.groupId, uploaded.Server, uploaded.Photo, uploaded.Hash, VK_API_VERSION)

	r, err := http.Get(requestString)
	if err != nil {
		return "", err
	}

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	s.logger.Printf("media_id: photo%d_%d", response.Response[0].OwnerId, response.Response[0].Id)

	return fmt.Sprintf("photo%d_%d", response.Response[0].OwnerId, response.Response[0].Id), nil
}

func (s VkPostService) uploadPhoto(uploadUrl string, filename string) (VkPhotoUploadResponse, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("photo", filename)
	if err != nil {
		return VkPhotoUploadResponse{}, err
	}

	fh, err := os.Open(filename)
	if err != nil {
		return VkPhotoUploadResponse{}, err
	}
	defer fh.Close()

	fi, err := fh.Stat()
	if err != nil {
		return VkPhotoUploadResponse{}, err
	}

	if fi.Size() > MAX_UPLOAD_SIZE {
		if err != nil {
			return VkPhotoUploadResponse{}, errors.New("file size is bigger than 10MB.")
		}
	}

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return VkPhotoUploadResponse{}, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(uploadUrl, contentType, bodyBuf)
	if err != nil {
		return VkPhotoUploadResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return VkPhotoUploadResponse{}, err
	}

	var uploaded VkPhotoUploadResponse
	err = json.Unmarshal(body, &uploaded)
	if err != nil {
		return VkPhotoUploadResponse{}, err
	}

	return uploaded, nil
}

func (s VkPostService) getPostingUrl() (string, error) {
	var response VkGetWallUploadServerResponseType
	methodName := "photos.getWallUploadServer"

	requestString := fmt.Sprintf("%s/%s?access_token=%s&group_id=%d&v=%s",
		BASE_VK_API_URL, methodName, s.accessToken, s.groupId, VK_API_VERSION)

	r, err := http.Get(requestString)
	if err != nil {
		return "", err
	}

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.Response.UploadUrl, nil
}
