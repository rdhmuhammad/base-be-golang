# Base Backend Golang

[![GitHub stars](https://img.shields.io/github/stars/rdhmuhammad/base-be-golang)](https://github.com/rdhmuhammad/base-be-golang/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/rdhmuhammad/base-be-golang?)](https://github.com/rdhmuhammad/base-be-golang/network/members)

***

> This repository is a skeleton for backend projects. The project tries to keep the business layer technology-agnostic, while the infrastructure layer intentionally uses practical tools such as Gin, GORM, Redis, MinIO, and Sentry.

## 🚀 Tech Stack

![Go](https://img.shields.io/badge/Language-V1.23.0-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Framework-Gin_Gonic-00ADD8?style=for-the-badge&logo=gin&logoColor=white)
![GORM](https://img.shields.io/badge/ORM-GORM-00ADD8?style=for-the-badge)
![MySQL](https://img.shields.io/badge/Database-MySQL-4479A1?style=for-the-badge&logo=mysql&logoColor=white)
![MongoDB](https://img.shields.io/badge/Database-MongoDB-47A248?style=for-the-badge&logo=mongodb&logoColor=white)
![Redis](https://img.shields.io/badge/Cache-Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white)
![MinIO](https://img.shields.io/badge/Object%20Storage-MinIO-C72E49?style=for-the-badge&logo=minio&logoColor=white)

## 📚 Table of Contents

<details>
<summary>Click to expand</summary>

- 🚀 [Getting Started](#-getting-started)
  - 📦 [Prerequisites](#-prerequisites)
  - ▶️ [How to Run](#-how-to-run)
- 🛠️ [Development Guide](#-development-guide)
  - ➕ [Creating New Endpoint](#-creating-new-endpoint)
  - 🔩 [Using Generic Repository](#-using-generic-repository)
  - 🔁 [Using DB Transaction](#-using-db-transaction)
  - 🔒 [Security Guide](#-security-guide)
  - 🧩 [Project Structure](#-project-structure)
- 🚢 [Deployment](#-deployment)
  - 🏗️ [Build Binary](#-build-binary)
  - ⚙️ [Run with systemd](#-run-with-systemd)
- 📖 [Additional Information](#-additional-information)

</details>

## 🚀 Getting Started

### 📦 Prerequisites

Install these tools before running the application:

- Go `1.23.x`. The workspace currently uses `go 1.23.10`.
- MySQL, used by `pkg/db`.
- Redis, used by session cache and the available idempotent middleware.
- MinIO or an S3-compatible endpoint, because `shared/api.Default()` initializes MinIO on boot.
- SMTP credentials if email verification is enabled.
- Git, Make, and a shell that can run Go commands.

This repository uses a Go workspace:

```txt
.
./iam_module
./crudgen_module
```

The root module is `base-be-golang`. The IAM module is a separate workspace module imported as `github.com/rdhmuhammad/base-be-golang/iam-module`.

Create an environment file before running the API. The default entrypoint reads `.env.stag` unless another file is passed with `-env`.

Minimal local environment shape:

```env
APP_PORT=8999
ENVIRONMENT=development

MYSQL_USER=root
MYSQL_PASSWORD=password
MYSQL_DATABASE=base_be
MYSQL_PORT=3306
MYSQL_HOST_DEV=127.0.0.1
MYSQL_HOST_STAG_DOCKER=mysql
MYSQL_HOST_DOCKER=mysql
DB_LOG_MODE=2

REDIS_HOST=127.0.0.1:6379
REDIS_PASSWORD=

MINIO_ENDPOINT=127.0.0.1:9000
MINIO_BUCKET=base-be
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

IAM_MODULE_OFF=false
SECRET=change-this-jwt-secret
EXPIRED_TOKEN_JWT=24
SECRET_USER_ID=change-this-user-secret
ENCRYPT_MESSAGE_PASSWORD=change-this-32-byte-key
FALLBACK_TIMEZONE=Asia/Jakarta
FALLBACK_LANG=id

EMAIL_VERIFICATION_OFF=true
HOTP_SECRET=change-this-otp-secret
EXPARATION_OTP_TIME=5
FRONT_END_HOST=http://localhost:3000
SMPT_SERVER_HOST=smtp.example.com
SMPT_SERVER_PORT=587
SUPPORT_EMAIL=support@example.com
SUPPORT_EMAIL_PASS=password

SENTRY_DSN=
SENTRY_ENVIRONMENT=development
LOG_LEVEL=debug
```

Use the exact names currently used in the code. Some names are intentionally kept as they exist in the project, for example `EXPARATION_OTP_TIME` and `SMPT_SERVER_HOST`.

### ▶️ How to Run

Download dependencies:

```bash
go work sync
go mod download
```

Run the API:

```bash
go run ./cmd/api -env .env.stag
```

Run with another environment file:

```bash
go run ./cmd/api -env .env.dev
```

The server starts Gin on:

```txt
http://localhost:${APP_PORT}/api/v1
```

Boot flow:

1. `cmd/api/api.go` loads the selected env file.
2. `api.Default()` initializes Gin, middleware, MySQL, Redis, and MinIO.
3. Controllers are registered with `start.Register(...)`.
4. `start.Start()` mounts every router under `/api/v1`.

## 🛠️ Development Guide

### ➕ Creating New Endpoint

Every HTTP feature follows this flow:

```txt
cmd/api -> controller -> usecase -> repository -> domain
```

The controller handles HTTP binding and response mapping. The usecase owns business rules. The repository owns database details. The domain owns entities and table names.

#### 1. Create the domain

Place business entities in `internal/core/domain` for the main app, or inside the module's own `internal/core/domain` when working in a module.

```go
package domain

type Article struct {
    ID    uint   `gorm:"primaryKey" json:"id"`
    Title string `gorm:"column:title" json:"title"`
    Body  string `gorm:"column:body" json:"body"`
}

func (Article) TableName() string {
    return "articles"
}
```

The generic repository requires the model to implement `schema.Tabler`, which means it must have `TableName() string`.

#### 2. Create the usecase and constructor

Usecases receive infrastructure through `base.Port` and receive the database connection from the API registration layer.

Create request/response DTOs close to the usecase:

```go
package article

type CreateArticleRequest struct {
    Title string `json:"title" validate:"required"`
    Body  string `json:"body" validate:"required"`
}
```

```go
package article

import (
    "base-be-golang/internal/core/domain"
    "base-be-golang/pkg/db"
    "base-be-golang/shared/base"
    "context"

    "gorm.io/gorm"
)

type Usecase struct {
    base.Port
    articleRepo db.GenericRepository[domain.Article]
}

func NewUsecase(dbConn *gorm.DB, port base.Port) Usecase {
    return Usecase{
        Port:        port,
        articleRepo: db.NewGenericeRepo(dbConn, domain.Article{}),
    }
}

func (u Usecase) Create(ctx context.Context, request CreateArticleRequest) error {
    article := domain.Article{
        Title: request.Title,
        Body:  request.Body,
    }

    _, err := u.articleRepo.Store(ctx, article)
    if err != nil {
        return u.ErrHandler.ErrorReturn(err)
    }

    return nil
}
```

Constructor rule:

- Use `NewUsecase(dbConn *gorm.DB, port base.Port)`.
- Store repositories in the usecase.
- Use `base.Port` for cross-cutting services such as `Security`, `Cache`, `Env`, `Clock`, `Davinci`, `Mailing`, and `ErrHandler`.
- Keep Gin out of the usecase.

#### 3. Create the controller and constructor

Controllers embed `base.BaseController`, define a small usecase interface, bind input, call the usecase, and return with `ctrl.Mapper.NewResponse`.

```go
package controller

import (
    "base-be-golang/internal/core/usecase/article"
    "base-be-golang/shared/base"
    "base-be-golang/shared/payload"
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type ArticleController struct {
    base.BaseController
    uc ArticleUsecase
}

type ArticleUsecase interface {
    Create(ctx context.Context, request article.CreateArticleRequest) error
}

func NewArticleController(
    dbConn *gorm.DB,
    port base.Port,
    controller base.BaseController,
) ArticleController {
    return ArticleController{
        BaseController: controller,
        uc:             article.NewUsecase(dbConn, port),
    }
}

func (ctrl ArticleController) Create(c *gin.Context) {
    var request article.CreateArticleRequest
    if errs := ctrl.Enigma.BindAndValidate(c, &request); len(errs) > 0 {
        c.JSON(http.StatusBadRequest, payload.DefaultInvalidInputFormResponse(errs))
        return
    }

    err := ctrl.uc.Create(c.Request.Context(), request)
    ctrl.Mapper.NewResponse(c, payload.NewSuccessResponseNoData("CreateArticleSuccess"), err)
}

func (ctrl ArticleController) Route(router *gin.RouterGroup) {
    articleRouter := router.Group("/articles")
    articleRouter.POST(
        "",
        ctrl.Security.Validate(),
        ctrl.Security.Authorize("ADMIN"),
        ctrl.Create,
    )
}
```

Constructor rule:

- Use `NewXController(dbConn *gorm.DB, port base.Port, controller base.BaseController)`.
- Return a value that implements `api.Router`.
- Implement `Route(handler *gin.RouterGroup)`.
- Register middleware in the route method.

#### 4. Register the controller to `cmd/api`

Open `cmd/api/api.go`, import the controller package, then register it in the `BUSINESS MODULE` section.

```go
start.Register(func(dbConn *gorm.DB, port base.Port, ctrl base.BaseController) api.Router {
    return controller.NewArticleController(dbConn, port, ctrl)
})
```

`api.Register` injects:

- `*gorm.DB`
- `base.Port`
- `base.BaseController`

Do not manually create these dependencies inside `cmd/api` unless the application bootstrap changes.

### 🔩 Using Generic Repository

The generic repository lives in `pkg/db/generic_repository.go`.

Create a repository:

```go
articleRepo := db.NewGenericeRepo(dbConn, domain.Article{})
```

Or keep a pointer when another component needs to replace the connection, for example inside a transaction:

```go
articleRepo := db.NewGenericeRepoPointr(dbConn, domain.Article{})
```

Common CRUD usage:

```go
created, err := articleRepo.Store(ctx, article)
items, err := articleRepo.FindAll(ctx)
detail, err := articleRepo.FindOneByID(ctx, id)
err = articleRepo.UpdateSelectedCols(ctx, detail, "title", "body")
err = articleRepo.DeleteByID(ctx, id)
```

Query helpers:

```go
conditions := db.Query(
    db.Equal("published", "status"),
    db.Search("golang", "title", "body"),
)

items, err := articleRepo.FindAllByExpression(ctx, conditions)
```

Array filters:

```go
conditions := db.Query(
    db.InArray([]uint{1, 2, 3}, "id"),
)
```

Pagination:

```go
items, total, err := articleRepo.FindPagedByExpression(
    ctx,
    conditions,
    db.PaginationQuery{
        PerPage: request.PerPage,
        Page:    request.Page,
    },
)
```

Joins and preloads:

```go
detail, err := articleRepo.FindOneByExpressionAndJoin(
    ctx,
    db.Query(db.Equal(id, "articles.id")),
    []string{"Author"},
    []string{"Comments"},
)
```

Select only specific columns by passing a pointer to a projection struct. Every selected field should have a `gorm:"column:<name>"` tag.

```go
type ArticleSummary struct {
    ID    uint   `gorm:"column:id" json:"id"`
    Title string `gorm:"column:title" json:"title"`
}

var summary ArticleSummary
err := articleRepo.FindOneByIDSelection(ctx, &summary, id)
```

Use a custom repository in `internal/adapter/repository` when:

- The query needs raw SQL, unions, or advanced joins.
- The query result is not the same shape as the domain entity.
- The query is reused by several usecases.

The IAM `UserRepo.UserDashboardList` is an example of a custom repository for a union query.

### 🔁 Using DB Transaction

Database transaction helper lives in `pkg/db/dbTransaction.go`.

Use it when one usecase needs several database writes to succeed or fail as one unit. The helper works with repositories that implement:

```go
type BaseRepository interface {
    SetupConnection(db *gorm.DB)
}
```

The generic repository pointer already implements this method, so create transaction repositories with `NewGenericeRepoPointr`.

Recommended usecase shape:

```go
type Usecase struct {
    base.Port
    dbConn *gorm.DB
}

func NewUsecase(dbConn *gorm.DB, port base.Port) Usecase {
    return Usecase{
        Port:   port,
        dbConn: dbConn,
    }
}
```

Transaction example:

```go
func (u Usecase) CreateWithTags(ctx context.Context, request CreateArticleRequest) (err error) {
    articleRepo := db.NewGenericeRepoPointr(u.dbConn, domain.Article{})
    tagRepo := db.NewGenericeRepoPointr(u.dbConn, domain.ArticleTag{})

    trx := db.NewDBTransaction(u.dbConn, articleRepo, tagRepo)
    trx.Begin()
    defer func() {
        trxErr := trx.End(err)
        if err == nil && trxErr != nil {
            err = trxErr
        }
    }()

    article, err := articleRepo.Store(ctx, domain.Article{
        Title: request.Title,
        Body:  request.Body,
    })
    if err != nil {
        return u.ErrHandler.ErrorReturn(err)
    }

    _, err = tagRepo.BulkStore(ctx, buildArticleTags(article.ID, request.TagIDs))
    if err != nil {
        return u.ErrHandler.ErrorReturn(err)
    }

    return nil
}
```

How it works:

- `NewDBTransaction(dbConn, repos...)` receives the original `*gorm.DB` and every repository that should join the transaction.
- `Begin()` calls `db.Begin()` and runs `SetupConnection(begin)` on every registered repository.
- `End(err)` rolls back when `err != nil`, and commits when `err == nil`.

Rules:

- Use pointer repositories for transaction scope: `db.NewGenericeRepoPointr(...)`.
- Create transaction repositories inside the usecase method, then use them only for that transaction.
- Use a named return error, like `(err error)`, so the deferred `End(err)` sees the final error.
- Return immediately after a failed transaction step so `End(err)` rolls back.
- For custom repositories, add `SetupConnection(db *gorm.DB)` so they satisfy `db.BaseRepository`.

Custom repository transaction support:

```go
type articleRepo struct {
    db *gorm.DB
}

func (repo *articleRepo) SetupConnection(dbConn *gorm.DB) {
    repo.db = dbConn
}
```

### 🔒 Security Guide

Security is provided by `iam_module`. It handles JWT authentication, authorization, login session cache, registration, OTP verification, and user management.

The module supports single JWT token authentication. Login stores token data in request context and stores richer user session data in Redis. The cached session is updated when user data changes.

JWT user data follows this shape:

```json
{
  "userId": "<generated-auth-code>",
  "email": "user@example.com",
  "lang": "id",
  "timezone": "Asia/Jakarta",
  "roleName": "USER"
}
```

Security logic is placed in `iam_module`. This base project already provides basic usecases for registration flow and user management. User context is split between mobile user context and dashboard/admin context.

Toggle IAM security with:

```env
IAM_MODULE_OFF=true
```

When `IAM_MODULE_OFF=true`, `shared/base.NewAuth` returns `EmptyAuth`, and IAM controllers are not registered in `cmd/api/api.go`. If you decide to delete or replace `iam_module`, the main module can still run as long as the security port is replaced or disabled correctly.

#### Get session data

Use these methods from `base.Port.Security` inside usecases:

```go
type Security interface {
    GetUserContext(ctx context.Context) payload.UserData
    GetSessionLogin(ctx context.Context, sessionData *payload.SessionDataUser) error
    GetSession(ctx context.Context, authCode string, sessionData *payload.SessionDataUser) error
}
```

Example:

```go
var session payload.SessionDataUser
err := u.Security.GetSessionLogin(ctx, &session)
if err != nil {
    return err
}
```

#### Apply authentication and authorization

Use these methods in controller routes:

```go
type Security interface {
    Validate() gin.HandlerFunc
    Authorize(roles ...string) gin.HandlerFunc
}
```

Example:

```go
adminRouter.GET(
    "/users",
    ctrl.Security.Validate(),
    ctrl.Security.Authorize("ADMIN"),
    ctrl.GetListUser,
)
```

`Validate()` reads the `Authorization: Bearer <token>` header, validates the JWT with `SECRET`, attaches `payload.UserData` into the request context, and updates `last_active`.

`Authorize(...)` checks the role attached by `Validate()`.

#### Set or refresh session

Use `SetSession` when a usecase changes data that should be reflected in Redis session data:

```go
err := u.Security.SetSession(ctx, payload.SessionDataUser{
    UserReference: user.AuthCode,
    RoleName:      user.GetRoleName(),
    TimeZone:      login.Timezone,
    Lang:          login.Lang,
    Email:         user.Email,
    Name:          user.FullName,
    IsVerified:    user.GetIsVerified(),
})
```

Security-related environment keys:

- `IAM_MODULE_OFF`
- `SECRET`
- `EXPIRED_TOKEN_JWT`
- `SECRET_USER_ID`
- `ENCRYPT_MESSAGE_PASSWORD`
- `FALLBACK_TIMEZONE`
- `FALLBACK_LANG`
- `EMAIL_VERIFICATION_OFF`
- `HOTP_SECRET`
- `EXPARATION_OTP_TIME`
- `FRONT_END_HOST`
- `SMPT_SERVER_HOST`
- `SMPT_SERVER_PORT`
- `SUPPORT_EMAIL`
- `SUPPORT_EMAIL_PASS`

### 🧩 Project Structure

This project follows a domain-driven, layered structure. The intended dependency direction is:

```txt
controller -> usecase -> repository -> domain
```

Project layout:

```txt
cmd/api
  Application entrypoint. Loads env, registers controllers, starts Gin.

internal/core/domain
  Main application entities and domain behavior.

internal/core/usecase
  Main application business rules and orchestration.

internal/adapter/controller
  Main application HTTP controllers.

internal/adapter/repository
  Main application database queries that are too specific for generic repository.

pkg
  Infrastructure packages and reusable adapters: db, cache, middleware, mapper,
  logger, localize, mailing, MinIO storage, environment, clock, and utilities.

shared/api
  Gin API bootstrap, router interface, controller registration, and server start.

shared/base
  Ports and base controller dependencies injected into controllers and usecases.

shared/payload
  Shared request, response, pagination, and session payloads.

resource/message
  Localization message files.

iam_module
  Separate workspace module for authentication, authorization, registration,
  user management, IAM middleware, IAM domain, and IAM repositories.

crudgen_module
  Separate workspace module reserved for CRUD generator tooling.
```

Base project rules:

- `cmd/api` is the composition root. Register controllers there, but keep business logic out of it.
- Controllers should stay thin: bind request, validate request, call usecase, return response.
- Usecases should contain business rules and use `base.Port` for shared services.
- Repositories should contain persistence details. Prefer generic repository for simple CRUD and custom repositories for complex SQL.
- Domain structs should define table mapping, entity behavior, and small state helpers. Do not import Gin in domain or usecase packages.
- Shared response formatting should go through `ctrl.Mapper.NewResponse`.
- Shared validation should go through `ctrl.Enigma`.
- Security middleware should be applied in controller `Route` methods.
- New cross-cutting dependencies should be added to `shared/base.Port` only when they are broadly useful.
- Keep each module's `internal` package private to that module boundary.

## 🚢 Deployment

### 🏗️ Build Binary

Build for the current operating system:

```bash
go build -o bin/api ./cmd/api
```

Build a Linux binary from another platform:

```bash
GOOS=linux GOARCH=amd64 go build -o bin/base-be-api ./cmd/api
```

Run the binary:

```bash
./bin/base-be-api -env .env.stag
```

For production, place the binary, env file, `resource/message`, and any required resource templates on the target machine. The app reads resources by relative path, so run the binary from the project deployment directory or keep the same resource layout beside the binary.

### ⚙️ Run with systemd

Example service file:

```ini
[Unit]
Description=Base Backend Golang API
After=network.target mysql.service redis.service

[Service]
Type=simple
WorkingDirectory=/opt/base-be-golang
ExecStart=/opt/base-be-golang/bin/base-be-api -env /opt/base-be-golang/.env.stag
Restart=always
RestartSec=5
User=www-data
Group=www-data

[Install]
WantedBy=multi-user.target
```

Install and start:

```bash
sudo cp base-be-golang.service /etc/systemd/system/base-be-golang.service
sudo systemctl daemon-reload
sudo systemctl enable base-be-golang
sudo systemctl start base-be-golang
sudo systemctl status base-be-golang
```

Read logs:

```bash
journalctl -u base-be-golang -f
```

## 📖 Additional Information

- API routes are mounted under `/api/v1`.
- Localization files are loaded from `resource/message/*.json`.
- `IAM_MODULE_OFF=true` disables IAM route registration and uses empty auth middleware.
- `DB_LOG_MODE` follows GORM log levels: `1` silent, `2` error, `3` warn, `4` info.
- `go.work` includes the root module, `iam_module`, and `crudgen_module`; run `go work sync` after changing workspace dependencies.
- `crudgen_module` currently contains early generator scaffolding and usage documentation placeholder.
- Diagrams are stored in `resource/diagram` and `iam_module/resource/diagram`.
