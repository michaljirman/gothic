package provider

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/netlify/gotrue/conf"
	"golang.org/x/oauth2"
)

// Gitlab

type gitlabProvider struct {
	*oauth2.Config
	External conf.ExternalConfiguration
}

func defaultBase(base string) string {
	if base == "" {
		return "https://gitlab.com"
	}

	baseLen := len(base)
	if base[baseLen-1] == '/' {
		return base[:baseLen-1]
	}

	return base
}

func NewGitlabProvider(ext conf.ExternalConfiguration) Provider {
	base := defaultBase(ext.URL)
	return &gitlabProvider{
		Config: &oauth2.Config{
			ClientID:     ext.Key,
			ClientSecret: ext.Secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  base + "/oauth/authorize",
				TokenURL: base + "/oauth/token",
			},
		},
		External: ext,
	}
}

func (g gitlabProvider) GetOAuthToken(ctx context.Context, code string) (*oauth2.Token, error) {
	res, err := http.PostForm(g.Endpoint.TokenURL, url.Values{
		"client_id":     {g.External.Key},
		"client_secret": {g.External.Secret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {g.External.RedirectURI},
	})

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		e := struct {
			Error   string `json:"error"`
			Message string `json:"error_description"`
		}{}

		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}

		return nil, errors.New(e.Message)
	}

	dst := &oauth2.Token{}
	if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
		log.Println("Error response from gitlab: %v", err)
		return nil, err
	}

	return dst, nil
}

func (g gitlabProvider) GetUserEmail(ctx context.Context, tok *oauth2.Token) (string, error) {
	client := g.Client(ctx, tok)
	res, err := client.Get("https://gitlab.com/api/v4/user")
	if err != nil {
		log.Printf("Error decoding request to gitlab for user email: %v", err)
		return "", err
	}
	defer res.Body.Close()

	user := struct {
		Email string `json:"email"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return "", err
	}

	return user.Email, nil
}
