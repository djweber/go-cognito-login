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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var UserPoolID string

var rootCmd = &cobra.Command{
	Use:   "olog",
	Short: "olog is a simple oauth2 client credentials flow tool",
	Run: func(cmd *cobra.Command, args []string) {
		readEnvs()
	},
}

type (
	oauthCredentials struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
)

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func readEnvs() {

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

	var maxResults int64 = 60

	ups, err := cidp.ListUserPools(&cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: &maxResults,
	})

	if err != nil {
		exit(err)
	}

	poolIndex, err := fuzzyfinder.Find(
		ups.UserPools,
		func(i int) string {
			return *ups.UserPools[i].Name
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	up, err := cidp.DescribeUserPool(&cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: ups.UserPools[poolIndex].Id,
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	upc, err := cidp.ListUserPoolClients(&cognitoidentityprovider.ListUserPoolClientsInput{
		UserPoolId: ups.UserPools[poolIndex].Id,
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	clientIndex, err := fuzzyfinder.Find(
		upc.UserPoolClients,
		func(i int) string {
			return *upc.UserPoolClients[i].ClientName
		},
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	co, err := cidp.DescribeUserPoolClient(&cognitoidentityprovider.DescribeUserPoolClientInput{
		UserPoolId: upc.UserPoolClients[clientIndex].UserPoolId,
		ClientId:   upc.UserPoolClients[clientIndex].ClientId,
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	makeRequest(*up.UserPool.Domain, co)
}

func makeRequest(tokenUrl string, co *cognitoidentityprovider.DescribeUserPoolClientOutput) {
	params := url.Values{}
	params.Add("grant_type", `client_credentials`)
	params.Add("client_id", *co.UserPoolClient.ClientId)
	params.Add("client_secret", *co.UserPoolClient.ClientSecret)

	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/oauth2/token", tokenUrl), body)

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
