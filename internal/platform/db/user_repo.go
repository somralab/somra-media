package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// UserAccount is a persisted login identity.
type UserAccount struct {
	ID           string
	Username     string
	PasswordHash string
	Disabled     bool
	Roles        []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserRepo persists user accounts and role assignments.
type UserRepo struct {
	q Querier
}

// NewUserRepo returns a repository bound to q.
func NewUserRepo(q Querier) *UserRepo {
	return &UserRepo{q: q}
}

var (
	ErrUserNotFound      = errors.New("db user: not found")
	ErrUserAlreadyExists = errors.New("db user: already exists")
)

// Count returns the number of registered users.
func (r *UserRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.q.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_account`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("db user count: %w", err)
	}
	return n, nil
}

// Create inserts a user with the given roles.
func (r *UserRepo) Create(ctx context.Context, id, username, passwordHash string, roleNames []string) (UserAccount, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return UserAccount{}, fmt.Errorf("db user create: username is required")
	}
	if passwordHash == "" {
		return UserAccount{}, fmt.Errorf("db user create: password hash is required")
	}
	if id == "" {
		return UserAccount{}, fmt.Errorf("db user create: id is required")
	}

	_, err := r.q.ExecContext(ctx, `
		INSERT INTO user_account (id, username, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, datetime('now'), datetime('now'))
	`, id, username, passwordHash)
	if err != nil {
		if isUniqueViolation(err) {
			return UserAccount{}, ErrUserAlreadyExists
		}
		return UserAccount{}, fmt.Errorf("db user create: %w", err)
	}

	if _, err := r.q.ExecContext(ctx, `
		INSERT INTO user_profile (user_id) VALUES (?)
	`, id); err != nil {
		return UserAccount{}, fmt.Errorf("db user create profile: %w", err)
	}

	for _, roleName := range roleNames {
		if err := r.assignRole(ctx, id, roleName); err != nil {
			return UserAccount{}, err
		}
	}

	return r.GetByID(ctx, id)
}

func (r *UserRepo) assignRole(ctx context.Context, userID, roleName string) error {
	res, err := r.q.ExecContext(ctx, `
		INSERT INTO user_role (user_id, role_id)
		SELECT ?, id FROM role WHERE name = ?
	`, userID, roleName)
	if err != nil {
		return fmt.Errorf("db user assign role %q: %w", roleName, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db user assign role rows: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("db user assign role %q: role not found", roleName)
	}
	return nil
}

// GetByID returns a user with roles.
func (r *UserRepo) GetByID(ctx context.Context, id string) (UserAccount, error) {
	var u UserAccount
	var disabled int
	var created, updated string
	err := r.q.QueryRowContext(ctx, `
		SELECT id, username, password_hash, disabled, created_at, updated_at
		FROM user_account WHERE id = ?
	`, id).Scan(&u.ID, &u.Username, &u.PasswordHash, &disabled, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return UserAccount{}, ErrUserNotFound
	}
	if err != nil {
		return UserAccount{}, fmt.Errorf("db user get by id: %w", err)
	}
	u.Disabled = disabled != 0
	u.CreatedAt = parseSQLiteTime(created)
	u.UpdatedAt = parseSQLiteTime(updated)
	u.Roles, err = r.loadRoles(ctx, id)
	if err != nil {
		return UserAccount{}, err
	}
	return u, nil
}

// GetByUsername returns a user with roles.
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (UserAccount, error) {
	var u UserAccount
	var disabled int
	var created, updated string
	err := r.q.QueryRowContext(ctx, `
		SELECT id, username, password_hash, disabled, created_at, updated_at
		FROM user_account WHERE username = ? COLLATE NOCASE
	`, strings.TrimSpace(username)).Scan(&u.ID, &u.Username, &u.PasswordHash, &disabled, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return UserAccount{}, ErrUserNotFound
	}
	if err != nil {
		return UserAccount{}, fmt.Errorf("db user get by username: %w", err)
	}
	u.Disabled = disabled != 0
	u.CreatedAt = parseSQLiteTime(created)
	u.UpdatedAt = parseSQLiteTime(updated)
	u.Roles, err = r.loadRoles(ctx, u.ID)
	if err != nil {
		return UserAccount{}, err
	}
	return u, nil
}

// List returns all users with roles.
func (r *UserRepo) List(ctx context.Context) ([]UserAccount, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT id, username, password_hash, disabled, created_at, updated_at
		FROM user_account ORDER BY username COLLATE NOCASE
	`)
	if err != nil {
		return nil, fmt.Errorf("db user list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []UserAccount
	for rows.Next() {
		var u UserAccount
		var disabled int
		var created, updated string
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &disabled, &created, &updated); err != nil {
			return nil, fmt.Errorf("db user list scan: %w", err)
		}
		u.Disabled = disabled != 0
		u.CreatedAt = parseSQLiteTime(created)
		u.UpdatedAt = parseSQLiteTime(updated)
		u.Roles, err = r.loadRoles(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// UpdatePassword replaces the password hash.
func (r *UserRepo) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	res, err := r.q.ExecContext(ctx, `
		UPDATE user_account SET password_hash = ?, updated_at = datetime('now') WHERE id = ?
	`, passwordHash, id)
	if err != nil {
		return fmt.Errorf("db user update password: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db user update password rows: %w", err)
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetDisabled toggles account disabled flag.
func (r *UserRepo) SetDisabled(ctx context.Context, id string, disabled bool) error {
	res, err := r.q.ExecContext(ctx, `
		UPDATE user_account SET disabled = ?, updated_at = datetime('now') WHERE id = ?
	`, boolToInt(disabled), id)
	if err != nil {
		return fmt.Errorf("db user set disabled: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db user set disabled rows: %w", err)
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetRoles replaces all role assignments for a user.
func (r *UserRepo) SetRoles(ctx context.Context, userID string, roleNames []string) error {
	if _, err := r.q.ExecContext(ctx, `DELETE FROM user_role WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("db user clear roles: %w", err)
	}
	for _, roleName := range roleNames {
		if err := r.assignRole(ctx, userID, roleName); err != nil {
			return err
		}
	}
	return nil
}

// PermissionsForUser returns deduplicated permission names for a user.
func (r *UserRepo) PermissionsForUser(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT DISTINCT p.name
		FROM user_role ur
		JOIN role_permission rp ON rp.role_id = ur.role_id
		JOIN permission p ON p.id = rp.permission_id
		WHERE ur.user_id = ?
		ORDER BY p.name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("db user permissions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var perms []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("db user permissions scan: %w", err)
		}
		perms = append(perms, name)
	}
	return perms, rows.Err()
}

func (r *UserRepo) loadRoles(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.q.QueryContext(ctx, `
		SELECT r.name FROM user_role ur
		JOIN role r ON r.id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY r.name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("db user load roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("db user load roles scan: %w", err)
		}
		roles = append(roles, name)
	}
	return roles, rows.Err()
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
