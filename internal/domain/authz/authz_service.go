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

package authz

import (
	"context"
	"strconv"

	"github.com/matrixhub-ai/matrixhub/internal/domain/user"
)

// IAuthzService permission verification service interface
type IAuthzService interface {
	// GetUserPermissions gets user's permission list in a project
	GetUserPermissions(ctx context.Context, userID string, projectID int) ([]Permission, error)

	// VerifyPlatformPermission verifies platform-level permission (gets user info from ctx)
	VerifyPlatformPermission(ctx context.Context, perm Permission) (bool, error)

	// VerifyProjectPermission verifies project-level permission
	VerifyProjectPermission(ctx context.Context, projectID int, perm Permission) (bool, error)
}

// IAuthzProjectRepo project repository interface required for permission verification
type IAuthzProjectRepo interface {
	GetUserProjectPermissions(ctx context.Context, userID string, projectID int) ([]Permission, error)
	GetUserPlatformPermissions(ctx context.Context, userID string) ([]Permission, error)
	GetProjectIDByName(ctx context.Context, name string) (int, error)
}

// AuthzService permission verification service
type AuthzService struct {
	projectRepo IAuthzProjectRepo
}

// NewAuthzService creates permission verification service
func NewAuthzService(projectRepo IAuthzProjectRepo) *AuthzService {
	return &AuthzService{
		projectRepo: projectRepo,
	}
}

// GetUserPermissions gets user's permission list in a project
func (s *AuthzService) GetUserPermissions(ctx context.Context, userID string, projectID int) ([]Permission, error) {
	platformPerms, err := s.projectRepo.GetUserPlatformPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	projectPerms, err := s.projectRepo.GetUserProjectPermissions(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	// Merge platform and project permissions
	return append(platformPerms, projectPerms...), nil
}

func (s *AuthzService) getUserIDFromCtx(ctx context.Context) (string, bool) {
	userIDValue := ctx.Value(user.UserIdCtxKey)
	if userIDValue == nil {
		return "", false
	}
	return strconv.Itoa(userIDValue.(int)), true
}

// VerifyPlatformPermission verifies platform-level permission (gets user info from ctx)
func (s *AuthzService) VerifyPlatformPermission(ctx context.Context, perm Permission) (bool, error) {
	userID, ok := s.getUserIDFromCtx(ctx)
	if !ok {
		return false, nil
	}

	// Get user's platform-level permissions
	permissions, err := s.projectRepo.GetUserPlatformPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	return MatchPermissions(permissions, perm), nil
}

// VerifyProjectPermission verifies project-level permission
func (s *AuthzService) VerifyProjectPermission(ctx context.Context, projectID int, perm Permission) (bool, error) {
	userID, ok := s.getUserIDFromCtx(ctx)
	if !ok {
		return false, nil
	}

	// Get user's permission list
	permissions, err := s.GetUserPermissions(ctx, userID, projectID)
	if err != nil {
		return false, err
	}

	// Check if there's matching permission using regex
	return MatchPermissions(permissions, perm), nil
}
