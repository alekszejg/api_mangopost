package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
)

type Token struct {
	*oauth2.Token
}

func (t *Token) Refresh(ctx context.Context) error {
	if t.AccessToken != "" && t.Expiry.Before(time.Now()) {
		newToken, err := config.TokenSource(ctx, t.Token).Token()
		if err != nil {
			return fmt.Errorf("token.Refresh() failed to refresh token: %s", err.Error())
		}

		if newToken.AccessToken != t.AccessToken {
			t.Token = newToken
			if err := t.Save(); err != nil {
				return fmt.Errorf("token.Refresh() failed to save refreshed token: %s", err.Error())
			}
		}
	}

	return nil
}

func (t *Token) Save() error {
	data, err := json.MarshalIndent(t.Token, "", "  ")
	if err != nil {
		return fmt.Errorf("token.Save() failed to marshal token data: %s", err.Error())
	}

	if err := os.WriteFile("gmail/token.json", data, 0600); err != nil {
		return fmt.Errorf("token.Save() failed to write to gmail/token.json: %s", err.Error())
	}

	return nil
}

func loadToken(ctx context.Context) (*Token, error) {
	data, err := os.ReadFile("gmail/token.json")
	if err != nil {
		return nil, fmt.Errorf("loadToken() failed to read gmail/token.json: %s", err.Error())
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("loadToken() failed to unmarshal token data: %s", err.Error())
	}

	t := &Token{Token: &token}
	if err := t.Refresh(ctx); err != nil {
		return nil, fmt.Errorf("loadToken() failed to call token.Refresh(): %s", err.Error())
	}

	return t, nil
}
