# Base Backend Golang
[![GitHub stars](https://img.shields.io/github/stars/rdhmuhammad/base-be-golang)](https://github.com/rdhmuhammad/base-be-golang/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/rdhmuhammad/base-be-golang?)](https://github.com/rdhmuhammad/base-be-golang/network/members)

***
> This repository is skeleton for Backend project, we seek to apply technology-agnostic, 
> but for several major layer bound to specific third-party such as gorm, gin-gonic, etc.

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

- 🛠️ [Development Guide](#️-development-guide)
    - ➕ [Creating New Endpoint](#-creating-new-endpoint)
    - 🔒 [Security Guide](#-security-guide)
    - 🧩 [Project Structure](#-project-structure)

- 🚢 [Deployment](#-deployment)
    - 🏗️ [Build Binary](#️-build-binary)
    - ⚙️ [Run with systemd](#️-run-with-systemd)

- 📖 [Additional Information](#-additional-information)

</details>

<a name="-security-guide"></a>
### 🔒 Security Guide
Security provide authentication with only jwt token, where it either singles token or refresh token. 
for each login, token data is stored to context while user entity stored to redish cache, cache also updated along with every user data augmentation. token data as follow: 
```json
{
  "timeZone": "Asia/Jakarta",
  "Lang": "id",
  "id": "<generated-code>"
}
```
Logic for security is placed at `iam_module`. We already provide basic usecase for `user management` and `registration flow`. 
User context is divide into mobile and dashboard.
you can toggle the security to on or off with environment key `IAM_MODULE_OFF`.
_If you decide to delete `iam_module`, its won't affect main module._

here, several security utility that may you need on usecase.

**Get session**
```go
type port interface{
    GetSessionAdmin(ctx context.Context, authCode string) (domain.UserAdmin, error)
    GetSessionMobile(ctx context.Context, authCode string) (domain.User, error)	
}
```

**Middleware to Apply Authentication and Validation**
```go
type port interface{
    Validate() gin.HandlerFunc
    Authorize(roles ...string) gin.HandlerFunc
}
```