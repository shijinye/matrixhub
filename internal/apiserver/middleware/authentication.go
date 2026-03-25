// Copyright The MatrixHub Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/matrixhub-ai/matrixhub/internal/domain/user"
)

var publicMethods = map[string]bool{
	"/matrixhub.v1alpha1.Login/Login": true,
}

var noRenewCookieMethods = map[string]bool{
	"/matrixhub.v1alpha1.Login/Logout": true,
}

func AuthInterceptor(sessionRepo user.ISessionRepo) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		token := getTokenFromContext(ctx, sessionRepo)
		ctx, err := sessionRepo.Load(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
		}
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}
		if !sessionRepo.Exists(ctx, user.UsernameCtxKey.String()) {
			return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
		}
		if !noRenewCookieMethods[info.FullMethod] {
			if err = commitAndWriteSessionCookie(ctx, sessionRepo); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		// set username, user id in context
		ctx = context.WithValue(ctx, user.UsernameCtxKey, sessionRepo.GetString(ctx, user.UsernameCtxKey.String()))
		ctx = context.WithValue(ctx, user.UserIdCtxKey, sessionRepo.GetInt(ctx, user.UserIdCtxKey.String()))

		return handler(ctx, req)
	}
}

func getTokenFromContext(ctx context.Context, session user.ISessionRepo) (token string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}
	sessionCookie := session.GetSessionCookie()
	cookieHeader := firstMD(md, "grpcgateway-cookie", "cookie")
	if cookieHeader == "" {
		return
	}

	return extractSessionToken(cookieHeader, sessionCookie.Name)
}

func extractSessionToken(cookieHeader, name string) string {
	h := http.Header{}
	h.Add("Cookie", cookieHeader)
	c, err := (&http.Request{Header: h}).Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

func firstMD(md metadata.MD, keys ...string) string {
	for _, k := range keys {
		if vs := md.Get(k); len(vs) > 0 && vs[0] != "" {
			return vs[0]
		}
	}
	return ""
}

func commitAndWriteSessionCookie(ctx context.Context, session user.ISessionRepo) error {
	switch session.Status(ctx) {
	case scs.Modified:
		token, expiry, err := session.Commit(ctx)
		if err != nil {
			return err
		}

		return writeSessionCookie(ctx, session, token, expiry)
	case scs.Destroyed:
		return writeSessionCookie(ctx, session, "", time.Time{})
	default:
	}

	return nil
}

func writeSessionCookie(ctx context.Context, session user.ISessionRepo, token string, expiry time.Time) error {
	sessionCookie := session.GetSessionCookie()
	cookie := &http.Cookie{
		Value:       token,
		Name:        sessionCookie.Name,
		Domain:      sessionCookie.Domain,
		HttpOnly:    sessionCookie.HttpOnly,
		Path:        sessionCookie.Path,
		SameSite:    sessionCookie.SameSite,
		Secure:      sessionCookie.Secure,
		Partitioned: sessionCookie.Partitioned,
	}

	if expiry.IsZero() {
		cookie.Expires = time.Unix(1, 0)
		cookie.MaxAge = -1
	} else if sessionCookie.Persist || session.GetBool(ctx, "__rememberMe") {
		cookie.Expires = time.Unix(expiry.Unix()+1, 0)
		cookie.MaxAge = int(time.Until(expiry).Seconds() + 1)
	}

	return grpc.SetHeader(ctx, metadata.Pairs("set-cookie", cookie.String()))
}
