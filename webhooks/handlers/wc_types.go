package handlers

import (
	"encoding/json"
)

type NewUser struct {
	DateCreated string `json:"date_created"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Username    string `json:"username"`
}

type NewOrder struct {
	ID        int    `json:"id"`
	Total     string `json:"total"`
	TotalTax  string `json:"total_tax"`
	PayMethod string `json:"payment_method"`
	Billing   struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`

		Phone string `json:"phone"`
		Email string `json:"email"`

		Address1 string `json:"address_1"`
		PostCode string `json:"postcode"`
		Company  string `json:"company"`
	} `json:"billing"`
	Vendor struct {
		Name    string `json:"shop_name"`
		Address struct {
			Street   string `json:"street_1"`
			PostCode string `json:"zip"`
		} `json:"address"`
	} `json:"store"`
	MetaData []struct {
		ID    int             `json:"id"`
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	} `json:"meta_data"`
}
