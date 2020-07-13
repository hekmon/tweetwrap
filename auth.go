package tweetwrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/dghubble/oauth1"
)

const (
	outOfBand      = "oob"
	tokenCredsFile = "token_credentials.json" // relative path to current working directory
)

func (w *Wrapper) getAuthURL() (authorizationURL *url.URL, err error) {
	w.tokens.RequestToken, _, err = w.oauth1Config.RequestToken()
	if err != nil {
		err = fmt.Errorf("can't get a request token: %w", err)
		return
	}
	authorizationURL, err = w.oauth1Config.AuthorizationURL(w.tokens.RequestToken)
	if err != nil {
		err = fmt.Errorf("can't generate an authorization URL (request token '%s'): %w", w.tokens.RequestToken, err)
	}
	return
}

func (w *Wrapper) finalizeUserAuth(PIN string) (err error) {
	accessToken, accessSecret, err := w.oauth1Config.AccessToken(w.tokens.RequestToken, "secret does not matter", PIN)
	if err != nil {
		err = fmt.Errorf("(requestToken: '%s') (PIN: %s): %w", w.tokens.RequestToken, PIN, err)
		return
	}
	w.tokens.AccessToken = oauth1.NewToken(accessToken, accessSecret)
	w.tokens.RequestToken = "" // user auth done, clean intermediary state
	return
}

type oauthTokens struct {
	AccessToken  *oauth1.Token
	RequestToken string
}

func (w *Wrapper) restoreStatus() (err error) {
	// handle file descriptor
	fd, err := os.Open(w.stateFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		} else {
			err = fmt.Errorf("can't open auth status file (%s): %w", w.stateFilePath, err)
		}
		return
	}
	defer fd.Close()
	// handle content
	if err = json.NewDecoder(fd).Decode(&w.tokens); err != nil {
		err = fmt.Errorf("extracting tokens failed: %w", err)
		return
	}
	return
}

// SaveStatus dump the current auth & requests tokens to disk.
// Wihtout a restored state, New() will always trigger a new auth process:
// call it before exiting !
func (w *Wrapper) SaveStatus() (err error) {
	// handle file descriptor
	fd, err := os.OpenFile(w.stateFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		err = fmt.Errorf("can't open state file for writting: %w", err)
		return
	}
	defer fd.Close()
	// handle content
	if err = json.NewEncoder(fd).Encode(w.tokens); err != nil {
		err = fmt.Errorf("encoding state as JSON failed: %w", err)
		return
	}
	return
}

// GetAuthedUser returns the account currently authenticated
func (w *Wrapper) GetAuthedUser() string {
	return w.user
}
