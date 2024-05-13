package hasmysecretleaked

import (
	"context"
	"fmt"
	"log"
	"net/http"

	gg "github.com/conjurdemos/cyberark-gitguardian-hmsl-remediation-integration-service/brimstone/pkg/gitguardian"
	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
)

var (
	version = "dev"
	DEBUG   bool
	HMSLURL = ""
	// GG_API_ENDPOINT = ""
	GG_API_TOKEN = ""
)

type JWTRequest struct {
	Audience     string `json:"audience"`
	AudienceType string `json:"audience_type"`
}
type AuthenticatePostResponse struct {
	Token      string `json:"token"`
	Detail     string `json:"detail,omitempty"`
	Status     string `json:"status,omitempty"`
	StatusCode int    `json:"statuscode,omitempty"`
}

// TODO: (low priority) move GG authentication into handler functions since the token retrieved is a short-lived session token
func NewClientAuthenticateWithGitGuardian(ctx context.Context, hmslurl *string, audiencetype *string, ggapiurl *string, ggapitoken *string) (*ClientWithResponses, error) {

	jwtreq := gg.PublicJwtCreateJSONRequestBody{
		Audience:     *hmslurl,
		AudienceType: audiencetype,
	}

	respAuth, respAuthErr := AuthenticateWithGitGuardian(ctx, *ggapiurl, *ggapitoken, jwtreq)
	if respAuthErr != nil {
		log.Fatalf("failed call to authenticate: %s", respAuthErr)
	}
	// authResponse, authResponseParseErr := ParseAuthenticateWithGitGuardianPostResponse(respAuth)
	authResponse, authResponseParseErr := gg.ParsePublicJwtCreateResponse(respAuth)

	if authResponseParseErr != nil {
		log.Fatalf("failed to parse auth response: %s", authResponseParseErr)
	}

	bearerTokenProvider, bearerTokenProviderErr := securityprovider.NewSecurityProviderBearerToken(*authResponse.JSON200.Token)
	if bearerTokenProviderErr != nil {
		panic(bearerTokenProviderErr)
	}

	return NewClientWithResponses(*hmslurl, WithRequestEditorFn(bearerTokenProvider.Intercept))
}

func AuthenticateWithGitGuardian(ctx context.Context, ggapiurl string, ggapitoken string, body gg.PublicJwtCreateJSONRequestBody) (*http.Response, error) {
	c, err := gg.NewClient(ggapiurl)
	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}
	req, err := gg.NewPublicJwtCreateRequest(ggapiurl, body)

	// Authorization: Token GG_API_TOKEN"
	apiToken := fmt.Sprintf("Token %s", ggapitoken)
	req.Header.Add("Authorization", apiToken)

	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}
