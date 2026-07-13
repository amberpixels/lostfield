package sample_exclude_fields_on

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

// ConvertArticle skips CreatedAt/UpdatedAt. This package is exercised with
// exclude-fields="CreatedAt,UpdatedAt", so no diagnostics are expected.
func ConvertArticle(a Article) ArticleDTO {
	return ArticleDTO{
		ID:    a.ID,
		Title: a.Title,
		Body:  a.Body,
	}
}

// Event has nested-path candidates for exclusion.
type Event struct {
	Name string
	Meta Meta
}

// Meta is a nested struct whose Internal field is excluded by full path (Meta.Internal).
type Meta struct {
	Source   string
	Internal string
}

// EventDTO is the output model.
type EventDTO struct {
	Name string
	Meta Meta
}

// ConvertEvent maps Meta partially: Meta.Internal is excluded via the full-path
// pattern "Meta\\.Internal", so only Meta.Source must be handled.
func ConvertEvent(e Event) EventDTO {
	return EventDTO{
		Name: e.Name,
		Meta: Meta{
			Source: e.Meta.Source,
		},
	}
}
