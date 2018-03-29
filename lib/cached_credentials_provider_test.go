package lib

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/mbndr/logo"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type mockSessionTokenProvider struct {
	awsAssumeRoleProvider
}

// override Retrieve() and AssumeRole() so we can bypass/mock calls to AWS services
func (m *mockSessionTokenProvider) Retrieve() (credentials.Value, error) {
	// lazy load credentials
	c, err := m.credsFromFile()
	if err == nil {
		m.log.Debugf("Found cached session token credentials")
		m.creds = c
	}

	if m.IsExpired() {
		m.log.Debugf("Detected expired or unset session token credentials, refreshing")
		c = &CachableCredentials{
			Expiration: time.Now().Add(m.sessionTokenDuration).Unix(),
			Value: credentials.Value{
				AccessKeyID:     "MockSessionTokenAccessKey",
				SecretAccessKey: "MockSessionTokenSecretKey",
				SessionToken:    "MockSessionToken",
				ProviderName:    "MockCredentialsProvider",
			},
		}
		m.creds = c
		m.Store()
	}

	return m.creds.Value, nil
}

func (m *mockSessionTokenProvider) AssumeRole() (credentials.Value, error) {
	v := credentials.Value{
		AccessKeyID:     "MockAssumeRoleAccessKey",
		SecretAccessKey: "MockAssumeRoleSecretKey",
		SessionToken:    "MockAssumeRoleSessionToken",
		ProviderName:    "MockCredentialsProvider",
	}
	return v, nil
}

func TestProviderDefaults(t *testing.T) {
	os.Unsetenv("AWS_CONFIG_FILE")
	m := new(mockSessionTokenProvider)
	m.setAttrs(new(AWSProfile), &SessionTokenProviderOptions{LogLevel: logo.INFO})

	t.Run("CacheFile", func(t *testing.T) {
		if !strings.HasSuffix(m.CacheFile(), filepath.Join(string(filepath.Separator), ".aws_session_token_")) {
			t.Errorf("Cache file path does not have expected contents: %s", m.CacheFile())
		}
	})
	t.Run("ExpirationEpoch", func(t *testing.T) {
		if !m.ExpirationTime().Equal(time.Unix(0, 0)) {
			t.Errorf("Expiration time for unset credentials != Unix epoch time")
		}
	})
	t.Run("NoCreds", func(t *testing.T) {
		if !m.IsExpired() {
			t.Errorf("Expected IsExpired() for unset credentials to be true, got %v", m.IsExpired())
		}
	})
	t.Run("AssumeRole", func(t *testing.T) {
		c, err := m.AssumeRole()
		if err != nil {
			t.Errorf("Unexpected error during AssumeRole(): %v", err)
		}
		if c.AccessKeyID != "MockAssumeRoleAccessKey" || c.SecretAccessKey != "MockAssumeRoleSecretKey" ||
			c.SessionToken != "MockAssumeRoleSessionToken" || c.ProviderName != "MockCredentialsProvider" {
			t.Errorf("Data mismatch when validating assume role credentials: %v", c)
		}
	})
}

func TestProviderCustomConfig(t *testing.T) {
	os.Setenv("AWS_CONFIG_FILE", "aws.cfg")
	p := AWSProfile{
		Region: "us-west-1",
	}

	opts := SessionTokenProviderOptions{
		LogLevel:             logo.INFO,
		SessionTokenDuration: 8 * time.Hour,
		RoleArn:              "arn:aws:iam::0123456789:role/mock-role",
		MfaSerial:            "mock-mfa",
	}
	m := new(mockSessionTokenProvider)
	m.setAttrs(&p, &opts)

	t.Run("CacheFile", func(t *testing.T) {
		if m.CacheFile() != ".aws_session_token_" {
			t.Errorf("Cache file path does not have expected contents: %s", m.CacheFile())
		}
	})
	t.Run("SessionTokenRetrieve", func(t *testing.T) {
		c, err := m.Retrieve()
		if err != nil {
			t.Errorf("Error in Retrieve(): %v", err)
		}
		if c.AccessKeyID != "MockSessionTokenAccessKey" || c.SecretAccessKey != "MockSessionTokenSecretKey" ||
			c.SessionToken != "MockSessionToken" || c.ProviderName != "MockCredentialsProvider" {
			t.Errorf("Data mismatch when validating assume role credentials: %v", c)
		}
	})

	// These tests require that something has called Retrieve()
	t.Run("SessionTokenExpirationTime", func(t *testing.T) {
		if m.ExpirationTime() == time.Unix(0, 0) {
			t.Errorf("Credentials have invalid expiration")
		}
	})
	t.Run("SessionTokenNotExpired", func(t *testing.T) {
		if m.IsExpired() {
			t.Errorf("Unexpected expired credentials")
		}
	})
	t.Run("CredsFromFile", func(t *testing.T) {
		c, err := m.credsFromFile()
		if err != nil {
			t.Errorf("Error loading cached credentials: %v", err)
		}
		if c == nil {
			t.Errorf("nil credentials read from file")
		}
	})
	t.Run("ExpiredCredentials", func(t *testing.T) {
		m.creds.Expiration = time.Now().Add(-5 * time.Second).Unix()
		if m.ExpirationTime().After(time.Now()) {
			t.Errorf("Unexpected future credential expiration time")
		}
		if !m.IsExpired() {
			t.Errorf("Expected expired credentials, but received valid")
		}
	})

	os.Remove(m.CacheFile())
	os.Unsetenv("AWS_CONFIG_FILE")
}

func TestNewProviderParams(t *testing.T) {
	os.Unsetenv("AWS_CONFIG_FILE")
	m := new(mockSessionTokenProvider)

	t.Run("NilProfile", func(t *testing.T) {
		defer func() {
			if x := recover(); x != nil {
				t.Logf("Got expected panic() from nil profile")
			} else {
				t.Errorf("Did not see expected panic() from nil profile")
			}
		}()
		m.setAttrs(nil, new(SessionTokenProviderOptions))
	})
	t.Run("NilOptions", func(t *testing.T) {
		defer func() {
			if x := recover(); x != nil {
				t.Logf("Got expected panic() from nil options")
			} else {
				t.Errorf("Did not see expected panic() from nil options")
			}
		}()
		m.setAttrs(new(AWSProfile), nil)
	})
	t.Run("BadArn", func(t *testing.T) {
		if err := m.setAttrs(new(AWSProfile), &SessionTokenProviderOptions{RoleArn: "bogus"}); err == nil {
			t.Errorf("Did not see expected error using bad RoleArn option")
		}
	})
	t.Run("DefaultSessionDuration", func(t *testing.T) {
		if m.sessionTokenDuration != SESSION_TOKEN_DEFAULT_DURATION {
			t.Errorf("Session token duration is not default value")
		}
	})
	t.Run("BelowMinSessionDuration", func(t *testing.T) {
		m.setAttrs(new(AWSProfile), &SessionTokenProviderOptions{SessionTokenDuration: 1 * time.Minute})
		if m.sessionTokenDuration != SESSION_TOKEN_MIN_DURATION {
			t.Errorf("Session token duration is not min value")
		}
	})
	t.Run("AboveMaxSessionDuration", func(t *testing.T) {
		m.setAttrs(new(AWSProfile), &SessionTokenProviderOptions{SessionTokenDuration: 100 * time.Hour})
		if m.sessionTokenDuration != SESSION_TOKEN_MAX_DURATION {
			t.Errorf("Session token duration is not max value")
		}
	})
	t.Run("BelowMinRoleDuration", func(t *testing.T) {
		m.setAttrs(new(AWSProfile), &SessionTokenProviderOptions{SessionTokenDuration: 1 * time.Minute})
		if m.assumeRoleDuration != ASSUME_ROLE_MIN_DURATION {
			t.Errorf("Assume role duration is not min value")
		}
	})
	t.Run("AboveMaxRoleDuration", func(t *testing.T) {
		m.setAttrs(new(AWSProfile), &SessionTokenProviderOptions{SessionTokenDuration: 18 * time.Hour})
		if m.assumeRoleDuration != ASSUME_ROLE_MAX_DURATION {
			t.Errorf("Assume role duration is not max value")
		}
	})
}
