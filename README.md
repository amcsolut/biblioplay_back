# API Backend Infinitrum - Go Version

Este é um backend API desenvolvido em Go (Golang) usando o framework Gin.

## Arquitetura

O projeto segue uma arquitetura em camadas (Clean Architecture) com separação clara de responsabilidades:

```
├── main.go                 # Ponto de entrada da aplicação
├── go.mod                  # Dependências Go
├── config/                 # Configurações da aplicação
├── api/v1/                 # API versão 1
│   ├── handlers/           # Controladores HTTP (equivalente aos controllers)
│   ├── middleware/         # Middlewares (autenticação, CORS, etc.)
│   ├── repositories/       # Camada de acesso a dados
│   ├── dto/               # Data Transfer Objects (equivalente aos schemas)
│   ├── services/          # Lógica de negócio
│   └── routes.go          # Definição de rotas
├── internal/              # Código interno da aplicação
│   ├── models/            # Modelos de banco de dados (GORM)
│   └── database/          # Configuração e conexão com banco
├── migrations/            # Migrações de banco de dados
└── cmd/                   # Comandos adicionais
```

## Tecnologias Utilizadas

- **Go 1.21+** - Linguagem de programação
- **Gin** - Framework web HTTP
- **GORM** - ORM para Go
- **PostgreSQL/SQL Server** - Banco de dados
- **JWT** - Autenticação
- **Docker** - Containerização
- **golang-migrate** - Migrações de banco

## Domínios da Aplicação

- **User** - Gerenciamento de usuários e autenticação
- **Organization** - Organizações e usuários vinculados
- **Pop** - Sistema de cards/kanban
- **Tag** - Sistema de tags
- **Permissions** - Controle de permissões
- **Export** - Funcionalidades de exportação
- **Custom Fields** - Campos customizáveis

## Configuração

1. **Clone o repositório**
   ```bash
   git clone <repository-url>
   cd api-backend-go-infinitrum
   ```

2. **Configure as variáveis de ambiente**
   ```bash
   cp .env.example .env
   # Edite o arquivo .env com suas configurações
   ```

3. **Instale as dependências**
   ```bash
   go mod tidy
   ```

4. **Execute as migrações**
   ```bash
   # Usando golang-migrate (instale primeiro)
   migrate -path migrations -database "your-database-url" up
   ```

5. **Execute a aplicação**
   ```bash
   go run main.go
   ```

## Docker

Para executar com Docker:

```bash
# Usando docker-compose (recomendado)
docker-compose up -d

# Ou construindo manualmente
docker build -t infinitrum-api .
docker run -p 8080:8080 infinitrum-api
```

## Endpoints Principais

### Autenticação
- `POST /api/v1/auth/login` - Login do usuário
- `POST /api/v1/auth/register` - Registro de usuário
- `POST /api/v1/auth/refresh` - Renovar token

### Usuários
- `GET /api/v1/users` - Listar usuários
- `GET /api/v1/users/:id` - Obter usuário
- `PUT /api/v1/users/:id` - Atualizar usuário
- `DELETE /api/v1/users/:id` - Deletar usuário

### Organizações
- `GET /api/v1/organizations` - Listar organizações
- `POST /api/v1/organizations` - Criar organização
- `GET /api/v1/organizations/:id` - Obter organização

### Pops (Cards/Kanban)
- `GET /api/v1/pops` - Listar pops
- `POST /api/v1/pops` - Criar pop
- `GET /api/v1/pops/:id/cards` - Listar cards do pop

## Estrutura de Dados

### Modelos Principais

- **User**: Usuários do sistema
- **Organization**: Organizações
- **OrganizationUser**: Relacionamento usuário-organização
- **Pop**: Quadros kanban
- **PopCard**: Cards dos quadros
- **PopPhase**: Fases/colunas dos quadros
- **PopField**: Campos customizados
- **Tag**: Tags do sistema

## Desenvolvimento

### Padrões Utilizados

- **Repository Pattern**: Camada de acesso a dados
- **Service Pattern**: Lógica de negócio
- **DTO Pattern**: Transferência de dados
- **Middleware Pattern**: Interceptadores de requisições

### Estrutura de Resposta

```json
{
  "data": {},
  "message": "Success",
  "status": 200
}
```

### Autenticação

A API utiliza JWT (JSON Web Tokens) para autenticação. Inclua o token no header:

```
Authorization: Bearer <your-jwt-token>
```

## Migrações

Para criar uma nova migração:

```bash
migrate create -ext sql -dir migrations -seq migration_name
```

## Testes

```bash
# Executar todos os testes
go test ./...

# Executar testes com coverage
go test -cover ./...
```

## Contribuição

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## Comparação com a Versão Python

| Aspecto | Python/FastAPI | Go/Gin |
|---------|----------------|---------|
| Framework | FastAPI | Gin |
| ORM | SQLAlchemy | GORM |
| Migrações | Alembic | golang-migrate |
| Validação | Pydantic | struct tags + validator |
| Performance | ~1000 req/s | ~5000 req/s |
| Memória | ~50MB | ~15MB |
| Build | Interpretado | Compilado |

## Próximos Passos

- [ ] Implementar todos os handlers
- [ ] Adicionar testes unitários
- [ ] Implementar cache com Redis
- [ ] Adicionar documentação Swagger
- [ ] Implementar rate limiting
- [ ] Adicionar logs estruturados
- [ ] Implementar health checks

