module github.com/evergreen-ci/aviation

go 1.16

require (
	github.com/evergreen-ci/gimlet v0.0.0-20220401151443-33c830c51cee
	github.com/jpillora/backoff v1.0.0
	github.com/mongodb/grip v0.0.0-20220401165023-6a1d9bb90c21
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.1
	google.golang.org/grpc v1.46.0
)

require github.com/phyber/negroni-gzip v1.0.0 // indirect
