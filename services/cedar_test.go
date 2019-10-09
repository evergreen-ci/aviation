package services

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDialCedar(t *testing.T) {
	ctx := context.TODO()
	client := &http.Client{Timeout: 5 * time.Minute}
	baseAddress := "cedar.mongodb.com"
	rpcPort := "7070"
	username := os.Getenv("LDAP_USER")
	password := os.Getenv("LDAP_PASSWORD")

	t.Run("ConnectToCedar", func(t *testing.T) {
		conn, err := DialCedar(ctx, client, baseAddress, rpcPort, username, password, 10)
		require.NoError(t, err)
		require.NotNil(t, conn)
		assert.NoError(t, conn.Close())
	})
	t.Run("IncorrectBaseAddress", func(t *testing.T) {
		conn, err := DialCedar(ctx, client, "cedar.mongo.com", rpcPort, username, password, 10)
		assert.Error(t, err)
		assert.Nil(t, conn)
	})
	t.Run("IncorrectUsernameAndPassword", func(t *testing.T) {
		conn, err := DialCedar(ctx, client, baseAddress, rpcPort, "bad_user", "bad_password", 10)
		assert.Error(t, err)
		assert.Nil(t, conn)
	})
}
