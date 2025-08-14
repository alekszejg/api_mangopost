package gmail

import (
	"context"
	"errors"
	"fmt"
	utils "my-api/utils"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var config *oauth2.Config
var emailUser string

func InitConfig() error {
	envs := []struct {
		name, value string
	}{
		{name: "G_CLIENT_ID", value: os.Getenv("G_CLIENT_ID")},
		{name: "G_SECRET", value: os.Getenv("G_SECRET")},
		{name: "MAIN_URL", value: os.Getenv("MAIN_URL")},
		{name: "G_MAIL", value: os.Getenv("G_MAIL")},
	}

	values := make(map[string]string)

	for _, env := range envs {
		if env.value == "" {
			return fmt.Errorf("failed to initialize OAuth config: missing %s value", env.name)
		} else {
			values[env.name] = env.value
		}
	}

	emailUser = values["G_MAIL"]

	config = &oauth2.Config{
		ClientID:     values["G_CLIENT_ID"],
		ClientSecret: values["G_SECRET"],
		RedirectURL:  values["MAIN_URL"] + "/auth/callback",
		Scopes:       []string{gmail.GmailReadonlyScope},
		Endpoint:     google.Endpoint,
	}

	return nil
}

func getGmailService(ctx context.Context) (*gmail.Service, error) {
	token, err := loadToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load token: %w", err)
	}

	client := config.Client(ctx, token.Token)
	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}
	return service, nil
}

// ONLY FOR INITIAL 1ST LAUNCH OR WHEN gmail/token.json IS LOST
// DONT FORGET! Change OAuthCallback url in Google Cloud in production
func OAuthHandler(ctx *gin.Context) {
	state := utils.GetRandomState()
	ctx.SetCookie("oauth_state", state, 3600, "/", "", false, true) // Keep state cookie for safety
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	ctx.Redirect(302, url)
}

func exchangeToken(ctx *gin.Context) (*Token, error) {
	code := ctx.Query("code")
	if code == "" {
		return nil, &utils.APIError{
			Err:    fmt.Errorf("invalid query parameter 'code'"),
			Status: 400,
		}
	}

	token, err := config.Exchange(ctx.Request.Context(), code)
	if err != nil {
		return nil, &utils.APIError{
			Err:    fmt.Errorf("failed to exchange oAuth tokens: %w", err),
			Status: 500,
		}
	}

	return &Token{Token: token}, nil
}

func OAuthCallback(ctx *gin.Context) {
	state := ctx.Query("state")
	cookieState, err := ctx.Cookie("oauth_state")
	if err != nil || state != cookieState {
		ctx.JSON(500, gin.H{"error": "Mismatch between recieved and saved state"})
		return
	}

	token, err := exchangeToken(ctx)
	if err != nil {
		var apiErr *utils.APIError
		if errors.As(err, &apiErr) {
			ctx.JSON(apiErr.Status, gin.H{"error": apiErr.Error()})
			return
		}

		ctx.JSON(500, gin.H{"error": "Internal Error"})
		return
	}

	if err := token.Save(); err != nil {
		ctx.JSON(500, gin.H{"error": fmt.Errorf("failed to save token: %w", err)})
		return
	}

	ctx.JSON(200, gin.H{"message": "Authentication was successful"})
}
