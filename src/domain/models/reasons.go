package models

import pb "gitlab.faza.io/protos/order"

// Reason Holds the reason for an Action like user cancelling an order
type Reason struct {
	Key         string
	Translation string
	Description string
	Cancel      bool
	Return      bool
	Responsible ReasonResponsible
}

func (r Reason) ToRPC() (p *pb.Reason) {
	p = &pb.Reason{
		Key:         r.Key,
		Description: r.Description,
	}
	return
}

// ReasonResponsible type to indicate who is responsible for a particular action
// for example if product is retuned because it is broken seller is responsible
type ReasonResponsible string

// all values for ReasonResponsible
const (
	ReasonResponsibleBuyer  = "BUYER"
	ReasonResponsibleSeller = "SELLER"
	ReasonResponsibleNone   = "NONE"
)

// ReasonConfig holds all the details of an action reason
type ReasonConfig struct {
	Key            string
	Translation    string
	HasDescription bool
	Cancel         bool
	Return         bool
	IsActive       bool
	Responsible    ReasonResponsible
}
