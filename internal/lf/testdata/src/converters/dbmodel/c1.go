package dbmodel

type Sample struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Price    int64  `json:"price"`
	Currency string `json:"currency"`
}
