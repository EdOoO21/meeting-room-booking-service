package ports

// PasswordHasher хеширует и проверяет пароли пользователей.
type PasswordHasher interface {
	// Hash строит безопасный хеш пароля.
	Hash(password string) (string, error)
	// Compare проверяет, соответствует ли пароль сохраненному хешу.
	Compare(hash, password string) error
}
