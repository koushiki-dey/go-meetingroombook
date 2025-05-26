package controllers

import (
	"fmt"
	"net/http"

	"github.com/koushikidey/go-meetingroombook/pkg/config"
	"github.com/koushikidey/go-meetingroombook/pkg/googleapi"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	session "github.com/koushikidey/go-meetingroombook/pkg/sessions"
)

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.GetStore().Get(r, "session")
	userID, ok := sess.Values["employee_id"].(uint)
	if !ok {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	authURL, err := googleapi.GetAuthURLWithUser(userID)
	if err != nil {
		http.Error(w, "Failed to create auth URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}
func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Error(w, "Missing code in request", http.StatusBadRequest)
		return
	}
	if state == "" {
		http.Error(w, "Missing state in request", http.StatusBadRequest)
		return
	}

	userID, err := googleapi.ParseState(state)
	if err != nil {
		http.Error(w, "Invalid state parameter: "+err.Error(), http.StatusBadRequest)
		return
	}

	token, err := googleapi.ExchangeCode(code)
	if err != nil {
		http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	newToken := models.GoogleToken{
		EmployeeID:   userID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}
	db := config.GetDB()
	if err := db.Create(&newToken).Error; err != nil {
		http.Error(w, "Failed to save Google token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Google Calendar authorization successful! You may close this tab.")
}
