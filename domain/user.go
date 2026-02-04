package domain

type User struct {
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	Password  string `db:"password"`
	IsAdmin   bool   `db:"is_admin"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}
