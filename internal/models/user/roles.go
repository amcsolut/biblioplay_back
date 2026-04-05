package user

const (
	// RoleLevelMember é o nível básico de membro (leitor/ouvinte)
	RoleLevelMember = 1
	// RoleLevelAuthor é o nível de autor (catálogo, comunidade como criador)
	RoleLevelAuthor = 2
	// RoleLevelAdmin é o nível máximo de administrador
	RoleLevelAdmin = 9
	// Espaço para níveis intermediários: 3, 4, 5, 6, 7, 8
)

// IsAdmin verifica se o usuário é administrador
func (u *User) IsAdmin() bool {
	return u.RoleLevel >= RoleLevelAdmin
}

// HasMinimumLevel verifica se o usuário tem o nível mínimo requerido
func (u *User) HasMinimumLevel(requiredLevel int) bool {
	return u.RoleLevel >= requiredLevel
}

