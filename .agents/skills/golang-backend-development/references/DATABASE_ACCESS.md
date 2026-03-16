# Database Access Examples

## ✅ Correct: Raw SQL with pgx

```go
// internal/repository/user_repository.go
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    query := `
        SELECT id, email, password, created_at, updated_at 
        FROM users 
        WHERE email = $1
    `
    
    var user models.User
    err := r.db.QueryRow(ctx, query, email).Scan(
        &user.ID,
        &user.Email,
        &user.Password,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
    
    if err == pgx.ErrNoRows {
        return nil, ErrUserNotFound
    }
    
    return &user, err
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
    query := `
        SELECT id, email, created_at 
        FROM users 
        ORDER BY created_at DESC 
        LIMIT $1 OFFSET $2
    `
    
    rows, err := r.db.Query(ctx, query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []*models.User
    for rows.Next() {
        var user models.User
        if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, &user)
    }
    
    return users, rows.Err()
}
```


## ❌ Incorrect: Using ORM

```go
// DON'T: Using GORM or other ORM
import "gorm.io/gorm"

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    result := r.db.Where("email = ?", email).First(&user)
    return &user, result.Error
}

// DON'T: Active Record pattern
func (u *User) Save() error {
    return db.Save(u).Error
}
```
