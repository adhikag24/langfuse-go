package prompt

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"golang.org/x/sync/errgroup"
)

type Client interface {
	GetPromptByName(ctx context.Context, name string) Builder
	CreatePrompt(ctx context.Context, spec *PromptSpec) (*PromptSpec, error)
}

func New(httpClient *http.Client, langfuseSpec LangfuseSpec) Client {
	return &prompthubImpl{
		httpClient:   httpClient,
		cache:        sync.Map{},
		langfuseSpec: langfuseSpec,
	}
}

type prompthubImpl struct {
	httpClient   *http.Client
	cache        sync.Map
	langfuseSpec LangfuseSpec
}

func (c *prompthubImpl) CreatePrompt(ctx context.Context, spec *PromptSpec) (*PromptSpec, error) {
	requestBytes, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/public/v2/prompts", c.langfuseSpec.BaseUrl),
		bytes.NewReader(requestBytes),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.authHeader())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("langfuse returned %d", resp.StatusCode)
	}

	var response PromptSpec
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *prompthubImpl) GetPromptByName(
	ctx context.Context,
	name string,
) Builder {
	if val, ok := c.cache.Load(name); ok {
		cached := val.(cachedPrompt)
		go c.refreshPrompt(context.Background(), name)
		return newBuilder(cached.Chat, cached.Text)
	}

	cached, err := c.fetchAndCachePrompt(ctx, name)
	if err != nil {
		// Return builder with empty prompt chat and text, which will return error when compiling.
		return newBuilder(nil, "")
	}

	return newBuilder(cached.Chat, cached.Text)
}

func (c *prompthubImpl) refreshPrompt(ctx context.Context, name string) {
	newCached, err := c.fetchAndCachePrompt(ctx, name)
	if err != nil {
		return
	}

	if val, ok := c.cache.Load(name); ok {
		cached := val.(cachedPrompt)
		if cached.Text == newCached.Text && isSamePrompt(cached.Chat, newCached.Chat) {
			return
		}
	}

	c.cache.Store(name, newCached)
}

func (c *prompthubImpl) fetchAndCachePrompt(ctx context.Context, name string) (*cachedPrompt, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/public/v2/prompts/%s", c.langfuseSpec.BaseUrl, name),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.authHeader())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("langfuse returned %d", resp.StatusCode)
	}

	var response response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var cached cachedPrompt

	switch response.Type {
	case "text":
		if err := json.Unmarshal(response.Prompt, &cached.Text); err != nil {
			return nil, err
		}

	case "chat":
		if err := json.Unmarshal(response.Prompt, &cached.Chat); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown prompt type: %s", response.Type)
	}

	c.cache.Store(name, cached)
	return &cached, nil
}

func (c *prompthubImpl) InvalidateCache(ctx context.Context, names []string) error {
	group := new(errgroup.Group)
	for _, n := range names {
		name := n
		group.Go(func() error {
			_, err := c.fetchAndCachePrompt(ctx, name)
			return err
		})
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("failed to prefill cache, err: %s", err.Error())
	}

	return nil
}

func (c *prompthubImpl) authHeader() string {
	raw := fmt.Sprintf("%s:%s", c.langfuseSpec.PublicKey, c.langfuseSpec.PrivateKey)
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

func isSamePrompt(a, b []PromptChat) bool {
	return reflect.DeepEqual(a, b)
}

type cachedPrompt struct {
	Text string
	Chat []PromptChat
}

type response struct {
	Type   string          `json:"type"`
	Prompt json.RawMessage `json:"prompt"`
}
