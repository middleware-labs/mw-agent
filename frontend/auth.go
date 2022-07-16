package frontend

import (
	"context"
	"errors"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configauth"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type AuthDetails struct {
	account string
	api     *FrontendApi
	expires int64
	key     int64
}

type Auth struct {
	config *ConfigAuth
	logger *zap.Logger
	cache  map[string]*AuthDetails
	api    *FrontendApi
	m      sync.Mutex
}

func (a *Auth) Start(ctx context.Context, host component.Host) error {
	//TODO implement me
	log.Println("Auth component starting.")
	a.cache = make(map[string]*AuthDetails)
	a.api = &FrontendApi{
		Server: a.config.Server,
		Token:  a.config.Token,
		Logger: a.logger,
	}
	return nil
}

func (a *Auth) Shutdown(ctx context.Context) error {
	//TODO implement me
	a.logger.Info("Auth component Shutdown.")
	a.cache = nil
	return nil
}

func (a *Auth) Authenticate(ctx context.Context, headers map[string][]string) (context.Context, error) {
	a.logger.Info("Authenticating request")

	for key, value := range headers {
		if strings.ToLower(key) != key {
			if headers[strings.ToLower(key)] == nil {
				headers[strings.ToLower(key)] = []string{}
			}
			headers[strings.ToLower(key)] = append(headers[strings.ToLower(key)], value...)
			headers[key] = nil
		}
	}
	if headers["Authorization"] != nil {
		headers["authorization"] = headers["Authorization"]
	}
	//return ctx, nil
	if headers["authorization"] != nil {

		for _, header := range headers["authorization"] {

			a.m.Lock()
			defer a.m.Unlock()

			if a.cache[header] != nil {
				a.logger.Info("Authenticating request via cache")
				return context.WithValue(ctx, "auth", a.cache[header]), nil
			}
			resp, err := a.api.Request(http.MethodPost, "/auth", map[string]any{
				"token": header,
			})
			if err != nil {
				a.logger.Info("Authenticating request failed, failed to verify token")
				a.logger.Error("failed to verify token "+err.Error(), zap.Error(err))
				return nil, err
			}
			if resp["account"] != nil {
				_, ok := resp["account"].(string)
				if ok {
					key := rand.Int63()
					expires := resp["expires"].(float64)
					a.cache[header] = &AuthDetails{
						api:     a.api,
						account: resp["account"].(string),
						key:     key,
					}
					// delete entry after expire.
					go func() {
						time.Sleep(time.Duration(expires) * time.Second)
						if a.cache[header] != nil && a.cache[header].key == key {
							log.Println("deleting cache " + header)
							delete(a.cache, header)
						}
					}()
					a.logger.Info("Authenticating request via api")
					return context.WithValue(ctx, "auth", a.cache[header]), nil
				}
			}
		}
	} else {
		a.logger.Warn("not authorization header got")
	}

	a.logger.Info("Authenticating request failed")
	return nil, errors.New("failed to authorize request")
	//return ctx, nil
}

var _ configauth.ServerAuthenticator = (*Auth)(nil)

var (
	errInvalidAuthenticationHeaderFormat = errors.New("invalid authorization header format")
	errFailedToObtainClaimsFromToken     = errors.New("failed to get the subject from the token issued by the OIDC provider")
	errClaimNotFound                     = errors.New("username claim from the OIDC configuration not found on the token returned by the OIDC provider")
	errUsernameNotString                 = errors.New("the username returned by the OIDC provider isn't a regular string")
	errGroupsClaimNotFound               = errors.New("groups claim from the OIDC configuration not found on the token returned by the OIDC provider")
)

// Config specifies how the Per-RPC bearer token based authentication data should be obtained.
type ConfigAuth struct {
	config.ExtensionSettings `mapstructure:",squash"`

	// BearerToken specifies the bearer token to use for every RPC.
	Server string `mapstructure:"server,omitempty"`
	Token  string `mapstructure:"token,omitempty"`
}

var _ config.Extension = (*ConfigAuth)(nil)
var errNoTokenProvided = errors.New("no bearer token provided")

// Validate checks if the extension configuration is valid
func (cfg *ConfigAuth) Validate() error {

	log.Println("NewAuthFactory ValidateValidateValidate " + cfg.Server)
	log.Println("NewAuthFactory ValidateValidateValidate " + cfg.Token)
	if cfg.Server == "" || cfg.Token == "" {
		log.Println("NewAuthFactory error " + errNoTokenProvided.Error())
		return errNoTokenProvided
	}
	return nil
}
