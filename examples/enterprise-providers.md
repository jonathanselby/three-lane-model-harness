# Enterprise provider examples — V0.2

These examples show configuration shape only. Use model identifiers/deployments that exist in your own cloud account.

Credentials are environment variables rather than command-line flags.

## Microsoft Foundry / Azure OpenAI

Using an API key:

```bash
export AZURE_OPENAI_ENDPOINT="https://YOUR-RESOURCE.openai.azure.com"
export AZURE_OPENAI_API_KEY="..."

go run ./cmd/harness \
  -complexity standard \
  -standard-provider azure \
  -standard-model YOUR-DEPLOYMENT \
  -prompt "Summarise these paragraphs and identify the main risks."
```

Using a short-lived Microsoft Entra bearer token instead:

```bash
export AZURE_OPENAI_ENDPOINT="https://YOUR-RESOURCE.openai.azure.com"
export AZURE_OPENAI_BEARER_TOKEN="..."
```

If both Azure credentials are present, the bearer token is preferred.

## Amazon Bedrock

V0.2 uses the Bedrock Runtime Converse API with a Bedrock bearer token supplied by the deployment environment.

```bash
export AWS_REGION="ap-southeast-2"
export AWS_BEARER_TOKEN_BEDROCK="..."

go run ./cmd/harness \
  -complexity complex \
  -complex-provider bedrock \
  -complex-model YOUR-BEDROCK-MODEL-ID \
  -prompt "Compare these two architecture proposals and explain the trade-offs."
```

The harness does not acquire or refresh AWS credentials. That remains a deployment/IAM responsibility.

## Google Vertex AI

Supply a project and short-lived Google Cloud access token:

```bash
export GOOGLE_CLOUD_PROJECT="your-project"
export VERTEX_LOCATION="australia-southeast1"
export GOOGLE_OAUTH_ACCESS_TOKEN="..."

go run ./cmd/harness \
  -complexity simple \
  -simple-provider vertex \
  -simple-model YOUR-GEMINI-MODEL \
  -prompt "Extract the invoice number from this text."
```

Token acquisition through ADC, workload identity or `gcloud` is deliberately outside this harness.

## Mixed-cloud lanes

The point of V0.2 is that lanes do not have to share a provider.

```bash
go run ./cmd/harness \
  -complexity complex \
  -simple-provider azure \
  -simple-model YOUR-AZURE-DEPLOYMENT \
  -standard-provider vertex \
  -standard-model YOUR-GEMINI-MODEL \
  -complex-provider bedrock \
  -complex-model YOUR-BEDROCK-MODEL-ID \
  -prompt "Compare these two architecture proposals and explain the trade-offs."
```

Only the selected lane is invoked. The caller still supplies the complexity class; the harness does not choose a cloud.

## Important boundary

`-sensitive` records caller metadata only. It does not stop a prompt being sent to a configured external provider. This repository is a routing example, not a security gateway.
