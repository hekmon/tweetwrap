package tweetwrap

import (
	"errors"
	"fmt"
	"net/url"

	twittgo "github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	twauth "github.com/dghubble/oauth1/twitter"
)

// Config contains all the values needed by New()
type Config struct {
	APIKey        string
	APIKeySecret  string
	PIN           string
	StateFilePath string // if empty the tokenCredsFile const will be used as path
}

// New returns an initialized and ready to used Twitter Wrapper.
// If err is nil but authorizationURL is not, caller must :
// 1- transmit the URL to its user, let him authorize the account
// to be used and recover the PIN verifier code.
// 2- once the PIN has been recovered, call New again with the correct PIN in Config.
// Intermediary auth state will be loaded automatically from disk.
// Only a nil err and authorizationURL mean the client within the wrapper can be used.
func New(conf Config) (w *Wrapper, authorizationURL *url.URL, err error) {
	/*
		Controller
	*/
	w = new(Wrapper)
	if conf.StateFilePath == "" {
		w.stateFilePath = tokenCredsFile
	} else {
		w.stateFilePath = conf.StateFilePath
	}
	/*
		Prepare app auth
	*/
	w.oauth1Config = oauth1.Config{
		ConsumerKey:    conf.APIKey,
		ConsumerSecret: conf.APIKeySecret,
		CallbackURL:    outOfBand,
		Endpoint:       twauth.AuthorizeEndpoint,
	}
	/*
		User auth
	*/
	if err = w.restoreStatus(); err != nil {
		err = fmt.Errorf("can't restore previous tokens: %w", err)
		return
	}
	if w.tokens.AccessToken == nil {
		// no previous access token has been found, should we generate an authURL or finalize an ongoing registration ?
		if w.tokens.RequestToken == "" {
			// new auth
			authorizationURL, err = w.getAuthURL()
			if err != nil {
				err = fmt.Errorf("getting an auth URL failed: %w", err)
				return
			}
			// save intermediary auth state
			if err = w.SaveStatus(); err != nil {
				err = fmt.Errorf("authentification URL generated but authentification state can't be saved: %w", err)
				return
			}
			return // caller will need to transmit the auth URL to its user and call New() again with PIN verification setup
		}
		// finalize ongoing auth
		if conf.PIN == "" {
			err = errors.New("ongoing authorization state loaded but no PIN provided")
			return
		}
		if err = w.finalizeUserAuth(conf.PIN); err != nil {
			err = fmt.Errorf("can't get user access token: %w", err)
			return
		}
		// save final auth state
		if err = w.SaveStatus(); err != nil {
			err = fmt.Errorf("can't save authentificated state: %w", err)
			return
		}
	}
	/*
		Spawn oauth ready client
	*/
	w.Client = twittgo.NewClient(w.oauth1Config.Client(oauth1.NoContext, w.tokens.AccessToken))
	f := new(bool)
	t := new(bool)
	*t = true
	user, _, err := w.Client.Accounts.VerifyCredentials(&twittgo.AccountVerifyParams{
		SkipStatus:   t,
		IncludeEmail: f,
	})
	if err != nil {
		err = fmt.Errorf("authentification verification failed: %w", err)
		return
	}
	w.user = fmt.Sprintf("%s (@%s)", user.Name, user.ScreenName)
	return
}

// Wrapper handles the 3-legged OAuth process automatically during its initialization.
// Must be instanciated with New()
type Wrapper struct {
	// controller config
	stateFilePath string
	// oauth config
	oauth1Config oauth1.Config
	tokens       oauthTokens
	user         string
	// wrapped client
	Client *twittgo.Client
}
