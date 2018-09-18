package authn

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/token"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type AuthInfoResolverFactory interface {
	NewResolver(context.Context, config.TenantConfiguration) AuthInfoResolver
}

type AuthInfoResolver interface {
	Resolve(*http.Request) (auth.AuthInfo, error)
}

type StatefulJWTAuthInfoResolverFactory struct {
	handler.ProviderGraph
}

func (f StatefulJWTAuthInfoResolverFactory) NewResolver(ctx context.Context, tenantConfig config.TenantConfiguration) AuthInfoResolver {
	r := &StatefulJWTAuthInfoResolver{}
	inject.DefaultInject(r, f.ProviderGraph, ctx, tenantConfig)
	return r
}

type StatefulJWTAuthInfoResolver struct {
	token.TokenStore   `dependency:"TokenStore"`
	auth.AuthInfoStore `dependency:"AuthInfoStore"`
}

func (r StatefulJWTAuthInfoResolver) Resolve(req *http.Request) (authInfo auth.AuthInfo, err error) {
	keyType := model.GetAccessKeyType(req)

	var resolver AuthInfoResolver
	if keyType == model.MasterAccessKey {
		resolver = masterkeyAuthInfoResolver{
			TokenStore:    r.TokenStore,
			AuthInfoStore: r.AuthInfoStore,
		}
	} else {
		resolver = nonMasterkeyAuthInfoResolver{
			TokenStore:    r.TokenStore,
			AuthInfoStore: r.AuthInfoStore,
		}
	}

	authInfo, err = resolver.Resolve(req)
	authInfo.AccessKeyType = keyType

	return
}

func GetAccessToken(r *http.Request) string {
	return r.Header.Get("X-Skygear-Access-Token")
}

type masterkeyAuthInfoResolver struct {
	token.TokenStore   `dependency:"TokenStore"`
	auth.AuthInfoStore `dependency:"AuthInfoStore"`
}

func (r masterkeyAuthInfoResolver) Resolve(req *http.Request) (authInfo auth.AuthInfo, err error) {
	tokenStr := GetAccessToken(req)
	token := &token.Token{}
	r.TokenStore.Get(tokenStr, token)

	if token.AuthInfoID == "" {
		token.AuthInfoID = "_god"
	}

	info := &skydb.AuthInfo{}
	if err = r.AuthInfoStore.GetAuth(token.AuthInfoID, info); err == skydb.ErrUserNotFound {
		info.ID = token.AuthInfoID

		if err = r.AuthInfoStore.CreateAuth(info); err == skydb.ErrUserDuplicated {
			// user already exists, error can be ignored
			err = nil
		}
	}

	authInfo.AuthInfo = info

	return
}

type nonMasterkeyAuthInfoResolver struct {
	token.TokenStore   `dependency:"TokenStore"`
	auth.AuthInfoStore `dependency:"AuthInfoStore"`
}

func (r nonMasterkeyAuthInfoResolver) Resolve(req *http.Request) (authInfo auth.AuthInfo, err error) {
	tokenStr := GetAccessToken(req)

	tkn := &token.Token{}
	err = r.TokenStore.Get(tokenStr, tkn)
	if err != nil {
		// TODO:
		// handle error properly
		return
	}

	authInfo.Token = tkn

	info := &skydb.AuthInfo{}
	err = r.AuthInfoStore.GetAuth(tkn.AuthInfoID, info)
	if err != nil {
		// TODO:
		// handle error properly
		return
	}

	authInfo.AuthInfo = info

	return
}
