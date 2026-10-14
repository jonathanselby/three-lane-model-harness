package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jonathanselby/three-lane-model-harness/internal/harness"
	"github.com/jonathanselby/three-lane-model-harness/internal/provider"
	"github.com/jonathanselby/three-lane-model-harness/internal/provider/azure"
	"github.com/jonathanselby/three-lane-model-harness/internal/provider/bedrock"
	"github.com/jonathanselby/three-lane-model-harness/internal/provider/mock"
	"github.com/jonathanselby/three-lane-model-harness/internal/provider/ollama"
	"github.com/jonathanselby/three-lane-model-harness/internal/provider/vertex"
)

func main() {
	complexity := flag.String("complexity", "", "workload complexity: simple, standard, complex")
	prompt := flag.String("prompt", "", "task or prompt to execute")
	sensitive := flag.Bool("sensitive", false, "record that the caller considers the task sensitive; does not enforce policy")

	simpleProvider := flag.String("simple-provider", "mock", "provider for simple lane")
	simpleModel := flag.String("simple-model", "mock-small", "model for simple lane")
	standardProvider := flag.String("standard-provider", "mock", "provider for standard lane")
	standardModel := flag.String("standard-model", "mock-standard", "model for standard lane")
	complexProvider := flag.String("complex-provider", "mock", "provider for complex lane")
	complexModel := flag.String("complex-model", "mock-strong", "model for complex lane")

	ollamaURL := flag.String("ollama-url", "http://localhost:11434", "Ollama base URL")
	azureEndpoint := flag.String("azure-endpoint", os.Getenv("AZURE_OPENAI_ENDPOINT"), "Microsoft Foundry / Azure OpenAI resource endpoint")
	bedrockRegion := flag.String("bedrock-region", firstNonEmpty(os.Getenv("AWS_REGION"), os.Getenv("AWS_DEFAULT_REGION")), "Amazon Bedrock region")
	bedrockURL := flag.String("bedrock-url", "", "optional Amazon Bedrock Runtime base URL override")
	vertexProject := flag.String("vertex-project", os.Getenv("GOOGLE_CLOUD_PROJECT"), "Google Cloud project for Vertex AI")
	vertexLocation := flag.String("vertex-location", firstNonEmpty(os.Getenv("VERTEX_LOCATION"), "global"), "Vertex AI location")
	vertexURL := flag.String("vertex-url", "", "optional Vertex AI base URL override")
	timeout := flag.Duration("timeout", 60*time.Second, "request timeout")
	flag.Parse()

	if *complexity == "" || *prompt == "" {
		fmt.Fprintln(os.Stderr, "-complexity and -prompt are required")
		flag.Usage()
		os.Exit(2)
	}

	cfg := harness.Config{
		Simple:   harness.Lane{Provider: *simpleProvider, Model: *simpleModel},
		Standard: harness.Lane{Provider: *standardProvider, Model: *standardModel},
		Complex:  harness.Lane{Provider: *complexProvider, Model: *complexModel},
	}

	providers := []provider.Provider{
		mock.Provider{},
		ollama.Provider{BaseURL: *ollamaURL},
		azure.Provider{
			Endpoint:    *azureEndpoint,
			APIKey:      os.Getenv("AZURE_OPENAI_API_KEY"),
			BearerToken: os.Getenv("AZURE_OPENAI_BEARER_TOKEN"),
		},
		bedrock.Provider{
			Region:      *bedrockRegion,
			BaseURL:     *bedrockURL,
			BearerToken: os.Getenv("AWS_BEARER_TOKEN_BEDROCK"),
		},
		vertex.Provider{
			ProjectID:   *vertexProject,
			Location:    *vertexLocation,
			BaseURL:     *vertexURL,
			AccessToken: os.Getenv("GOOGLE_OAUTH_ACCESS_TOKEN"),
		},
	}
	router, err := harness.New(cfg, providers...)
	if err != nil {
		fatalJSON(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	result, err := router.Run(ctx, harness.Request{
		Task:       *prompt,
		Complexity: harness.Complexity(*complexity),
		Sensitive:  *sensitive,
	})
	if err != nil {
		writeJSON(result)
		os.Exit(1)
	}
	writeJSON(result)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func writeJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func fatalJSON(err error) {
	writeJSON(map[string]string{"error": err.Error()})
	os.Exit(1)
}
