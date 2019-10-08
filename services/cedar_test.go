package services

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDialCedar(t *testing.T) {
	ctx := context.TODO()
	httpAddress := "https://cedar.mongodb.com"
	rpcAddress := "cedar.mongodb.com:7070"
	username := os.Getenv("LDAP_USER")
	password := os.Getenv("LDAP_PASSWORD")

	t.Run("ConnectToCedar", func(t *testing.T) {
		conn, err := DialCedar(ctx, httpAddress, rpcAddress, username, password, 10)
		require.NoError(t, err)
		require.NotNil(t, conn)
		assert.NoError(t, conn.Close())

		// make sure we extra slash doesn't make a difference
		conn, err = DialCedar(ctx, httpAddress+"/", rpcAddress, username, password, 10)
		require.NoError(t, err)
		require.NotNil(t, conn)
		assert.NoError(t, conn.Close())
	})
	t.Run("IncorrectUsernameAndPassword", func(t *testing.T) {
		conn, err := DialCedar(ctx, httpAddress, rpcAddress, "bad_user", "bad_password", 10)
		assert.Error(t, err)
		assert.Nil(t, conn)
	})
}
