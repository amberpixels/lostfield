package delegate

// ConvertTicketModelToProto converts a single Ticket to TicketProto
// This is the inner converter that handles the actual field mapping
func ConvertTicketModelToProto(t Ticket) *TicketProto {
	splits := make([]string, 0, len(t.ValidSplits)+len(t.IgnoredSplits))
	splits = append(splits, t.ValidSplits...)
	splits = append(splits, t.IgnoredSplits...)

	return &TicketProto{
		SourceRef:            t.ID,
		Section:              t.Section,
		Row:                  t.Row,
		IsOwnTickets:         t.IsOwnTickets,
		Format:               t.Format,
		Currency:             t.Currency,
		Service:              t.Service,
		Note:                 t.Note,
		Description:          t.Description,
		ShippingMethod:       t.ShippingMethod,
		TicketsCount:         t.TicketsCount,
		Price:                t.Price,
		IsLeavingSingleSeats: t.IsLeavingSingleSeats,
		Category:             t.Category,
		Splits:               splits,
		Seats:                t.Seats,
		ServiceInfo:          t.ServiceInfo,
		SeatsWarnings:        t.SeatsWarnings,
		RewriteDetails:       t.RewriteDetails,
		Display:              t.Display,
	}
}

// ConvertTicketsModelsToProto converts a slice of Tickets to a slice of TicketProtos using append
// This is the delegating converter that calls ConvertTicketModelToProto for each element
// The linter should recognize this pattern and skip validation, since the actual field
// mapping is delegated to ConvertTicketModelToProto which will be linted separately
func ConvertTicketsModelsToProto(tickets []Ticket) []*TicketProto {
	res := make([]*TicketProto, 0, len(tickets))

	for _, t := range tickets {
		res = append(res, ConvertTicketModelToProto(t))
	}

	return res
}

// ConvertTicketsModelsToProtoIndexed converts a slice of Tickets to a slice of TicketProtos using indexed assignment
// This variant uses indexed assignment (protos[i] = ...) instead of append, which is more efficient
// The linter should recognize this pattern and skip validation
func ConvertTicketsModelsToProtoIndexed(tickets []Ticket) []*TicketProto {
	protos := make([]*TicketProto, len(tickets))
	for i, t := range tickets {
		protos[i] = ConvertTicketModelToProto(t)
	}
	return protos
}
