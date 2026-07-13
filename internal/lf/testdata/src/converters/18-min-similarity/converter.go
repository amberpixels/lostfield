package sample_min_similarity

// Message is a domain model.
type Message struct {
	ID   string
	Text string
}

// MessageNewParams is an API-call parameter struct: it shares the "message"
// substring with Message but is NOT a conversion target. With min-similarity
// configured (e.g. 0.6), the low bigram similarity between the names excludes
// this function from converter detection; with the default substring matching
// it would be flagged.
type MessageNewParams struct {
	ID          string
	Text        string
	MaxTokens   int
	Temperature float64
}

// callAndMeter resembles the false-positive class seen in real codebases:
// it takes params and returns a struct whose name shares a substring.
func callAndMeter(params MessageNewParams) Message {
	_ = params.ID
	return Message{}
}

// UserModel converts to UserDTO: names are similar enough (dice >= 0.6)
// to stay detected as a converter even with min-similarity configured.
type UserModel struct {
	ID   int64
	Name string
}

// UserModelDTO is the output model.
type UserModelDTO struct {
	ID   int64
	Name string
}

// ConvertUser is an incomplete converter that must still be reported
// when min-similarity is enabled.
func ConvertUser(u UserModel) UserModelDTO { // want "incomplete converter with missing fields: u.Name, Name"
	return UserModelDTO{
		ID: u.ID,
	}
}
