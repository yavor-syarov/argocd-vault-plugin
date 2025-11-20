package config_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/config"
	"github.com/spf13/viper"
)

func TestNewConfig(t *testing.T) {
	testCases := []struct {
		environment  map[string]interface{}
		expectedType string
	}{
		{
			map[string]interface{}{
				"AVP_TYPE":         "vault",
				"AVP_AUTH_TYPE":    "github",
				"AVP_GITHUB_TOKEN": "token",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "token",
				"VAULT_TOKEN":   "token",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "approle",
				"AVP_ROLE_ID":   "role_id",
				"AVP_SECRET_ID": "secret_id",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":           "vault",
				"AVP_AUTH_TYPE":      "k8s",
				"AVP_K8S_MOUNT_PATH": "mount_point",
				"AVP_K8S_ROLE":       "role",
				"AVP_K8S_TOKEN_PATH": "toke_path",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault\n",
				"AVP_AUTH_TYPE": "k8s",
				"AVP_K8S_ROLE":  "role",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":           "vault",
				"AVP_AUTH_TYPE":      "k8s",
				"AVP_K8S_MOUNT_PATH": "mount_point",
				"AVP_K8S_ROLE":       "role",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":       "vault",
				"AVP_AUTH_TYPE":  "k8s",
				"AVP_MOUNT_PATH": "mount_point",
				"AVP_K8S_ROLE":   "role",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "userpass",
				"AVP_USERNAME":  "username",
				"AVP_PASSWORD":  "password",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":       "vault",
				"AVP_AUTH_TYPE":  "userpass",
				"AVP_MOUNT_PATH": "mount_path",
				"AVP_USERNAME":   "username",
				"AVP_PASSWORD":   "password",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":             "ibmsecretsmanager",
				"AVP_IBM_API_KEY":      "token",
				"AVP_IBM_INSTANCE_URL": "http://ibm",
			},
			"*backends.IBMSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":        "ibmsecretsmanager",
				"AVP_IBM_API_KEY": "token",
				"VAULT_ADDR":      "http://ibm",
			},
			"*backends.IBMSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":              "awssecretsmanager",
				"AWS_REGION":            "us-west-1",
				"AWS_ACCESS_KEY_ID":     "id",
				"AWS_SECRET_ACCESS_KEY": "key",
			},
			"*backends.AWSSecretsManager",
		},
		{ // auth via web identity federation is also possible
			map[string]interface{}{
				"AVP_TYPE":                    "awssecretsmanager",
				"AWS_REGION":                  "us-west-1",
				"AWS_WEB_IDENTITY_TOKEN_FILE": "/var/run/secrets/eks.amazonaws.com/serviceaccount/token",
				"AWS_ROLE_ARN":                "arn:aws:iam::111111111:role/argocd-repo-server-secretsmanager-my-cluster",
			},
			"*backends.AWSSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":                       "gcpsecretmanager",
				"GOOGLE_APPLICATION_CREDENTIALS": "../../fixtures/input/gac.json",
			},
			"*backends.GCPSecretManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":            "azurekeyvault",
				"AZURE_TENANT_ID":     "test",
				"AZURE_CLIENT_ID":     "test",
				"AZURE_CLIENT_SECRET": "test",
			},
			"*backends.AzureKeyVault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":                   "yandexcloudlockbox",
				"AVP_YCL_KEY_ID":             "test",
				"AVP_YCL_SERVICE_ACCOUNT_ID": "test",
				"AVP_YCL_PRIVATE_KEY": `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCIwbOQ4mB4LlFKNvvkot8qnKoffHLxVu2+DNpKC3WiPbof23bf
eHcFTj14/h3HP75dxH5GIop2C8HQzyGzScIEHMxqOpwgu8+tmHbsCAdWkbC03wQ0
1++nHmI6kAUx0mFDAXGovyDiR132iZ5lX2hEJ2Nd2g67SHV140sB6T0vRQIDAQAB
AoGASx2B4NnGvRxLwCTVVK71PzWP5/12MQNbUGFE4RjMQxH+kpL8ByDm1v4zm6qQ
dqmXiW9tIF7GiLJKgcPTseOYcdQkGlST1MgYAqtxMkGYYCP94cGna0qy4lIFBJee
B/dKY56UiIEtJbMvN/T9LFBx1Kw5jT4R5lhdysuabsqAt+ECQQDYHmfMee/Dzw+/
G4xlJfIfcQ4648/zf53hlA5MwCBbm6wv2KLkWglzSl9Vy54f/UM4VtIfywjmTkj+
C2b17Uq9AkEAof4tJwllt4AwIjIp1KEiBTY6z0Whoe9SO5RqFmBUkVTeiIuUxgGE
+NLCY+0NzG2FNglT96ik/Xxi+/uiy4wDKQJBAIQ9TpwyfIBe4a65R5XYuyd8AQ4N
uX+wNcYC1yElamdDgP+h2kJJyYCPIHiZ5/6A9LGzhk1H6gEqI8W26mBOuy0CQEcl
y88JYZNmyb07KwQogTioyMugWY01/3gLh0ysonfyPoraQ01z/WMLrjUVOKpAr/E7
x5VOjKiIqTDjJG0h4YECQQDR7tTAXzccGQmHhmN72mDB5LfWi8uSADT4gsimY82m
fDGt+yaf3RaZbVwHSVLzxiXGsu1WQJde3uJeNh5c6z+5
-----END RSA PRIVATE KEY-----`,
			},
			"*backends.YandexCloudLockbox",
		},
		{
			map[string]interface{}{
				"AVP_TYPE": "sops",
			},
			"*backends.LocalSecretManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":         "1passwordconnect",
				"OP_CONNECT_TOKEN": "token",
				"OP_CONNECT_HOST":  "opconnect.somedomain.com",
			},
			"*backends.OnePasswordConnect",
		},
		{
			map[string]interface{}{
				"ARGOCD_ENV_AVP_TYPE":         "vault",
				"ARGOCD_ENV_AVP_AUTH_TYPE":    "github",
				"ARGOCD_ENV_AVP_GITHUB_TOKEN": "token",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"ARGOCD_ENV_AVP_TYPE":         "vault",
				"AVP_TYPE":                    "not-valid-type",
				"ARGOCD_ENV_AVP_AUTH_TYPE":    "github",
				"ARGOCD_ENV_AVP_GITHUB_TOKEN": "token",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":               "keepersecretsmanager",
				"AVP_KEEPER_CONFIG_PATH": "/mnt/foobar",
			},
			"*backends.KeeperSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":             "delineasecretserver",
				"AVP_AUTH_TYPE":        "userpass",
				"AVP_DELINEA_URL":      "http://my-delinea-server",
				"AVP_DELINEA_USER":     "username",
				"AVP_DELINEA_PASSWORD": "password",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE": "kubernetessecret",
			},
			"*backends.KubernetesSecret",
		},
	}
	for _, tc := range testCases {
		for k, v := range tc.environment {
			os.Setenv(k, v.(string))
		}
		viper := viper.New()
		config, err := config.New(viper, &config.Options{})
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		xType := fmt.Sprintf("%T", config.Backend)
		if xType != tc.expectedType {
			t.Errorf("expected: %s, got: %s.", tc.expectedType, xType)
		}
		for k := range tc.environment {
			os.Unsetenv(k)
		}
	}
}

func TestNewConfigNoType(t *testing.T) {
	viper := viper.New()
	_, err := config.New(viper, &config.Options{})
	expectedError := "Must provide a supported Vault Type, received "

	if err.Error() != expectedError {
		t.Errorf("expected error %s to be thrown, got %s", expectedError, err)
	}
}

func TestNewConfigNoAuthType(t *testing.T) {
	os.Setenv("AVP_TYPE", "vault")
	viper := viper.New()
	_, err := config.New(viper, &config.Options{})
	expectedError := "Must provide a supported Authentication Type, received "

	if err.Error() != expectedError {
		t.Errorf("expected error %s to be thrown, got %s", expectedError, err)
	}
	os.Unsetenv("AVP_TYPE")
}

// Helper function that captures log output from a function call into a string
// Adapted from https://stackoverflow.com/a/26806093/170154
func captureOutput(f func()) string {
	var buf bytes.Buffer
	flags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0) // don't include any date or time in the logging messages
	f()
	log.SetOutput(os.Stderr)
	log.SetFlags(flags)
	return buf.String()
}

func TestNewConfigAwsRegionWarning(t *testing.T) {
	testCases := []struct {
		environment  map[string]interface{}
		expectedType string
		expectedLog  string
	}{
		{ // this test issues a warning for missing AWS_REGION env var
			map[string]interface{}{
				"AVP_TYPE":              "awssecretsmanager",
				"AWS_ACCESS_KEY_ID":     "id",
				"AWS_SECRET_ACCESS_KEY": "key",
			},
			"*backends.AWSSecretsManager",
			"warning: AWS_REGION env var not set, using AWS region us-east-2\n",
		},
		{ // no warning is issued
			map[string]interface{}{
				"AVP_TYPE":              "awssecretsmanager",
				"AWS_REGION":            "us-west-1",
				"AWS_ACCESS_KEY_ID":     "id",
				"AWS_SECRET_ACCESS_KEY": "key",
			},
			"*backends.AWSSecretsManager",
			"",
		},
	}

	for _, tc := range testCases {
		for k, v := range tc.environment {
			os.Setenv(k, v.(string))
		}
		viper.Set("verboseOutput", true)

		v := viper.New()
		output := captureOutput(func() {
			config, err := config.New(v, &config.Options{})
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			xType := fmt.Sprintf("%T", config.Backend)
			if xType != tc.expectedType {
				t.Errorf("expected: %s, got: %s.", tc.expectedType, xType)
			}
		})

		if !strings.Contains(output, tc.expectedLog) {
			t.Errorf("Unexpected warning issued. Expected: %s, actual: %s", tc.expectedLog, output)
		}

		for k := range tc.environment {
			os.Unsetenv(k)
		}
	}
}

func TestNewConfigMissingParameter(t *testing.T) {
	testCases := []struct {
		environment  map[string]interface{}
		expectedType string
	}{
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "github",
				"AVP_GH_TOKEN":  "token",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "token",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "approle",
				"AVP_ROLEID":    "role_id",
				"AVP_SECRET_ID": "secret_id",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "k8s",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "userpass",
				"AVP_USERNAME":  "username",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "userpass",
				"AVP_PASSWORD":  "password",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "vault",
				"AVP_AUTH_TYPE": "userpass",
			},
			"*backends.Vault",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":        "ibmsecretsmanager",
				"AVP_IAM_API_KEY": "token",
			},
			"*backends.IBMSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":   "ibmsecretsmanager",
				"VAULT_ADDR": "http://vault",
			},
			"*backends.IBMSecretsManager",
		},
		{ //  WebIdentityEmptyRoleARNErr will occur if 'AWS_WEB_IDENTITY_TOKEN_FILE' was set but 'AWS_ROLE_ARN' was not set.
			map[string]interface{}{
				"AVP_TYPE":                    "awssecretsmanager",
				"AWS_REGION":                  "us-west-1",
				"AWS_WEB_IDENTITY_TOKEN_FILE": "/var/run/secrets/eks.amazonaws.com/serviceaccount/token",
			},
			"*backends.AWSSecretsManager",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":                   "yandexcloudlockbox",
				"AVP_YCL_KEY_ID":             "test",
				"AVP_YCL_SERVICE_ACCOUNT_ID": "test",
			},
			"*backends.YandexCloudLockbox",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":         "1passwordconnect",
				"OP_CONNECT_TOKEN": "token",
			},
			"*backends.OnePasswordConnect",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":         "delineasecretserver",
				"AVP_AUTH_TYPE":    "userpass",
				"AVP_DELINEA_URL":  "http://my-delinea-server",
				"AVP_DELINEA_USER": "username",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":             "delineasecretserver",
				"AVP_AUTH_TYPE":        "userpass",
				"AVP_DELINEA_URL":      "http://my-delinea-server",
				"AVP_DELINEA_PASSWORD": "password",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":             "delineasecretserver",
				"AVP_AUTH_TYPE":        "userpass",
				"AVP_DELINEA_USER":     "username",
				"AVP_DELINEA_PASSWORD": "password",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":         "delineasecretserver",
				"AVP_AUTH_TYPE":    "userpass",
				"AVP_DELINEA_USER": "username",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":             "delineasecretserver",
				"AVP_AUTH_TYPE":        "userpass",
				"AVP_DELINEA_PASSWORD": "password",
			},
			"*backends.DelineaSecretServer",
		},
		{
			map[string]interface{}{
				"AVP_TYPE":      "delineasecretserver",
				"AVP_AUTH_TYPE": "userpass",
			},
			"*backends.DelineaSecretServer",
		},
	}
	for _, tc := range testCases {
		for k, v := range tc.environment {
			os.Setenv(k, v.(string))
		}
		viper := viper.New()
		_, err := config.New(viper, &config.Options{})
		if err == nil {
			t.Fatalf("%s should not instantiate", tc.expectedType)
		}
		for k := range tc.environment {
			os.Unsetenv(k)
		}
	}
}

func TestExternalConfig(t *testing.T) {
	os.Setenv("AVP_TYPE", "vault")
	viper := viper.New()
	viper.SetDefault("VAULT_ADDR", "http://my-vault:8200/")
	config.New(viper, &config.Options{})
	if os.Getenv("VAULT_ADDR") != "http://my-vault:8200/" {
		t.Errorf("expected VAULT_ADDR env to be set from external config, was instead: %s", os.Getenv("VAULT_ADDR"))
	}
	os.Unsetenv("AVP_TYPE")
	os.Unsetenv("VAULT_ADDR")
}

const avpConfig = `AVP_TYPE: awssecretsmanager
AWS_ACCESS_KEY_ID: AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
AWS_REGION: us-west-2`

var expectedEnvVars = map[string]string{
	"AVP_TYPE":              "", // shouldn't be an env var
	"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
	"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	"AWS_REGION":            "us-west-2",
}

func TestExternalConfigAWS(t *testing.T) {
	// Test setting AWS_* env variables from external AVP config, note setting
	// env vars is necessary to pass AVP config entries to the AWS golang SDK
	tmpFile, err := os.CreateTemp("", "avpConfig.*.yaml")
	if err != nil {
		t.Errorf("Cannot create temporary file %s", err)
	}

	defer os.Remove(tmpFile.Name()) // clean up the file afterwards

	if _, err = tmpFile.WriteString(avpConfig); err != nil {
		t.Errorf("Failed to write to temporary file %s", err)
	}

	viper := viper.New()
	if _, err = config.New(viper, &config.Options{ConfigPath: tmpFile.Name()}); err != nil {
		t.Errorf("config.New returned error: %s", err)
	}

	if viper.GetString("AVP_TYPE") != "awssecretsmanager" {
		t.Errorf("expected AVP_TYPE to be set from external config, was instead: %s", viper.GetString("AVP_TYPE"))
	}

	for envVar, expected := range expectedEnvVars {
		if actual := os.Getenv(envVar); actual != expected {
			t.Errorf("expected %s env to be %s, was instead: %s", envVar, expected, actual)
		}
	}

	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_REGION")
}

func TestExternalConfigSOPS(t *testing.T) {
	const avpSOPSConfig = `AVP_TYPE: sops
SOPS_AGE_KEY_FILE: age`

	expectedSOPSEnvVars := map[string]string{
		"AVP_TYPE":          "", // shouldn't be an env var
		"SOPS_AGE_KEY_FILE": "age",
	}

	// Test setting SOPS_* env variables from external AVP config, note setting
	// env vars is necessary to pass AVP config entries to SOPS
	tmpFile, err := os.CreateTemp("", "avpSOPSConfig.*.yaml")
	if err != nil {
		t.Errorf("Cannot create temporary file %s", err)
	}

	defer os.Remove(tmpFile.Name()) // clean up the file afterwards

	if _, err = tmpFile.WriteString(avpSOPSConfig); err != nil {
		t.Errorf("Failed to write to temporary file %s", err)
	}

	viper := viper.New()
	if _, err = config.New(viper, &config.Options{ConfigPath: tmpFile.Name()}); err != nil {
		t.Errorf("config.New returned error: %s", err)
	}

	if viper.GetString("AVP_TYPE") != "sops" {
		t.Errorf("expected AVP_TYPE to be set from external config, was instead: %s", viper.GetString("AVP_TYPE"))
	}

	for envVar, expected := range expectedSOPSEnvVars {
		if actual := os.Getenv(envVar); actual != expected {
			t.Errorf("expected %s env to be %s, was instead: %s", envVar, expected, actual)
		}
	}

	os.Unsetenv("SOPS_AGE_KEY_FILE")
}

func TestNewConfigCyberArkSecretsManager(t *testing.T) {
	t.Run("Missing AVP_SECRETS_MANAGER_URL", func(t *testing.T) {
		os.Setenv("AVP_TYPE", "cyberarksecretsmanager")
		os.Setenv("AVP_SECRETS_MANAGER_ACCOUNT", "cyberark-account")
		os.Setenv("AVP_SECRETS_MANAGER_SSL_CERT", "cert")
		os.Setenv("AVP_SECRETS_MANAGER_TOKEN_FILE", "/path/to/token/file")
		defer os.Unsetenv("AVP_TYPE")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_ACCOUNT")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_SSL_CERT")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_TOKEN_FILE")

		v := viper.New()
		cfg, err := config.New(v, &config.Options{})

		if err == nil || cfg != nil {
			t.Fatal("Expected error due to missing AVP_SECRETS_MANAGER_URL, but got none")
		}
		if !strings.Contains(err.Error(), "required for CyberArk Secrets Manager") {
			t.Errorf("Expected error message to contain 'required for CyberArk Secrets Manager', but got: %s", err.Error())
		}
	})

	t.Run("Missing AVP_SECRETS_MANAGER_ACCOUNT", func(t *testing.T) {
		os.Setenv("AVP_TYPE", "cyberarksecretsmanager")
		os.Setenv("AVP_SECRETS_MANAGER_URL", "http://my-cyberark-url")
		os.Setenv("AVP_SECRETS_MANAGER_SSL_CERT", "cert")
		os.Setenv("AVP_SECRETS_MANAGER_TOKEN_FILE", "/path/to/token/file")
		defer os.Unsetenv("AVP_TYPE")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_URL")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_SSL_CERT")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_TOKEN_FILE")

		v := viper.New()
		cfg, err := config.New(v, &config.Options{})

		if err == nil || cfg != nil {
			t.Fatal("Expected error due to missing AVP_SECRETS_MANAGER_ACCOUNT, but got none")
		}
		if !strings.Contains(err.Error(), "required for CyberArk Secrets Manager") {
			t.Errorf("Expected error message to contain 'required for CyberArk Secrets Manager', but got: %s", err.Error())
		}
	})

	t.Run("Missing AVP_SECRETS_MANAGER_SSL_CERT", func(t *testing.T) {
		os.Setenv("AVP_TYPE", "cyberarksecretsmanager")
		os.Setenv("AVP_SECRETS_MANAGER_URL", "http://my-cyberark-url")
		os.Setenv("AVP_SECRETS_MANAGER_ACCOUNT", "cyberark-account")
		os.Setenv("AVP_SECRETS_MANAGER_TOKEN_FILE", "/path/to/token/file")
		defer os.Unsetenv("AVP_TYPE")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_URL")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_ACCOUNT")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_TOKEN_FILE")

		v := viper.New()
		cfg, err := config.New(v, &config.Options{})

		if err == nil || cfg != nil {
			t.Fatal("Expected error due to missing AVP_SECRETS_MANAGER_SSL_CERT, but got none")
		}
		if !strings.Contains(err.Error(), "required for CyberArk Secrets Manager") {
			t.Errorf("Expected error message to contain 'required for CyberArk Secrets Manager', but got: %s", err.Error())
		}
	})

	t.Run("Missing AVP_SECRETS_MANAGER_TOKEN_FILE", func(t *testing.T) {
		os.Setenv("AVP_TYPE", "cyberarksecretsmanager")
		os.Setenv("AVP_SECRETS_MANAGER_URL", "http://my-cyberark-url")
		os.Setenv("AVP_SECRETS_MANAGER_ACCOUNT", "cyberark-account")
		os.Setenv("AVP_SECRETS_MANAGER_SSL_CERT", "cert")
		defer os.Unsetenv("AVP_TYPE")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_URL")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_ACCOUNT")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_SSL_CERT")

		v := viper.New()
		cfg, err := config.New(v, &config.Options{})

		if err == nil || cfg != nil {
			t.Fatal("Expected error due to missing AVP_SECRETS_MANAGER_TOKEN_FILE, but got none")
		}
		if !strings.Contains(err.Error(), "required for CyberArk Secrets Manager") {
			t.Errorf("Expected error message to contain 'required for CyberArk Secrets Manager', but got: %s", err.Error())
		}
	})

	t.Run("Invalid certificate", func(t *testing.T) {
		os.Setenv("AVP_TYPE", "cyberarksecretsmanager")
		os.Setenv("AVP_SECRETS_MANAGER_URL", "http://my-cyberark-url")
		os.Setenv("AVP_SECRETS_MANAGER_ACCOUNT", "cyberark-account")
		os.Setenv("AVP_SECRETS_MANAGER_SSL_CERT", "cert")
		os.Setenv("AVP_SECRETS_MANAGER_TOKEN_FILE", "/path/to/token/file")
		defer os.Unsetenv("AVP_TYPE")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_URL")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_ACCOUNT")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_SSL_CERT")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_TOKEN_FILE")

		v := viper.New()
		cfg, err := config.New(v, &config.Options{})

		if err == nil || cfg != nil {
			t.Fatal("Expected error due to invalid certificate, but got none")
		}
		if !strings.Contains(err.Error(), "Can't append Secrets Manager SSL cert") {
			t.Errorf("Expected error message to contain 'Can't append Secrets Manager SSL cert', but got: %s", err.Error())
		}
	})

	t.Run("All required environment variables set and client creation", func(t *testing.T) {
		validSSLCert := `-----BEGIN CERTIFICATE-----
MIIFoTCCA4mgAwIBAgICEAAwDQYJKoZIhvcNAQELBQAwdjELMAkGA1UEBhMCVVMx
FjAUBgNVBAgMDU1hc3NhY2h1c2V0dHMxDzANBgNVBAcMBk5ld3RvbjERMA8GA1UE
CgwIQ3liZXJBcmsxDzANBgNVBAsMBkNvbmp1cjEaMBgGA1UEAwwRSW50ZXJtZWRp
YXRlIENBIDEwIBcNMjAwNDI0MTYzODM1WhgPMjEyMDAzMzExNjM4MzVaMHIxCzAJ
BgNVBAYTAlVTMRYwFAYDVQQIDA1NYXNzYWNodXNldHRzMQ8wDQYDVQQHDAZOZXd0
b24xETAPBgNVBAoMCEN5YmVyQXJrMQ8wDQYDVQQLDAZDb25qdXIxFjAUBgNVBAMM
DWNvbmp1ci1zZXJ2ZXIwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDO
9kp5mFGvwM5sFlhECmqbaL5DPQXzk9CYWkHltPV6phWMgH/6c52gDs3DCERsRaXj
WUrUp2dPtcF5E3qhWzNIlC/vwVuXGjvpLY43CVOGSqczoctgZMs9Q0NRcR3G/RLl
O8dLwfOdGvNZYg80bMx8xB2zs24rAe8pVtvOTvbFzXWoLkXOoLYdq2Ce/mgn2R5b
9FAUdjOFTtlXLLElHv4WFdOIDhyALf1Q2nBrykGw5wehWclOgpZG43rom0ppUqdN
M+i+1/Me8dNtPW+oewrmjZM9IpmN3nwe33sNRBTwuZYQfTiFOw4iValUU1yMt3OA
yqrrHzHDcK+Uq2/ShsGIHKU9e9Hxz6pCyx04wRMZ5Q6Z5QTrCcYThipOb3NU16fO
fF5KZ3StNdRcE/Bv6M7lTz+R69zNs6wOj5tx7AshVABPhJMEmmMuFjZxybRJYcET
L/m4Vxk/H2+D7zGNIbuTOF7htfUv0FWQDx6OY8cNk+ePbW8TJuVVvKWeRt5ApzRl
IdPvq+bTYAMuy6IsCnSsSKseuOFw7Y0x3HvdHhM408LaSHTIZ4AbQ+5/eKJQZqgS
n5mNo7PKKxvqki2pO7XsPQdHUJ4ZgHoNa54xKNJ0eMLNYGTcsfZ+HGhHP24ZQNaa
1PpYU0yCCGyQL02nJaYUvr77hnYgLXY2HIUXNsQPSwIDAQABozswOTA3BgNVHREE
MDAugg1jb25qdXItc2VydmVygh1jb25qdXItc2VydmVyLm15Y29tcGFueS5sb2Nh
bDANBgkqhkiG9w0BAQsFAAOCAgEASJguTyHmJdAWad7JsAPPxHAwBKf+KFGZMq7l
LdPq4fePlhuudsYhdhJv/PcfnZBpFgIZpiDWwi28HdRE7VqyvYSbBQUJ7snszJLL
XbK4V1ZiHT9+mhnD/qVcBG7mTzD3fF3A+CwzPhQze3ws2RYN+a8Ex8dozw4H2DRP
p1J79bDqLRbRslDzdZ7GK0htgj5FsEIA8IBXlerEAy3ZXIJzKdLKYjYu/10+2i0H
Tx3CaSQcMZaQL+JR2VklvPlNLnglcKwHr/R9rqJ6k49vO5VGOQ9A1I2HsWe7liBO
A4NAe1e8W2lJyklcCuFlEO9IFFBD1Ia+7rTeQcPuu+K4gxZcbZYohw/qmFcfyepy
j4V6Afu2mJir+pLWUl2Q/Y4GaUVNTN1J8FfC7ayhTU9fh0eU4VpcGq4+0HKcsSka
ikqg4bE950GlLYijOZUS+OJnMQbvhNuDhT0gJyBhafvutZLNBKG+PA735IRdPb6y
6boiqvclGag32KS4NR5rk0JRGYN4PLntj3FsAEadQKtc/1nsos0WenQRSs0187bP
3C50m0UvZZ8IP/bj2FJshgMs5HPQwUTNACuRi4kWTyQfeRjLYtKojHXFAfY9gqwF
fg/3pECLly8p8at6Ry6CicuOPSXWS+Mf0+74iVnBEg/UEDddVriK2E6/MPzPSIia
dC+b73B=
-----END CERTIFICATE-----`

		os.Setenv("AVP_TYPE", "cyberarksecretsmanager")
		os.Setenv("AVP_SECRETS_MANAGER_URL", "http://my-cyberark-url")
		os.Setenv("AVP_SECRETS_MANAGER_ACCOUNT", "cyberark-account")
		os.Setenv("AVP_SECRETS_MANAGER_SSL_CERT", validSSLCert)
		os.Setenv("AVP_SECRETS_MANAGER_TOKEN_FILE", "/path/to/token/file")
		defer os.Unsetenv("AVP_TYPE")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_URL")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_ACCOUNT")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_SSL_CERT")
		defer os.Unsetenv("AVP_SECRETS_MANAGER_TOKEN_FILE")

		v := viper.New()
		cfg, err := config.New(v, &config.Options{})

		if err != nil || cfg == nil {
			t.Errorf("Expected no error, but got: %s", err)
		}

		// Assert that environment variables are correctly read
		if v.GetString("AVP_TYPE") != "cyberarksecretsmanager" {
			t.Errorf("Expected AVP_TYPE to be 'cyberarksecretsmanager', but got: %s", v.GetString("AVP_TYPE"))
		}
		if v.GetString("AVP_SECRETS_MANAGER_URL") != "http://my-cyberark-url" {
			t.Errorf("Expected AVP_SECRETS_MANAGER_URL to be 'http://my-cyberark-url', but got: %s", v.GetString("AVP_SECRETS_MANAGER_URL"))
		}
		if v.GetString("AVP_SECRETS_MANAGER_ACCOUNT") != "cyberark-account" {
			t.Errorf("Expected AVP_SECRETS_MANAGER_ACCOUNT to be 'cyberark-account', but got: %s", v.GetString("AVP_SECRETS_MANAGER_ACCOUNT"))
		}
		if v.GetString("AVP_SECRETS_MANAGER_SSL_CERT") != validSSLCert {
			t.Errorf("Expected AVP_SECRETS_MANAGER_SSL_CERT to match validSSLCert")
		}
		if v.GetString("AVP_SECRETS_MANAGER_TOKEN_FILE") != "/path/to/token/file" {
			t.Errorf("Expected AVP_SECRETS_MANAGER_TOKEN_FILE to be '/path/to/token/file', but got: %s", v.GetString("AVP_SECRETS_MANAGER_TOKEN_FILE"))
		}
	})
}
