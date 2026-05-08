# chirpy

Chirpy is an API based on Twitter that allows users to post and retrieve tweets. This is a project built to help me learn Go backend development. The focus areas here were REST APIs, JWT Authentication, and PostgreSQL.

## Tech Stack

- Go
- PostgreSQL
- sqlc
- Goose
- JWT authentication

## Features

- User authentication with JWTs
- Refresh token support
- Chirp CRUD endpoints
- PostgreSQL persistence
- SQL generated with sqlc
- Goose migrations

## Getting Started

### Install Golang if needed:

```
sudo apt-get update
sudo apt-get -y install golang-go
```

### Install Goose:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Clone Chirpy repo locally:

```
git clone https://github.com/mikelawson03/chirpy <destination_path>
```

### Create database:

1. Connect to PostgreSQL:

- Mac: `psql postgres`
- Linux/WSL: `sudo -u postgres psql`

2. Create new database:

```
CREATE DATABASE chirpy;
```

3. Connect to database:

```
\c chirpy
```

4. Set user password:

```
ALTER USER {username} WITH PASSWORD {password}
```

### Set up local environment:
1. Create a .env file in the same directory where you installed the project

2. Using the username and password you set when creating the database, add the database string to this file:

```
DB_URL="postgres://{username}:{password}@localhost:5432/chirpy?sslmode=disable
```

3. Create a secret string for authentication. You can either enter your own or create a random string in the command line with `openssl rand -base64 64`. Add this string to your .env file (make sure to wrap the string in quotes):
```
SECRET="{secret_string}"
```
4. To use the reset endpoint, you'll need to be using a dev environment. If you'd like to enable this, add the following to your env file as well: 
```PLATFORM ="dev"```


### Run database Migrations:
From the sql/schema directory, run the goose migrations (note that the database URL will be the same one you used in your .env file):
```
goose "{database_url}" up
```

### Build Go package and run
```
go build && ./chirpy
```

## API Endpoints:

### Authorization
Provide access token in the Authorization header:

```http
Authorization: Bearer {token}
```

### GET /api/healthz
Authentication: None
Return OK if server is ready for requests

Response: `200 OK`

### POST /api/users
_Authentication: None_

Create a new user.

Request:
```json
{
    "email": "{email_address}",
    "password": "{password}"
}
```
Response: `201 Created`
```json
{
    "id":"{user id}",
    "created_at":"{created timestamp}",
    "updated_at":"{last updated timestamp}",
    "email":"{email address}"
}
```

### PUT /api/users
_Authentication: Access Token_
Update user information. 

Request: `200 OK`
```json
{
    "email": "{email_address}",
    "password": "{password}"
}
```
Response:
```json
{
    "id":"{user id}",
    "created_at":"{created timestamp}",
    "updated_at":"{last updated timestamp}",
    "email":"{email address}"
}
```

### POST /api/login
_Authentication: None_
Login with specified user.

Request:
```json
{
    "email": "{email_address}",
    "password": "{password}"
}
```

Response: `200 OK`
```json
{
    "id":"{user_id}",
    "created_at":"{created timestamp}",
    "updated_at":"{last updated timestamp>",
    "email":"{email address>",
    "token":"{access token>",
    "refresh_token":"{refresh_token}"
}
```

### POST /api/refresh
_Authentication: Refresh Token_
Will return new Access token given header with valid, unexpired Refresh Token.

Response: `200 OK`
```json
{
    "token": "{access token}"
}

```

### POST /api/revoke
_Authentication: Access Token_
Revoke refresh token. Will revoke specified refresh token. Access token will remain active for up to 1 hour.

Response: `204` 

### GET /api/chirps
_Authentication: None_
Returns all chirps from all users. 

Response: `200 OK`
```json
[
    {"id":"{chirp id}",
    "created_at":"{created timestamp}",
    "updated_at":"{updated timestamp}",
    "body":"{chirp body}",
    "user_id":"{user id}"},

    {"id":"{chirp id}",
    "created_at":"{created timestamp}",
    "updated_at":"{updated timestamp}",
    "body":"{chirp body}",
    "user_id":"{user id}"}
    
    ...
]
```

### GET /api/chirps/{chirp_id}
_Authentication: None_
Returns chirp with specified ID

Response: `200 OK`
```json
{
    "id":"{chirp id}",
    "created_at":"{created timestamp}",
    "updated_at":"{updated timestamp}",
    "body":"{chirp body}",
    "user_id":"{user id}"
}
```

### GET /api/chirps?author_id={user_id}&sort=asc
_Authentication: None_
Get chirps from specified user.

Optional query parameters:
- `sort=asc`
- `sort=desc`

Return: `200 OK`
```json
[
    {"id":"{chirp id}",
    "created_at":"{created timestamp}",
    "updated_at":"{updated timestamp}",
    "body":"{chirp body}",
    "user_id":"{user id}"},

    {"id":"{chirp id}",
    "created_at":"{created timestamp}",
    "updated_at":"{updated timestamp}",
    "body":"{chirp body}",
    "user_id":"{user id}"}
    
    ...
]
```
### POST /api/chirps
_Authentication: Access Token_
Post new chirp. 

Request: `201 Created`
```json
{
  "body": "{chirp body}"
}
```
Response:
```json
{
    "id":"{chirp id}",
    "created_at":"{created timestamp}",
    "updated_at":"{updated timestamp}",
    "body":"{chirp body}",
    "user_id":"{user id}"
}
```

### DELETE /api/chirps/{chirpID}
_Authentication: Access Token_
Delete specified chirp. 

Response: `204 No Content`

### POST /admin/reset/
_Authentication: None_
Reset database. Removes all users, chirps, and refresh tokens from their respective tables. Requires dev environment.

Response: `200 OK`