package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/shopos/crm-integration-service/internal/adapter"
	"github.com/shopos/crm-integration-service/internal/domain"
	"github.com/shopos/crm-integration-service/internal/store"
)

// Servicer encapsulates CRM integration business logic.
type Servicer struct {
	store   *store.SyncStore
	adapter *adapter.CrmAdapter
}

// New constructs a Servicer.
func New(st *store.SyncStore, ad *adapter.CrmAdapter) *Servicer {
	return &Servicer{store: st, adapter: ad}
}

// SyncContacts pushes the provided customers to the given CRM and records the result.
func (s *Servicer) SyncContacts(crmSystem domain.CrmSystem, customers []map[string]interface{}) domain.SyncResult {
	result := domain.SyncResult{
		SyncID:     newID(),
		CrmSystem:  crmSystem,
		EntityType: "contact",
		Errors:     []string{},
	}

	synced := 0
	errs := []string{}

	for _, c := range customers {
		formatted := s.adapter.FormatContact(crmSystem, c)
		if formatted == nil || len(formatted) == 0 {
			errs = append(errs, fmt.Sprintf("customer %v could not be formatted", c["id"]))
			continue
		}

		// Simulate CRM API call — in production this would be an HTTP/gRPC call.
		crmID := fmt.Sprintf("%s-%s", crmSystem, newShortID())
		contact := s.adapter.ParseCrmContact(crmSystem, buildCrmResponse(crmSystem, crmID, c))
		contact.ShopOsCustomerID = fmt.Sprintf("%v", c["id"])
		s.store.SaveContact(&contact)
		synced++
	}

	result.ItemsSynced = synced
	result.Errors = errs
	result.CompletedAt = time.Now().UTC()
	s.store.SaveResult(&result)
	return result
}

// SyncDeals pushes orders to the given CRM as deals and records the result.
func (s *Servicer) SyncDeals(crmSystem domain.CrmSystem, orders []map[string]interface{}) domain.SyncResult {
	result := domain.SyncResult{
		SyncID:     newID(),
		CrmSystem:  crmSystem,
		EntityType: "deal",
		Errors:     []string{},
	}

	synced := 0
	errs := []string{}

	for _, o := range orders {
		if fmt.Sprintf("%v", o["id"]) == "" && fmt.Sprintf("%v", o["orderId"]) == "" {
			errs = append(errs, "order missing id")
			continue
		}

		// Simulate CRM deal creation.
		crmID := fmt.Sprintf("%s-DEAL-%s", crmSystem, newShortID())
		deal := s.adapter.ParseCrmDeal(crmSystem, buildDealCrmResponse(crmSystem, crmID, o))
		deal.ShopOsOrderID = fmt.Sprintf("%v", firstOf(o["id"], o["orderId"]))
		_ = deal // In production, persist or forward to order domain.
		synced++
	}

	result.ItemsSynced = synced
	result.Errors = errs
	result.CompletedAt = time.Now().UTC()
	s.store.SaveResult(&result)
	return result
}

// GetContact retrieves a contact by CRM system and CRM-side ID.
func (s *Servicer) GetContact(crmSystem domain.CrmSystem, crmID string) (*domain.CrmContact, error) {
	return s.store.GetContactByCrmId(crmSystem, crmID)
}

// ListContacts returns all contacts for a CRM system.
func (s *Servicer) ListContacts(crmSystem domain.CrmSystem) []*domain.CrmContact {
	return s.store.ListContacts(crmSystem)
}

// GetFieldMapping returns the field mapping for a CRM system.
func (s *Servicer) GetFieldMapping(crmSystem domain.CrmSystem) map[string]string {
	return s.adapter.GetFieldMapping(crmSystem)
}

// GetSyncHistory returns past sync results for a CRM system.
func (s *Servicer) GetSyncHistory(crmSystem domain.CrmSystem, limit int) []*domain.SyncResult {
	return s.store.ListResults(crmSystem, limit)
}

// --- helpers ---

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func newShortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func buildCrmResponse(crmSystem domain.CrmSystem, crmID string, c map[string]interface{}) map[string]interface{} {
	switch crmSystem {
	case domain.CrmSalesforce:
		return map[string]interface{}{
			"Id":        crmID,
			"FirstName": c["firstName"],
			"LastName":  c["lastName"],
			"Email":     c["email"],
			"Phone":     c["phone"],
		}
	case domain.CrmHubSpot:
		return map[string]interface{}{
			"id": crmID,
			"properties": map[string]interface{}{
				"firstname": c["firstName"],
				"lastname":  c["lastName"],
				"email":     c["email"],
				"phone":     c["phone"],
				"company":   c["company"],
			},
		}
	case domain.CrmZoho:
		return map[string]interface{}{
			"id":         crmID,
			"First_Name": c["firstName"],
			"Last_Name":  c["lastName"],
			"Email":      c["email"],
			"Phone":      c["phone"],
		}
	case domain.CrmPipedrive:
		return map[string]interface{}{
			"id":    crmID,
			"name":  fmt.Sprintf("%v %v", c["firstName"], c["lastName"]),
			"email": c["email"],
			"phone": c["phone"],
		}
	}
	return map[string]interface{}{"id": crmID}
}

func buildDealCrmResponse(crmSystem domain.CrmSystem, crmID string, o map[string]interface{}) map[string]interface{} {
	amount := 0.0
	if v, ok := o["total"].(float64); ok {
		amount = v
	}
	switch crmSystem {
	case domain.CrmSalesforce:
		return map[string]interface{}{
			"Id":              crmID,
			"StageName":       "Closed Won",
			"Amount":          amount,
			"CurrencyIsoCode": "USD",
		}
	case domain.CrmHubSpot:
		return map[string]interface{}{
			"id": crmID,
			"properties": map[string]interface{}{
				"dealstage": "closedwon",
				"amount":    amount,
			},
		}
	case domain.CrmZoho:
		return map[string]interface{}{
			"id":       crmID,
			"Stage":    "Closed Won",
			"Amount":   amount,
			"Currency": "USD",
		}
	case domain.CrmPipedrive:
		return map[string]interface{}{
			"id":       crmID,
			"stage_id": "won",
			"value":    amount,
			"currency": "USD",
		}
	}
	return map[string]interface{}{"id": crmID}
}

func firstOf(vals ...interface{}) interface{} {
	for _, v := range vals {
		if v != nil && fmt.Sprintf("%v", v) != "" {
			return v
		}
	}
	return nil
}
