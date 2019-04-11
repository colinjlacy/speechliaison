package key_access

import (
	"fmt"
	"os"
)

func GetKey() []byte {
	return []byte(fmt.Sprintf(`{
		"type": "%s",
		"project_id": "%s",
		"private_key_id": "%s",
		"private_key": "%s",
		"client_email": "%s",
		"client_id": "%s",
		"auth_uri": "%s",
		"token_uri": "%s",
		"auth_provider_x509_cert_url": "%s",
		"client_x509_cert_url": "%s"
	}`,
		os.Getenv("TYPE"),
		os.Getenv("PROJECT_ID"),
		os.Getenv("PRIVATE_KEY_ID"),
		os.Getenv("PRIVATE_KEY"),
		os.Getenv("CLIENT_EMAIL"),
		os.Getenv("CLIENT_ID"),
		os.Getenv("AUTH_URI"),
		os.Getenv("TOKEN_URI"),
		os.Getenv("AUTH_PROVIDER_X509_CERT_URL"),
		os.Getenv("CLIENT_X509_CERT_URL"),
	))
}
