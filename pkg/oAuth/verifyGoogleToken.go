package oauth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Haidarr-h/backend-go/dto"
	"github.com/Haidarr-h/backend-go/internal/config"
)

func VerifyGoogleToken(idToken string, cfg *config.Config) (*dto.GoogleUserInfo, error) {
	url := "https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// check if google accept it (if expired, malformed, fake)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid google token")
	}

	// read the body to get user data
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// put the user data to the userInfo
	var userInfo dto.GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	// make sure the token intended for our app
	clientID := cfg.GoogleClientID
	var tokenData map[string]interface{}

	json.Unmarshal(body, &tokenData)
	if aud, ok := tokenData["aud"].(string); !ok || aud != clientID {
		return nil, errors.New("token audience mismatch")
	}

	return &userInfo, nil
}
