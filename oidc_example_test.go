package oidc

import (
	"crypto/sha256"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zitadel/oidc/v3/example/server/storage"
	"github.com/zitadel/oidc/v3/pkg/op"
	"golang.org/x/text/language"
	"milthm.dev/translation-zitadel-oidc/util"
)

type OidcService struct {
	handler *op.Provider
}

const (
	testIssuer    = "https://localhost:9998/"
	pathLoggedOut = "/logged-out"
)

var (
	testConfig = &op.Config{
		CryptoKey:                sha256.Sum256([]byte("test")),
		DefaultLogoutRedirectURI: pathLoggedOut,
		CodeMethodS256:           true,
		AuthMethodPost:           true,
		AuthMethodPrivateKeyJWT:  true,
		GrantTypeRefreshToken:    true,
		RequestObjectSupported:   true,
		SupportedClaims:          op.DefaultSupportedClaims,
		SupportedUILocales:       []language.Tag{language.English},
		DeviceAuthorization: op.DeviceAuthorizationConfig{
			Lifetime:     5 * time.Minute,
			PollInterval: 5 * time.Second,
			UserFormPath: "/device",
			UserCode:     op.UserCodeBase20,
		},
	}
)

func newTestProvider(config *op.Config) (*op.Provider, error) {
	oidcStorage := storage.NewStorage(storage.NewUserStore(testIssuer))
	keySet := &op.OpenIDKeySet{Storage: oidcStorage}
	return op.NewProvider(config, oidcStorage,
		op.StaticIssuer(testIssuer),
		op.WithAllowInsecure(),
		op.WithAccessTokenKeySet(keySet),
		op.WithIDTokenHintKeySet(keySet),
	)
}

func Example() {
	srv := &OidcService{}
	var err error

	// Create an Oidc client
	srv.handler, err = newTestProvider(testConfig)
	if err != nil {
		panic(err)
	}

	// Serve a Request
}

func (s *Server) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	shadowWriter := utils.NewShadowResponseWriter()

	// Use oidc Serving
	s.handler.ServeHTTP(shadowWriter, request)

	// Process Error Handling
	processErrorPage := false
	if shadowWriter.IsError() &&
		strings.HasPrefix(shadowWriter.Header().Get("Content-Type"), "text/plain") &&
		strings.HasPrefix(request.Header.Get("Accept"), "text/html") {
		processErrorPage = true
	}

	if processErrorPage {
		errBody := shadowWriter.Body()
		errMsg := errBody.String()
		errorResponse := s.makeErrorPageResponse(errMsg)

		errHeader := shadowWriter.Header()
		responseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
		for k, v := range errHeader {
			if k == "Content-Type" {
				continue
			}
			for _, vv := range v {
				responseWriter.Header().Add(k, vv)
			}
		}

		responseWriter.WriteHeader(shadowWriter.StatusCode())
		_, err := responseWriter.Write(errorResponse)
		if err != nil {
			logx.Errorw("failed to write shadow response")
		}
	} else {
		err := shadowWriter.WriteTo(responseWriter)
		if err != nil {
			logx.Errorw("failed to write shadow response")
		}
	}
}

func (s *Server) makeErrorPageResponse(message string) []byte {
	return []byte("<!DOCTYPE html>" +
		"<html lang=\"en\">" +
		"<body>" +
		"<h1>" + message + "</h1>" +
		"</body>" +
		"</html>")
}
