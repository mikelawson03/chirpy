# chirpy

Chirpy is an API based on Twitter than allows users to post and retrieve tweets

## Tech Stack

- Go
- PostgreSQL
- sqlc
- Goose
- JWT authentication

## Getting Started

### Install Golang if needed:

```
sudo apt-get update
sudo apt-get -y install golang-go
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
ALTER USER <username> WITH PASSWORD <password>
```

### Set up local environment:
1. Create a .env file in the same directory where you installed the project

2. Using the username and password you set when creating the database, add the database string to this file:

```
DB_URL="postgres://<username>:<password>@localhost:5432/chirpy?ssl=disable
```

3. Create a secret string for authentication. You can either enter your own or create a random string in the command line with `openssl rand -base64 64`. Add this string to your .env file (make sure to wrap the string in quotes):
```
SECRET="<secret_string>"
```
4. To use the reset endpoint, you'll need to be using a dev environment. If you'd like to enable this, add the following to your env file as well: 
```PLATFORM ="dev"```


### Run database Migrations:
From the sql/schema directory, run the goose migrations (note that the database URL will be the same one you used in your .env file):
```
goose "<database_url>" up
```

### Build Go package and run
```
go build && ./chirpy
```

## API Endpoints:

### Authorization
Provide access token in the Authorization header:

```http
Authorization: Bearer <token>
```

### GET /api/chirps
Return all chirps

### GET /api/chirps?auth_id={user_id}
Get all chirps from user

### GET /api/healthz
Return OK if server is ready for requests

### POST /api/users
Create a new user.

Request:
```json
{
    "email": "<email_address>",
    "password": "<password>"
}
```
Response:
```json
{
    "id":"<user id>",
    "created_at":"<created timestamp>",
    "updated_at":"<last updated timestamp>",
    "email":"<email address>"
}
```

### PUT /api/users
Update user information. Requires authentication via Bearer token.

Request:
```json
{
    "email": "<email_address>",
    "password": "<password>"
}
```
Response:
```json
{
    "id":"<user id>",
    "created_at":"<created timestamp>",
    "updated_at":"<last updated timestamp>",
    "email":"<email address>"
}

### POST /api/login

Login with specified user.

Request:
```json
{
    "email": "<email_address>",
    "password": "<password>"
}
```

Response:
```json
{
    "id":"{user_id}",
    "created_at":"<created timestamp>",
    "updated_at":"<last updated timestamp>",
    "email":"<email address>",
    "token":"<access token>",
    "refresh_token":"<refresh_token>"}

```

### POST /api/refresh
Will return new Access token given header with valid, unexpired Refresh Token. Requires authentication via Bearer token.

Response:

```json
{
    "token": "<access token>"
}

```

### POST /api/revoke
Revoke refresh token. Will revoke specified refresh token. Access token will remain active for up to 1 hour. Requires authentication via Bearer token.

Will return `204` on success with no body.

### GET api/chirps/
Returns all chirps from all users. Will return JSON on successful refresh:
```json
[
    {"id":"<chirp id>",
    "created_at":"<created timestamp>",
    "updated_at":"<updated timestamp>",
    "body":"<chirp body>",
    "user_id":"<user id>"},

    {"id":"<chirp id>",
    "created_at":"<created timestamp>",
    "updated_at":"<updated timestamp>",
    "body":"<chirp body>",
    "user_id":"<user id>"}
    
    ...
]
```

### GET api/chirps/{chirp_id}
Returns chirp with specified ID

Return:
```json
{
    "id":"<chirp id>",
    "created_at":"<created timestamp>",
    "updated_at":"<updated timestamp>",
    "body":"<chirp body>",
    "user_id":"<user id>"
}
```

### GET /api/chirps?author_id={user_id}&sort=asc
Get chirps from specified user.

Optional query parameters:
- `sort=asc`
- `sort=desc`

Return:
```
json
[
    {"id":"<chirp id>",
    "created_at":"<created timestamp>",
    "updated_at":"<updated timestamp>",
    "body":"<chirp body>",
    "user_id":"<user id>"},

    {"id":"<chirp id>",
    "created_at":"<created timestamp>",
    "updated_at":"<updated timestamp>",
    "body":"<chirp body>",
    "user_id":"<user id>"}
    
    ...
]
```
### POST api/chirps
Post new chirp. Authentication required via Bearer token.

Request:
```json
{
  "body": "Everyone is telling me I'll never win if I fall in love with a girl from Marin",
  "user_id": "31a71526-8417-4f47-ac1a-98540448b1a2"
}
```
Response:
```json
{
    "id":"<chirp id>",
    "created_at":"<created timestamp>",
    "updated_at":"<updated timestamp>",
    "body":"<chirp body>",
    "user_id":"<user id>"
}
```

### PUT api/user
