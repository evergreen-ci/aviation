package aviation

import (
	"context"
	"testing"

	"github.com/evergreen-ci/gimlet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestAuthRequiredInterceptors(t *testing.T) {
	const (
		username       = "testUser"
		userAPIKey     = "123abc"
		headerUserName = "user"
		headerKeyName  = "key"
	)

	opts, err := gimlet.NewBasicUserOptions(username)
	require.NoError(t, err)
	opts = opts.Email("test@test.com").Key(userAPIKey)
	user := gimlet.NewBasicUser(opts)
	um, err := gimlet.NewBasicUserManager([]gimlet.BasicUser{*user}, nil)
	require.NoError(t, err)
	conf := gimlet.UserMiddlewareConfiguration{
		HeaderUserName: headerUserName,
		HeaderKeyName:  headerKeyName,
	}

	for _, test := range []struct {
		name string
		ctx  context.Context
		err  bool
	}{
		{
			name: "ValidAuth",
			ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
				headerUserName: {username},
				headerKeyName:  {userAPIKey},
			}),
		},
		{
			name: "MissingMetadata",
			ctx:  context.TODO(),
			err:  true,
		},
		{
			name: "MissingAPIKey",
			ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
				headerUserName: {username},
			}),
			err: true,
		},
		{
			name: "UserDNE",
			ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
				headerUserName: {"DNE"},
				headerKeyName:  {userAPIKey},
			}),
			err: true,
		},
		{
			name: "IncorrectAPIKey",
			ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
				headerUserName: {username},
				headerKeyName:  {"incorrect"},
			}),
			err: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Run("Unary", func(t *testing.T) {
				interceptor := MakeAuthenticationRequiredUnaryInterceptor(um, conf)
				_, err = interceptor(test.ctx, nil, nil, mockUnaryHandler)

				if test.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
			t.Run("Stream", func(t *testing.T) {
				interceptor := MakeAuthenticationRequiredStreamInterceptor(um, conf)
				err = interceptor(nil, &mockServerStream{ctx: test.ctx}, nil, mockStreamHandler)

				if test.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})

		})
	}
}
