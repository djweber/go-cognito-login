go-cognito-login

Utility for retrieving Cognito-based OAuth2 client credentials written in Go

# prerequisites

This tool leverages the Go AWS SDK to retrieve Cognito user pool information. It currently only
supports the environment-based AWS credentials. Make sure they are exported before using it.

# usage

Run the `gcl` executable and select your user pool and client to retrieve an access token.
