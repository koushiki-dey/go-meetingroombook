package googleapi

import (
	"context"
	"net/http"

	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

var config *oauth2.Config

func InitOAuth(clientID, clientSecret, redirectURL string) {
	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}
}

//	func GetAuthURL() string {
//		return config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
//	}
func GetAuthURLWithUser(userID uint) (string, error) {
	state, err := CreateState(userID)
	if err != nil {
		return "", err
	}
	return config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

func ExchangeCode(code string) (*oauth2.Token, error) {
	return config.Exchange(context.Background(), code)
}

func GetClient(token *oauth2.Token) *http.Client {
	return config.Client(context.Background(), token)
}

func generateNonce(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func CreateState(userID uint) (string, error) {
	nonce, err := generateNonce(8)
	if err != nil {
		return "", err
	}
	raw := fmt.Sprintf("%d:%s", userID, nonce)
	return base64.URLEncoding.EncodeToString([]byte(raw)), nil
}

func ParseState(state string) (uint, error) {
	decoded, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return 0, err
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid state format")
	}
	var userID uint
	_, err = fmt.Sscanf(parts[0], "%d", &userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
