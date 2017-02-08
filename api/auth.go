package main

import (
	"encoding/gob"
	"os"
	"net/http"
	"errors"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"

	"google.golang.org/api/plus/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const (
	defaultSessionID = "default"
	googleProfileSessionKey = "google_profile"
	oauthTokenSessionKey    = "oauth_token"
	additionalUserSeesionKey = "additional_nanika"
	sessionExpirationMinute = 60
)

var (
	OAuthConfig *oauth2.Config
	SessionStore sessions.Store
	SuperUser bool
)

func init() {
	gob.Register(&oauth2.Token{})
	gob.Register(&Profile{})

	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	SuperUser = os.Getenv("USE_SUPER_USER") != ""

	OAuthConfig = configureOAuthClient(clientId, clientSecret)

	cookieStore := sessions.NewCookieStore([]byte("something-very-secret"))
	cookieStore.Options = &sessions.Options{
		HttpOnly: true,
	}
	SessionStore = cookieStore
}

func loginHandler(w http.ResponseWriter, r *http.Request)  {
	uuid, err1 := uuid.NewUUID()
	if err1 != nil {
		c := appengine.NewContext(r)
		log.Infof(c, "%v", err1)
	}
	sessionID := uuid.String()

	oauthFlowSession, err := SessionStore.New(r, sessionID)
	if err != nil {
		return // appErrorf(err, "could not create oauth session: %v", err)
	}
	oauthFlowSession.Options.MaxAge = sessionExpirationMinute * 60 

	if err := oauthFlowSession.Save(r, w); err != nil {
		return // appErrorf(err, "could not save session: %v", err)
	}

	url := OAuthConfig.AuthCodeURL(sessionID, oauth2.ApprovalForce,	oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
	// return nil
}

func oauthCallbackHandler(w http.ResponseWriter, r *http.Request) /* *appError */ {
	ctx := appengine.NewContext(r)

	code := r.FormValue("code")
	tok, err := OAuthConfig.Exchange(ctx, code)
	if err != nil {
		log.Infof(ctx, "could not get auth token: %v", err)
		return
	}

	session, err := SessionStore.New(r, defaultSessionID)
	if err != nil {
		log.Infof(ctx, "could not get default session: %v", err)
		return
	}

	profile, err := fetchProfile(ctx, tok)
	if err != nil {
		log.Infof(ctx, "could not fetch Google profile: %v", err)
		return
	}

	session.Values[oauthTokenSessionKey] = tok
	stripped := stripProfile(profile)
	log.Debugf(ctx, "success! profile:%v", profile)

	session.Values[googleProfileSessionKey] = stripped
	if err := session.Save(r, w); err != nil {
		log.Infof(ctx, "could not save session: %v", err)
		return
	}

	http.Redirect(w, r, "/signup", http.StatusFound)
}

func fetchProfile(ctx context.Context, tok *oauth2.Token) (*plus.Person, error) {
	client := oauth2.NewClient(ctx, OAuthConfig.TokenSource(ctx, tok))
	plusService, err := plus.New(client)
	if err != nil {
		return nil, err
	}
	return plusService.People.Get("me").Do()
}

func logoutHandler(w http.ResponseWriter, r *http.Request)  {
	ctx := appengine.NewContext(r)

	session, err := SessionStore.New(r, defaultSessionID)
	if err != nil {
		log.Infof(ctx, "could not get default session: %v", err)
		return
	}
	session.Options.MaxAge = -1 // Clear session.
	if err := session.Save(r, w); err != nil {
		log.Infof(ctx, "could not save session: %v", err)
		return
	}
	redirectURL := r.FormValue("redirect")
	if redirectURL == "" {
		redirectURL = "/"
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func storeAdditionalInfo(user string, r *http.Request) error {
	ctx := appengine.NewContext(r)

	session, err := SessionStore.New(r, defaultSessionID)
	if err != nil {
		log.Infof(ctx, "could not get default session: %v", err)
		return err
	}

	session.Values[additionalUserSeesionKey] = user

	return nil
}

func profileFromSession(r *http.Request) *Profile {
	if SuperUser {
		return &Profile{
			ID: "0",
			DisplayName: "Super User",
			ImageURL: "",
		}
	}

	session, err := SessionStore.Get(r, defaultSessionID)
	if err != nil {
		return nil
	}
	tok, ok := session.Values[oauthTokenSessionKey].(*oauth2.Token)
	if !ok || !tok.Valid() {
		return nil
	}
	profile, ok := session.Values[googleProfileSessionKey].(*Profile)
	if !ok {
		return nil
	}
	return profile
}

func userFromSession(r *http.Request) (string, error) {
	if SuperUser {
		return "super", nil
	}

	session, err := SessionStore.Get(r, defaultSessionID)
	if err != nil {
		return "", err
	}
	tok, ok := session.Values[oauthTokenSessionKey].(*oauth2.Token)
	if !ok || !tok.Valid() {
		return "", err
	}
	user, ok := session.Values[additionalUserSeesionKey].(string)
	if !ok {
		return "", errors.New("user session is not found in session")
	}
	return user, nil
}

type Profile struct {
	ID, DisplayName, ImageURL string
}

func stripProfile(p *plus.Person) *Profile {
	return &Profile{
		ID:          p.Id,
		DisplayName: p.DisplayName,
		ImageURL:    p.Image.Url,
	}
}

func configureOAuthClient(clientID, clientSecret string) *oauth2.Config {
	redirectURL := os.Getenv("OAUTH2_CALLBACK")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/v1/oauth2callback"
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}
