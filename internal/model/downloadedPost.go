package model

type DownloadedPost struct {
	Id uint `json:"id" db:"id"`

	Filenames []string `json:"filenames" db:"filenames"`
	SourceUrl string   `json:"source_url" db:"source"`

	Posted bool `json:"posted" db:"posted"`
}
