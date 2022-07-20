package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "olog",
	Short: "olog is a simple oauth2 client credentials flow tool",
	Run: func(cmd *cobra.Command, args []string) {
		readEnvs()
	},
}

type (
	config struct {
		Envs map[string]env
	}
	env struct {
		TokenURL   string `toml:"token_url"`
		ClientID   string `toml:"client_id"`
		UserPoolID string `toml:"user_pool_id"`
	}
	fzfEnv struct {
		Name string
		Env  env
	}
	oauthCredentials struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
)

func readEnvs() {
	home, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	f := fmt.Sprintf("%s/.olog/config.toml", home)

	var c config

	_, err = toml.DecodeFile(f, &c)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	envs := []fzfEnv{}

	for k, v := range c.Envs {
		envs = append(envs, fzfEnv{
			Name: k,
			Env:  v,
		})
	}

	i, err := fuzzyfinder.Find(
		envs,
		func(i int) string {
			return envs[i].Name
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	makeRequest(envs[i])
}

func makeRequest(env fzfEnv) {
	fmt.Printf("Environment: %s\n", env.Name)

	cfg := &aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewEnvCredentials(),
	}

	sess, err := session.NewSession(cfg)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cidp := cognitoidentityprovider.New(sess)

	co, err := cidp.DescribeUserPoolClient(&cognitoidentityprovider.DescribeUserPoolClientInput{
		ClientId:   &env.Env.ClientID,
		UserPoolId: &env.Env.UserPoolID,
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	params := url.Values{}
	params.Add("grant_type", `client_credentials`)
	params.Add("client_id", *co.UserPoolClient.ClientId)
	params.Add("client_secret", *co.UserPoolClient.ClientSecret)

	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", env.Env.TokenURL, body)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	var creds oauthCredentials

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &creds)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Access Token: %s\n", creds.AccessToken)
	fmt.Printf("Expires In: %d\n", creds.ExpiresIn)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
