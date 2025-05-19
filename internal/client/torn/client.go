package torn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client interface {
	FetchGymEnergy(ctx context.Context, stat string) (StatMap, error)
	FetchTornUser(ctx context.Context, tornID string) (*User, error)
	FetchDiscordID(ctx context.Context, tornID int) (string, error)
}

type tornClient struct {
	apiKey string
	client *http.Client
}

func NewTornClient(apiKey string) Client {
	return &tornClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *tornClient) FetchGymEnergy(ctx context.Context, stat string) (StatMap, error) {
	url := fmt.Sprintf("https://api.torn.com/faction/?selections=contributors&stat=%s&key=%s", stat, t.apiKey)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var parsed struct {
		Contributors StatMap `json:"contributors"`
	}

	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	return parsed.Contributors, nil
}

func (t *tornClient) FetchTornUser(ctx context.Context, tornID string) (*User, error) {
	url := fmt.Sprintf("https://api.torn.com/user/%s?key=%s&selections=profile", tornID, t.apiKey)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &user, nil
}

func (t *tornClient) FetchDiscordID(ctx context.Context, tornID int) (string, error) {
	url := fmt.Sprintf("https://api.torn.com/user/%d?key=%s&selections=discord", tornID, t.apiKey)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := t.client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	var parsed struct {
		Discord Discord `json:"discord"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return "", err
	}
	return parsed.Discord.DiscordID, nil
}
