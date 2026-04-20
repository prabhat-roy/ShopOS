package adapter

import (
	"fmt"
	"time"

	"github.com/shopos/crm-integration-service/internal/domain"
)

// CrmAdapter translates between ShopOS internal structures and
// CRM-system-specific field layouts.
type CrmAdapter struct{}

// New returns a new CrmAdapter.
func New() *CrmAdapter {
	return &CrmAdapter{}
}

// FormatContactForSalesforce converts a generic customer map to Salesforce Contact fields.
func (a *CrmAdapter) FormatContactForSalesforce(customer map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"LastName":    stringVal(customer, "lastName"),
		"FirstName":   stringVal(customer, "firstName"),
		"Email":       stringVal(customer, "email"),
		"Phone":       stringVal(customer, "phone"),
		"AccountId":   stringVal(customer, "companyId"),
		"MobilePhone": stringVal(customer, "mobile"),
		"MailingCity": stringVal(customer, "city"),
		"Description": fmt.Sprintf("ShopOS Customer ID: %v", customer["id"]),
	}
}

// FormatContactForHubSpot converts a generic customer map to HubSpot contact properties.
func (a *CrmAdapter) FormatContactForHubSpot(customer map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"properties": map[string]interface{}{
			"lastname":    stringVal(customer, "lastName"),
			"firstname":   stringVal(customer, "firstName"),
			"email":       stringVal(customer, "email"),
			"phone":       stringVal(customer, "phone"),
			"company":     stringVal(customer, "company"),
			"city":        stringVal(customer, "city"),
			"shopos_id":   stringVal(customer, "id"),
		},
	}
}

// FormatContactForZoho converts a generic customer map to Zoho CRM Contact fields.
func (a *CrmAdapter) FormatContactForZoho(customer map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"Last_Name":  stringVal(customer, "lastName"),
		"First_Name": stringVal(customer, "firstName"),
		"Email":      stringVal(customer, "email"),
		"Phone":      stringVal(customer, "phone"),
		"Account_Name": map[string]interface{}{
			"name": stringVal(customer, "company"),
		},
		"Mobile":      stringVal(customer, "mobile"),
		"Description": fmt.Sprintf("ShopOS Customer ID: %v", customer["id"]),
	}
}

// FormatContactForPipedrive converts a generic customer map to Pipedrive person fields.
func (a *CrmAdapter) FormatContactForPipedrive(customer map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":  fmt.Sprintf("%v %v", customer["firstName"], customer["lastName"]),
		"email": []map[string]interface{}{{"value": stringVal(customer, "email"), "primary": true}},
		"phone": []map[string]interface{}{{"value": stringVal(customer, "phone"), "primary": true}},
		"org_id": stringVal(customer, "companyId"),
	}
}

// ParseCrmContact converts a raw CRM contact payload to a domain CrmContact.
func (a *CrmAdapter) ParseCrmContact(crmSystem domain.CrmSystem, data map[string]interface{}) domain.CrmContact {
	contact := domain.CrmContact{
		CrmSystem: crmSystem,
		SyncedAt:  time.Now().UTC(),
	}

	switch crmSystem {
	case domain.CrmSalesforce:
		contact.CrmID = stringVal(data, "Id")
		contact.FirstName = stringVal(data, "FirstName")
		contact.LastName = stringVal(data, "LastName")
		contact.Email = stringVal(data, "Email")
		contact.Phone = stringVal(data, "Phone")
		contact.Company = stringVal(data, "AccountId")

	case domain.CrmHubSpot:
		contact.CrmID = stringVal(data, "id")
		if props, ok := data["properties"].(map[string]interface{}); ok {
			contact.FirstName = stringVal(props, "firstname")
			contact.LastName = stringVal(props, "lastname")
			contact.Email = stringVal(props, "email")
			contact.Phone = stringVal(props, "phone")
			contact.Company = stringVal(props, "company")
		}

	case domain.CrmZoho:
		contact.CrmID = stringVal(data, "id")
		contact.FirstName = stringVal(data, "First_Name")
		contact.LastName = stringVal(data, "Last_Name")
		contact.Email = stringVal(data, "Email")
		contact.Phone = stringVal(data, "Phone")
		if acc, ok := data["Account_Name"].(map[string]interface{}); ok {
			contact.Company = stringVal(acc, "name")
		}

	case domain.CrmPipedrive:
		contact.CrmID = fmt.Sprintf("%v", data["id"])
		contact.Email = stringVal(data, "email")
		contact.Phone = stringVal(data, "phone")
		if name := stringVal(data, "name"); name != "" {
			contact.FirstName = name
		}
	}

	return contact
}

// ParseCrmDeal converts a raw CRM deal payload to a domain CrmDeal.
func (a *CrmAdapter) ParseCrmDeal(crmSystem domain.CrmSystem, data map[string]interface{}) domain.CrmDeal {
	deal := domain.CrmDeal{
		CrmSystem: crmSystem,
		SyncedAt:  time.Now().UTC(),
		Currency:  "USD",
	}

	switch crmSystem {
	case domain.CrmSalesforce:
		deal.CrmID = stringVal(data, "Id")
		deal.Stage = stringVal(data, "StageName")
		deal.Amount = floatVal(data, "Amount")
		deal.Currency = stringVal(data, "CurrencyIsoCode")

	case domain.CrmHubSpot:
		deal.CrmID = stringVal(data, "id")
		if props, ok := data["properties"].(map[string]interface{}); ok {
			deal.Stage = stringVal(props, "dealstage")
			deal.Amount = floatVal(props, "amount")
		}

	case domain.CrmZoho:
		deal.CrmID = stringVal(data, "id")
		deal.Stage = stringVal(data, "Stage")
		deal.Amount = floatVal(data, "Amount")
		deal.Currency = stringVal(data, "Currency")

	case domain.CrmPipedrive:
		deal.CrmID = fmt.Sprintf("%v", data["id"])
		deal.Stage = stringVal(data, "stage_id")
		deal.Amount = floatVal(data, "value")
		deal.Currency = stringVal(data, "currency")
	}

	if deal.CloseDate.IsZero() {
		deal.CloseDate = time.Now().UTC().AddDate(0, 1, 0)
	}
	return deal
}

// GetFieldMapping returns the canonical ShopOS field → CRM field mapping.
func (a *CrmAdapter) GetFieldMapping(crmSystem domain.CrmSystem) map[string]string {
	switch crmSystem {
	case domain.CrmSalesforce:
		return map[string]string{
			"lastName":  "LastName",
			"firstName": "FirstName",
			"email":     "Email",
			"phone":     "Phone",
			"company":   "AccountId",
			"orderId":   "OpportunityName",
			"amount":    "Amount",
			"stage":     "StageName",
		}
	case domain.CrmHubSpot:
		return map[string]string{
			"lastName":  "lastname",
			"firstName": "firstname",
			"email":     "email",
			"phone":     "phone",
			"company":   "company",
			"orderId":   "dealname",
			"amount":    "amount",
			"stage":     "dealstage",
		}
	case domain.CrmZoho:
		return map[string]string{
			"lastName":  "Last_Name",
			"firstName": "First_Name",
			"email":     "Email",
			"phone":     "Phone",
			"company":   "Account_Name",
			"orderId":   "Deal_Name",
			"amount":    "Amount",
			"stage":     "Stage",
		}
	case domain.CrmPipedrive:
		return map[string]string{
			"lastName":  "name",
			"firstName": "name",
			"email":     "email",
			"phone":     "phone",
			"company":   "org_id",
			"orderId":   "title",
			"amount":    "value",
			"stage":     "stage_id",
		}
	}
	return map[string]string{}
}

// FormatContact dispatches to the correct CRM formatter.
func (a *CrmAdapter) FormatContact(crmSystem domain.CrmSystem, customer map[string]interface{}) map[string]interface{} {
	switch crmSystem {
	case domain.CrmSalesforce:
		return a.FormatContactForSalesforce(customer)
	case domain.CrmHubSpot:
		return a.FormatContactForHubSpot(customer)
	case domain.CrmZoho:
		return a.FormatContactForZoho(customer)
	case domain.CrmPipedrive:
		return a.FormatContactForPipedrive(customer)
	}
	return map[string]interface{}{}
}

// --- helpers ---

func stringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func floatVal(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		}
	}
	return 0
}
