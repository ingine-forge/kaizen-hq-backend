package torn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client interface {
	FetchGymEnergy(ctx context.Context, stat string) (StatMap, error)
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
