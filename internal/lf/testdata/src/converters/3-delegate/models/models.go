package modelsDelegate

// Ticket represents a single ticket in the input model
type Ticket struct {
	ID                   string
	Currency             string
	Format               string
	Price                int64
	ShippingMethod       string
	TicketsCount         int
	Section              string
	Category             string
	Row                  int
	Service              string
	Note                 string
	Description          string
	Display              string
	ValidSplits          []string
	IgnoredSplits        []string
	Seats                []string
	ServiceInfo          string
	IsLeavingSingleSeats bool
	IsOwnTickets         bool
	SeatsWarnings        []string
	RewriteDetails       string
}

// TicketProto represents a single ticket in the output protobuf model
type TicketProto struct {
	SourceRef            string
	Section              string
	Row                  int
	IsOwnTickets         bool
	Format               string
	Currency             string
	Service              string
	Note                 string
	Description          string
	ShippingMethod       string
	TicketsCount         int
	Price                int64
	IsLeavingSingleSeats bool
	Category             string
	Splits               []string
	Seats                []string
	ServiceInfo          string
	SeatsWarnings        []string
	RewriteDetails       string
	Display              string
}
