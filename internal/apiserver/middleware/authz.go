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
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"

	"github.com/matrixhub-ai/matrixhub/internal/domain/authz"
	"github.com/matrixhub-ai/matrixhub/internal/domain/user"
)

// GinAuthn is a Gin middleware that loads the session and injects user ID into the request context.
// It must run before Authz so that VerifyPlatformPermission can find the user ID.
func GinAuthn(sessionRepo user.ISessionRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionCookie := sessionRepo.GetSessionCookie()
		cookie, err := c.Request.Cookie(sessionCookie.Name)
		if err != nil || cookie.Value == "" {
			c.Next()
			return
		}

		ctx, err := sessionRepo.Load(c.Request.Context(), cookie.Value)
		if err != nil {
			c.Next()
			return
		}

		if !sessionRepo.Exists(ctx, user.UserIdCtxKey.String()) {
			c.Next()
			return
		}

		userID := sessionRepo.GetInt(ctx, user.UserIdCtxKey.String())
		ctx = context.WithValue(ctx, user.UserIdCtxKey, userID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

var (
	publicPaths = map[string]methods{
		"/healthz": {"GET"},

		// Login/Logout
		"/api/v1alpha1/login":  {"POST"},
		"/api/v1alpha1/logout": {"POST"},

		// Current user (operating own data, no platform permission needed)
		"/api/v1alpha1/current-user":                  {"GET"},
		"/api/v1alpha1/current-user/reset-password":   {"POST"},
		"/api/v1alpha1/current-user/access-tokens":    {"GET"},
		"/apis/v1alpha1/current-user/access-tokens":   {"POST"},
		"/apis/v1alpha1/current-user/access-tokens/*": {"DELETE"},
		"/api/v1alpha1/current-user/projects/role":    {"GET"},

		// Project management (project-level permission handled in Handler)
		"/api/v1alpha1/projects":                  {"GET", "POST"},
		"/api/v1alpha1/projects/*":                {"GET", "PUT", "DELETE"},
		"/api/v1alpha1/projects/*/members":        {"GET", "POST", "DELETE"},
		"/api/v1alpha1/projects/*/members/*/role": {"PUT"},

		// Models (project-level permission handled in Handler)
		"/api/v1alpha1/models":                {"GET", "POST"},
		"/api/v1alpha1/models/task-labels":    {"GET"},
		"/api/v1alpha1/models/library-labels": {"GET"},
		"/api/v1alpha1/models/*/*":            {"GET", "DELETE"},
		"/api/v1alpha1/models/*/*/revisions":  {"GET"},
		"/api/v1alpha1/models/*/*/commits":    {"GET"},
		"/api/v1alpha1/models/*/*/commits/*":  {"GET"},
		"/api/v1alpha1/models/*/*/tree":       {"GET"},
		"/api/v1alpha1/models/*/*/blob":       {"GET"},

		// Datasets (project-level permission handled in Handler)
		"/api/v1alpha1/datasets":               {"GET", "POST"},
		"/api/v1alpha1/datasets/task-labels":   {"GET"},
		"/api/v1alpha1/datasets/*/*":           {"GET", "DELETE"},
		"/api/v1alpha1/datasets/*/*/revisions": {"GET"},
		"/api/v1alpha1/datasets/*/*/commits":   {"GET"},
		"/api/v1alpha1/datasets/*/*/commits/*": {"GET"},
		"/api/v1alpha1/datasets/*/*/tree":      {"GET"},
		"/api/v1alpha1/datasets/*/*/blob":      {"GET"},
	}

	permissionPaths = map[string]map[string]methods{
		// User management
		authz.UserRead.String(): {
			"/api/v1alpha1/users":   {"GET"},
			"/api/v1alpha1/users/*": {"GET"},
		},
		authz.UserCreate.String(): {
			"/api/v1alpha1/users":             {"POST"},
			"/apis/v1alpha1/users/*/sysadmin": {"PUT"},
		},
		authz.UserDelete.String(): {
			"/api/v1alpha1/users/*": {"DELETE"},
		},
		authz.UserResetPassword.String(): {
			"/api/v1alpha1/users/*/reset-password": {"POST"},
		},

		// Registry management
		authz.RegistryRead.String(): {
			"/api/v1alpha1/registries":      {"GET"},
			"/api/v1alpha1/registries/*":    {"GET"},
			"/api/v1alpha1/registries/ping": {"POST"},
		},
		authz.RegistryCreate.String(): {
			"/api/v1alpha1/registries": {"POST"},
		},
		authz.RegistryUpdate.String(): {
			"/api/v1alpha1/registries/*": {"PUT"},
		},
		authz.RegistryDelete.String(): {
			"/api/v1alpha1/registries/*": {"DELETE"},
		},

		// Sync policy management
		authz.SyncRead.String(): {
			"/api/v1alpha1/sync-policies":              {"GET"},
			"/api/v1alpha1/sync-policies/*":            {"GET"},
			"/api/v1alpha1/sync-policies/*/sync-tasks": {"GET"},
		},
		authz.SyncCreate.String(): {
			"/api/v1alpha1/sync-policies":              {"POST"},
			"/api/v1alpha1/sync-policies/*/sync-tasks": {"POST"},
		},
		authz.SyncUpdate.String(): {
			"/api/v1alpha1/sync-policies/*":                   {"PUT"},
			"/api/v1alpha1/sync-policies/*/sync-tasks/*/stop": {"POST"},
		},
		authz.SyncDelete.String(): {
			"/api/v1alpha1/sync-policies/*": {"DELETE"},
		},
	}
)

type methods []string

func (ms methods) check(m string) bool {
	for _, v := range ms {
		if strings.EqualFold(v, m) {
			return true
		}
	}
	return false
}

type pathPermission struct {
	direct   map[string]*pathEntry
	wildcard []*pathEntry
}

type pathEntry struct {
	path         string
	regex        *regexp.Regexp
	methods      methods
	methodsPerms map[string][]string
}

type authorizer struct {
	public *pathPermission
	roles  *pathPermission
}

var authorizerInst authorizer

func init() {
	authorizerInst = authorizer{}

	publicPerms := make(map[string]*pathEntry)
	for k, v := range publicPaths {
		publicPerms[k] = &pathEntry{
			path:    k,
			methods: v,
		}
	}
	authorizerInst.public = parsePermission(publicPerms)

	perms := make(map[string]*pathEntry)
	for perm, paths := range permissionPaths {
		for p, ms := range paths {
			if _, ok := perms[p]; !ok {
				perms[p] = &pathEntry{
					path:         p,
					methodsPerms: make(map[string][]string),
				}
			}
			perms[p].merge(ms, perm)
		}
	}
	authorizerInst.roles = parsePermission(perms)
}

func (p *pathEntry) merge(ms methods, perm string) {
	for _, v := range ms {
		p.methods = append(p.methods, v)
		p.methodsPerms[v] = append(p.methodsPerms[v], perm)
	}
}

func parsePermission(ms map[string]*pathEntry) *pathPermission {
	p := &pathPermission{
		direct: make(map[string]*pathEntry),
	}
	for k, v := range ms {
		if !strings.Contains(k, "*") {
			p.direct[k] = v
		} else {
			v.regex = regexp.MustCompile(strings.ReplaceAll(k, "*", "[^*/]+") + "$")
			p.wildcard = append(p.wildcard, v)
		}
	}
	return p
}

func (a *authorizer) publicCheck(path, method string) bool {
	ms, ok := a.public.direct[path]
	if ok && ms.methods.check(method) {
		return true
	}
	for _, v := range a.public.wildcard {
		if v.regex.MatchString(path) && v.methods.check(method) {
			return true
		}
	}
	return false
}

func (a *authorizer) getPerms(path, method string) (out []string) {
	if ms, ok := a.roles.direct[path]; ok && ms.methods.check(method) {
		out = append(out, ms.methodsPerms[method]...)
		return
	}
	for _, v := range a.roles.wildcard {
		if v.regex.MatchString(path) && v.methods.check(method) {
			out = append(out, v.methodsPerms[method]...)
		}
	}
	return
}

type deniedErr struct {
	Code    codes.Code `json:"code"`
	Message string     `json:"message"`
}

// Authz Gin platform-level permission middleware
func Authz(verifyFunc func(ctx context.Context, perm authz.Permission) (bool, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Writer.Status() == http.StatusNotFound || c.Writer.Status() == http.StatusMethodNotAllowed {
			c.Abort()
			return
		}

		p := c.Request.URL.Path
		method := c.Request.Method

		// Public paths pass directly
		if authorizerInst.publicCheck(p, method) {
			c.Next()
			return
		}

		// Get permission for the path
		perms := authorizerInst.getPerms(p, method)
		for _, perm := range perms {
			allowed, err := verifyFunc(c.Request.Context(), authz.Permission(perm))
			if err == nil && allowed {
				c.Next()
				return
			}
		}

		// No permission path matched, pass by default (may be unregistered new API)
		if len(perms) == 0 {
			c.Next()
			return
		}

		// Permission check failed
		c.JSON(http.StatusForbidden, deniedErr{
			Code:    codes.PermissionDenied,
			Message: "permission denied",
		})
		c.Abort()
	}
}
