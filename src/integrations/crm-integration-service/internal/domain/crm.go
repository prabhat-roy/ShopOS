package domain

import "time"

// CrmSystem represents a supported CRM platform.
type CrmSystem string

const (
	CrmSalesforce CrmSystem = "SALESFORCE"
	CrmHubSpot    CrmSystem = "HUBSPOT"
	CrmZoho       CrmSystem = "ZOHO"
	CrmPipedrive  CrmSystem = "PIPEDRIVE"
)

// ValidCrmSystem returns true when c is a recognised CRM system.
func ValidCrmSystem(c CrmSystem) bool {
	switch c {
	case CrmSalesforce, CrmHubSpot, CrmZoho, CrmPipedrive:
		return true
	}
	return false
}

// CrmContact is a customer record that has been synced to a CRM.
type CrmContact struct {
	CrmID           string    `json:"crmId"`
	ShopOsCustomerID string   `json:"shopOsCustomerId"`
	Email           string    `json:"email"`
	FirstName       string    `json:"firstName"`
	LastName        string    `json:"lastName"`
	Phone           string    `json:"phone"`
	Company         string    `json:"company"`
	CrmSystem       CrmSystem `json:"crmSystem"`
	SyncedAt        time.Time `json:"syncedAt"`
}

// CrmDeal is an order that has been synced to a CRM as a deal/opportunity.
type CrmDeal struct {
	CrmID        string    `json:"crmId"`
	ShopOsOrderID string   `json:"shopOsOrderId"`
	CrmSystem    CrmSystem `json:"crmSystem"`
	Stage        string    `json:"stage"`
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	CloseDate    time.Time `json:"closeDate"`
	SyncedAt     time.Time `json:"syncedAt"`
}

// SyncResult captures the outcome of a single CRM sync run.
type SyncResult struct {
	SyncID      string    `json:"syncId"`
	CrmSystem   CrmSystem `json:"crmSystem"`
	EntityType  string    `json:"entityType"`
	ItemsSynced int       `json:"itemsSynced"`
	Errors      []string  `json:"errors"`
	CompletedAt time.Time `json:"completedAt"`
}
