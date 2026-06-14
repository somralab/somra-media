package db

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrityCheck_OK(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	d, err := Open(ctx, newTestConfig(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })
	require.NoError(t, MigrateUp(ctx, d, nil))

	msg, err := IntegrityCheck(ctx, d)
	require.NoError(t, err)
	require.Equal(t, "ok", msg)
}

func TestInitialize_WithLogger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cfg := newTestConfig(t)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	d, err := Initialize(ctx, cfg, logger)
	require.NoError(t, err)
	t.Cleanup(func() { _ = d.Close() })

	current, target, err := MigrateStatus(ctx, d)
	require.NoError(t, err)
	require.Equal(t, target, current)
}

func TestUserRepo_SetRolesClearsAll(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewUserRepo(d.Querier())

	userID := uuid.NewString()
	_, err := repo.Create(ctx, userID, "kelly", "hash", []string{"admin", "user"})
	require.NoError(t, err)

	require.NoError(t, repo.SetRoles(ctx, userID, nil))
	got, err := repo.GetByID(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, got.Roles)
}

func TestMigrateUp_NilDB(t *testing.T) {
	t.Parallel()
	require.Error(t, MigrateUp(context.Background(), nil, nil))
	_, _, err := MigrateStatus(context.Background(), nil)
	require.Error(t, err)
}

func TestLoginAttemptRepo_RecordFailureWithoutLock(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewLoginAttemptRepo(d.Querier())

	la, err := repo.RecordFailure(ctx, "10.0.0.2", LoginAttemptIP, 0, time.Minute)
	require.NoError(t, err)
	assert.Equal(t, 1, la.FailedCount)
	require.Nil(t, la.LockedUntil)
}

func TestSessionRepo_CreateMinimal(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	users := NewUserRepo(d.Querier())
	sessions := NewSessionRepo(d.Querier())

	id := uuid.NewString()
	_, err := users.Create(ctx, id, "minimal", "hash", []string{"user"})
	require.NoError(t, err)

	sid := uuid.NewString()
	require.NoError(t, sessions.Create(ctx, SessionRecord{
		ID: sid, UserID: id, TokenHash: "minimal-hash", ExpiresAt: time.Now().Add(time.Hour),
	}))
	got, err := sessions.GetByID(ctx, sid)
	require.NoError(t, err)
	assert.Equal(t, id, got.UserID)
}

func TestUserRepo_AdminPermissions(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewUserRepo(d.Querier())

	userID := uuid.NewString()
	_, err := repo.Create(ctx, userID, "admin-user", "hash", []string{"admin"})
	require.NoError(t, err)

	perms, err := repo.PermissionsForUser(ctx, userID)
	require.NoError(t, err)
	require.Contains(t, perms, "users:manage")
	require.Contains(t, perms, "library:read")
}

func TestMediaRepo_UpsertFileLookupID(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	libRepo := NewLibraryRepo(d.Querier())
	mediaRepo := NewMediaRepo(d.Querier())
	dir := t.TempDir()
	lib, err := libRepo.Create(ctx, "Files", LibraryKindMovie, []string{dir}, false)
	require.NoError(t, err)
	itemID, err := mediaRepo.CreateItem(ctx, lib.ID, LibraryKindMovie, "Movie", nil)
	require.NoError(t, err)

	path := dir + "/movie.mkv"
	id1, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID, Path: path, FileName: "movie.mkv", SizeBytes: 10,
	})
	require.NoError(t, err)
	id2, err := mediaRepo.UpsertFile(ctx, MediaFile{
		LibraryID: lib.ID, MediaItemID: &itemID, Path: path, FileName: "movie.mkv", SizeBytes: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, id1, id2)
}

func TestLibraryRepo_MultiplePaths(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })

	repo := NewLibraryRepo(d.Querier())
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	lib, err := repo.Create(ctx, "Multi", LibraryKindMovie, []string{dir1, dir2}, true)
	require.NoError(t, err)
	require.Len(t, lib.Paths, 2)

	libs, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, libs, 1)
	require.Len(t, libs[0].Paths, 2)
}

func TestLoginAttemptRepo_ResetAfterFailures(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewLoginAttemptRepo(d.Querier())

	_, err := repo.RecordFailure(ctx, "1.2.3.4", LoginAttemptIP, 1, time.Minute)
	require.NoError(t, err)
	require.NoError(t, repo.Reset(ctx, "1.2.3.4", LoginAttemptIP))
	la, err := repo.Get(ctx, "1.2.3.4", LoginAttemptIP)
	require.NoError(t, err)
	assert.Equal(t, 0, la.FailedCount)
}

func TestUserRepo_DuplicateUsername(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewUserRepo(d.Querier())

	id1 := uuid.NewString()
	_, err := repo.Create(ctx, id1, "dup", "hash", []string{"user"})
	require.NoError(t, err)

	id2 := uuid.NewString()
	_, err = repo.Create(ctx, id2, "dup", "hash", []string{"user"})
	require.ErrorIs(t, err, ErrUserAlreadyExists)
}

func TestUserRepo_SetRolesInvalidRole(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewUserRepo(d.Querier())

	userID := uuid.NewString()
	_, err := repo.Create(ctx, userID, "roles", "hash", []string{"user"})
	require.NoError(t, err)

	err = repo.SetRoles(ctx, userID, []string{"does-not-exist"})
	require.Error(t, err)
}

func TestProfileRepo_GetNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewProfileRepo(d.Querier())

	_, err := repo.Get(ctx, "missing")
	require.ErrorIs(t, err, ErrProfileNotFound)
}

func TestSessionRepo_GetByIDNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewSessionRepo(d.Querier())

	_, err := repo.GetByID(ctx, "missing")
	require.ErrorIs(t, err, ErrSessionNotFound)
}

func TestLibraryRepo_DeleteNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewLibraryRepo(d.Querier())

	err := repo.Delete(ctx, 99999)
	require.ErrorIs(t, err, ErrLibraryNotFound)
}

func TestUserRepo_UpdatePasswordNotFound(t *testing.T) {
	ctx := context.Background()
	d := openTestDB(t)
	t.Cleanup(func() { _ = d.Close() })
	repo := NewUserRepo(d.Querier())

	err := repo.UpdatePassword(ctx, "missing", "hash")
	require.ErrorIs(t, err, ErrUserNotFound)
}
