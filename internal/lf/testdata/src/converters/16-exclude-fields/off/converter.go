package sample_exclude_fields_off

import "time"

// Article is the input model with typical DB-managed timestamp fields.
type Article struct {
	ID        int64
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ArticleDTO is the output model.
type ArticleDTO struct {
	ID    int64
	Title string
	Body  string
}

// ConvertArticle skips CreatedAt/UpdatedAt. Without exclude-fields configured,
// the skipped timestamps are reported as missing.
func ConvertArticle(a Article) ArticleDTO { // want "incomplete converter with missing fields: a.CreatedAt, a.UpdatedAt"
	return ArticleDTO{
		ID:    a.ID,
		Title: a.Title,
		Body:  a.Body,
	}
}
