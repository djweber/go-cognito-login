# olog
Utility for retrieving Cognito-based OAuth2 client credentials written in Go

# configuration
Create a configuration file at `~/.olog/config.toml` with the following structure:

```
[envs.dev]
token_url = "https://your.token.url/oauth2/token"
user_pool_id = "your-user-pool-id"
client_id = "your-client-id"

[envs.stg]
token_url = "https://your.token.url/oauth2/token"
user_pool_id = "your-user-pool-id"
client_id = "your-client-id"
```

# usage

This tool leverages the Go AWS SDK to retrieve Cognito user pool information. It currently only
supports the environment-based AWS credentials. Make sure they are exported before using it.
