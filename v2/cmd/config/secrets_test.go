package config

import (
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafeas/voucher/v2/repository"
)

func TestNonExistantEjson(t *testing.T) {
	viper.Set("ejson.secrets", "../../../testdata/bad.ejson")
	viper.Set("ejson.dir", "../../../testdata/key")
	t.Cleanup(viper.Reset)

	_, err := ReadSecrets()
	require.Equal(
		t,
		err.Error(),
		"stat ../../../testdata/bad.ejson: no such file or directory",
		"did not fail appropriately, actual error is:",
		err,
	)
}

func TestGetRepositoryKeyRing(t *testing.T) {
	viper.Set("ejson.secrets", "../../../testdata/test.ejson")
	viper.Set("ejson.dir", "../../../testdata/key")
	t.Cleanup(viper.Reset)

	data, err := ReadSecrets()
	require.NoError(t, err)
	assert.Equal(t, repository.KeyRing{
		"organization-name": repository.Auth{
			Token: "asdf1234",
		},
		"organization2-name": repository.Auth{
			Username: "testUser",
			Password: "testPassword",
		},
	}, data.RepositoryAuthentication)
}

func TestGetRepositoryKeyRingNoEjson(t *testing.T) {
	viper.Set("ejson.secrets", "../../../testdata/test.ejson")
	viper.Set("ejson.dir", "../../../testdata/nokey")
	t.Cleanup(viper.Reset)

	data, err := ReadSecrets()
	require.Nil(t, data)
	assert.Error(t, err)
}

func TestGetPGPKeyRing(t *testing.T) {
	viper.Set("ejson.secrets", "../../../testdata/test.ejson")
	viper.Set("ejson.dir", "../../../testdata/key")
	t.Cleanup(viper.Reset)

	data, err := ReadSecrets()
	require.NoError(t, err)
	keyRing, err := data.getPGPKeyRing()
	require.NoError(t, err)
	assert.NotNil(t, keyRing)
}

func TestReadSops(t *testing.T) {
	viper.Set("sops.file", "../../../testdata/test.sops.json")
	t.Cleanup(viper.Reset)

	// Capture and restore GNUPGHOME variable
	existingHome := os.Getenv("GNUPGHOME")
	t.Cleanup(func() { os.Setenv("GNUPGHOME", existingHome) })

	// Overwrite GNUPGHOME, shell to GPG to load the test private key
	testHome := t.TempDir()
	os.Setenv("GNUPGHOME", testHome)
	cmd := exec.Command("gpg", "--import", "../../../testdata/testkey.asc")
	err := cmd.Run()
	require.NoError(t, err)
}
