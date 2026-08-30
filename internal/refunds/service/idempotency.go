package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func refundRequestHash(orderID int, input refunds.Input) string {
	ticketIDs := append([]int(nil), input.TicketIDs...)
	sort.Ints(ticketIDs)
	payload, _ := json.Marshal(struct {
		OrderID   int   `json:"order_id"`
		All       bool  `json:"all"`
		TicketIDs []int `json:"ticket_ids"`
	}{OrderID: orderID, All: input.All, TicketIDs: ticketIDs})
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])
}
