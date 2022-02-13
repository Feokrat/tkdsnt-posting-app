package model

type Post struct {
	Id uint `json:"id"`

	Filenames []string `json:"paths"`
	SourceUrl string   `json:"source_url"`

	Posted bool `json:"posted"`
}
