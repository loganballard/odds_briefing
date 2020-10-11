package odds_briefing

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentialsFilePathExists(t *testing.T) {
	filePath := getCredsFilePath()
	assert.NotNil(t, filePath, "Can't get Credentials file path")
}

func TestCredentialsFileExists(t *testing.T) {
	filePath := getCredsFilePath()
	fileInfo, err := os.Stat(filePath)
	assert.Nil(t, err, "Credentials file doesn't exist at Credentials file path")
	assert.NotNil(t, fileInfo, "File cant be stat'd")
}

func TestCredentialsFileLoadingNoErrors(t *testing.T) {
	var credFile credentials
	credFile.loadCredentials()
}

func TestCredentialsFileHasAllNecessaryApiKeys(t *testing.T) {
	var credFile credentials
	credFile.loadCredentials()
	assert.NotEmpty(t, credFile.OddsApiKey, "missing odds api key")
	assert.NotEmpty(t, credFile.TwilioSid, "missing twilio sid")
	assert.NotEmpty(t, credFile.TwilioAuthKey, "missing twilio auth key")
}
