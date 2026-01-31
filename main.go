package langfusego

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/adhikag24/langfuse-go/prompt"
)

type Langfuse struct {
	Prompt prompt.Client
}

func New(httpClient *http.Client) Langfuse {
	langfuseSpec, err := getLangfuseSpec()
	if err != nil {
		panic(fmt.Errorf("failed to initialize langfuse-go, error:%w", err))
	}

	return Langfuse{
		Prompt: prompt.New(httpClient, prompt.LangfuseSpec{
			PublicKey:  langfuseSpec.baseUrl,
			PrivateKey: langfuseSpec.secretKey,
			BaseUrl:    langfuseSpec.baseUrl,
		}),
	}
}

type langfuseSpec struct {
	baseUrl   string
	publicKey string
	secretKey string
}

func getLangfuseSpec() (*langfuseSpec, error) {
	secretKey := os.Getenv("LANGFUSE_SECRET_KEY")
	publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	baseUrl := os.Getenv("LANGFUSE_BASE_URL")

	if publicKey == "" || secretKey == "" {
		return nil, errors.New("langfuse public key and secret key must be provided")
	}

	if baseUrl == "" {
		baseUrl = defaultLangfuseBaseUrl
	}

	return &langfuseSpec{
		baseUrl:   baseUrl,
		secretKey: secretKey,
		publicKey: publicKey,
	}, nil
}
