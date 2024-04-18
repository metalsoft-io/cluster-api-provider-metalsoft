package metalsoft

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v2"
)

const (
	userEmailEnvVar      = "METALSOFT_USER_EMAIL"
	apiKeyEnvVar         = "METALSOFT_API_KEY"
	apiEndpointEnvVar    = "METALSOFT_ENDPOINT"
	datacenterEnvVar     = "METALSOFT_DATACENTER"
	loggingEnabledEnvVar = "METALSOFT_LOGGING_ENABLED"

	credentialsFileEnvVar = "METALSOFT_CREDENTIALS_FILE_PATH"
)

const endpointPath = "/api/developer/developer"

type MetalSoftClient struct {
	*metalcloud.Client
}

type credential struct {
	UserEmail string `json:"user_email"`
	APIKey    string `json:"api_key"`
	Endpoint  string `json:"endpoint"`
	Logging   bool   `json:"logging,omitempty"`
}

// GetClient returns a new MetalSoft client.
func GetClient() (*MetalSoftClient, error) {
	// Try to get credentials from a file first.
	credentials, err := getCredentialsFromFile()
	if err != nil {
		fmt.Println("Error getting credentials from file:", err)
		fmt.Println("Falling back to environment variables...")
		// If that fails, fall back to environment variables.
		credentials, err = getCredentialsFromEnv()
		if err != nil {
			return nil, err
		}
	}
	// log.Printf("credentials: %+v", credentials)
	client, err := metalcloud.GetMetalcloudClient(credentials.UserEmail, credentials.APIKey, credentials.Endpoint, credentials.Logging, "", "", "")
	if err != nil {
		return nil, err
	}

	return &MetalSoftClient{client}, nil
}

func getCredentialsFromEnv() (*credential, error) {
	userEmail, exists := os.LookupEnv(userEmailEnvVar)
	if !exists {
		return nil, fmt.Errorf("%s is not defined", userEmailEnvVar)
	}

	err := validateUserEmail(userEmail)
	if err != nil {
		return nil, err
	}

	apiKey, exists := os.LookupEnv(apiKeyEnvVar)
	if !exists {
		return nil, fmt.Errorf("%s is not defined", apiKeyEnvVar)
	}

	err = validateAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	apiEndpoint, exists := os.LookupEnv(apiEndpointEnvVar)
	if !exists {
		return nil, fmt.Errorf("%s is not defined", apiEndpointEnvVar)
	}

	apiEndpoint, err = developerEndpoint(endpointPath, apiEndpoint)
	if err != nil {
		return nil, err
	}

	var loggingEnabled bool
	loggingEnabledStr, exists := os.LookupEnv(loggingEnabledEnvVar)
	if !exists || loggingEnabledStr == "" {
		loggingEnabled = false
	} else {
		loggingEnabled = loggingEnabledStr == "true"
	}

	return &credential{
		UserEmail: userEmail,
		APIKey:    apiKey,
		Endpoint:  apiEndpoint,
		Logging:   loggingEnabled,
	}, nil
}

func getCredentialsFromFile() (*credential, error) {
	credsPath, exists := os.LookupEnv(credentialsFileEnvVar)
	if !exists {
		return nil, fmt.Errorf("%s is not defined", credentialsFileEnvVar)
	}

	byteValue, err := os.ReadFile(credsPath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("reading credentials from file %s: %w", credsPath, err)
	}

	var credential credential
	err = json.Unmarshal(byteValue, &credential)
	if err != nil {
		return nil, err
	}

	// log.Printf("credentials: %+v", credential)
	err = validateUserEmail(credential.UserEmail)
	if err != nil {
		return nil, err
	}

	err = validateAPIKey(credential.APIKey)
	if err != nil {
		return nil, err
	}

	apiEndpoint, err := developerEndpoint(endpointPath, credential.Endpoint)
	if err != nil {
		return nil, err
	}

	credential.Endpoint = apiEndpoint
	return &credential, nil
}

func developerEndpoint(endpointPath string, apiEndpoint string) (string, error) {
	if !strings.HasPrefix(apiEndpoint, "https://") {
		return "", fmt.Errorf("%s must start with https://", apiEndpointEnvVar)
	}

	parsedURL, err := url.Parse(apiEndpoint)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %v", err)
	}

	baseURL := parsedURL.Scheme + "://" + parsedURL.Host
	if baseURL == "" {
		return "", fmt.Errorf("invalid baseURL")
	}

	developerApiEndpoint := baseURL + endpointPath
	return developerApiEndpoint, nil
}

func validateUserEmail(userEmail string) error {
	const pattern = "^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$"

	matched, _ := regexp.MatchString(pattern, userEmail)

	if !matched {
		return fmt.Errorf("user email is not valid")
	}

	return nil
}

func validateAPIKey(apiKey string) error {
	const pattern = "^\\d+\\:[0-9a-zA-Z]*$"

	matched, _ := regexp.MatchString(pattern, apiKey)

	if !matched {
		return fmt.Errorf("API Key is not valid. It should start with a number followed by a semicolon followed by alphanumeric characters <id>:<chars> ")
	}

	return nil
}
