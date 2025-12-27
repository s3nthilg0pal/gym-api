# Gym API

A Go API for managing gym entries using GORM ORM.

## Endpoints

### GET /entry
Retrieves all entries from the Entry table.

**Response:**
```json
[
  {
    "date": "2023-12-27T00:00:00Z",
    "visited": true
  }
]
```

### POST /entry
Adds a new entry with the given date and visited=true.

**Headers:**
- `X-API-Key`: API key for authentication

**Payload:**
```json
{
  "date": "2023-12-27"
}
```

**Response:**
```json
{
  "message": "entry added"
}
```

## Environment Variables

- `DATABASE_URL`: PostgreSQL connection string (libpq format)
  Example: `host=your-host port=5432 user=your-user password=your-password dbname=your-db sslmode=disable`
- `API_KEY`: API key for POST endpoint (default: "default-secret")
- `PORT`: Port to run on (default: 8080)

## Database

The API uses GORM ORM with PostgreSQL. The `entry` table is auto-migrated on startup with columns:
- `id` (primary key, auto-increment)
- `date` (timestamp)
- `visited` (boolean)

## Running

1. Set environment variables
2. Run `go run .`

## Docker

Build and run with Docker:

```bash
docker build -t gym-api .
docker run -e DATABASE_URL="host=your-host port=5432 user=your-user password=your-password dbname=your-db sslmode=disable" -e API_KEY="your-api-key" -p 8080:8080 gym-api
```

## Kubernetes

Deploy the Docker image to your K8s cluster with appropriate environment variables and database connection.