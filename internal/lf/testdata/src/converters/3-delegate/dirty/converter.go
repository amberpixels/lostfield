package sample_delegate

import (
	sampleDelegate "converters/3-delegate/models"
)

// ConvertTicketModelToProto_Incomplete is missing several fields in the conversion
func ConvertTicketModelToProto_Incomplete(t sampleDelegate.Ticket) *sampleDelegate.TicketProto { // want "incomplete converter"
	splits := make([]string, 0, len(t.ValidSplits)+len(t.IgnoredSplits))
	splits = append(splits, t.ValidSplits...)
	splits = append(splits, t.IgnoredSplits...)

	return &sampleDelegate.TicketProto{
		SourceRef:      t.ID,
		Section:        t.Section,
		Row:            t.Row,
		Format:         t.Format,
		Currency:       t.Currency,
		Service:        t.Service,
		Note:           t.Note,
		Description:    t.Description,
		ShippingMethod: t.ShippingMethod,
		Price:          t.Price,
		Category:       t.Category,
		Splits:         splits,
		Seats:          t.Seats,
		// Missing: IsOwnTickets, TicketsCount, IsLeavingSingleSeats, ServiceInfo, SeatsWarnings, RewriteDetails, Display
	}
}

// ConvertTicketsModelsToProtoWithIncomplete delegates to the incomplete converter
func ConvertTicketsModelsToProtoWithIncomplete(tickets []sampleDelegate.Ticket) []*sampleDelegate.TicketProto {
	res := make([]*sampleDelegate.TicketProto, 0, len(tickets))

	for _, t := range tickets {
		res = append(res, ConvertTicketModelToProto_Incomplete(t))
	}

	return res
}
