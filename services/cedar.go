package services

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/evergreen-ci/aviation"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type userCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// DialCedar is a convenience function for creating a RPC client connection
// with cedar via gRPC. The username and password are the LDAP credentials for
// the cedar service.
func DialCedar(ctx context.Context, httpAddress, rpcAddress, username, password string, retries int) (*grpc.ClientConn, error) {
	httpAddress = strings.TrimRight(httpAddress, "/")

	creds := &userCredentials{
		Username: username,
		Password: password,
	}
	credsPayload, err := json.Marshal(creds)
	if err != nil {
		return nil, errors.Wrap(err, "problem building credentials payload")
	}

	ca, err := makeCedarCertRequest(ctx, httpAddress+"/rest/v1/admin/ca", nil)
	if err != nil {
		return nil, errors.Wrap(err, "problem getting cedar root cert")
	}
	crt, err := makeCedarCertRequest(ctx, httpAddress+"/rest/v1/admin/users/certificate", bytes.NewBuffer(credsPayload))
	if err != nil {
		return nil, errors.Wrap(err, "problem getting cedar user cert")
	}
	key, err := makeCedarCertRequest(ctx, httpAddress+"/rest/v1/admin/users/certificate/key", bytes.NewBuffer(credsPayload))
	if err != nil {
		return nil, errors.Wrap(err, "problem getting cedar user key")
	}

	tlsConf, err := aviation.GetClientTLSConfig(ca, crt, key)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating TLS config")
	}

	return aviation.Dial(ctx, aviation.DialOptions{
		Address: rpcAddress,
		Retries: retries,
		TLSConf: tlsConf,
	})
}

func makeCedarCertRequest(ctx context.Context, url string, body io.Reader) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Minute}

	req, err := http.NewRequest(http.MethodGet, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating http request")
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "problem with request")
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "problem reading response")
	}

	if resp.StatusCode != http.StatusOK {
		return out, errors.Errorf("failed request with status code %d", resp.StatusCode)
	}

	return out, nil
}
