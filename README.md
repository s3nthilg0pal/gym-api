# Gym API

A Go API for managing gym entries using GORM ORM.

**CORS:** Enabled with default settings allowing all origins, methods, and headers.

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

**Response (new entry):**
```json
{
  "message": "entry added"
}
```

**Response (entry already exists):**
```json
{
  "message": "entry already exists"
}
```

### GET /health
Health check endpoint that verifies database connectivity.

**Response (healthy):**
```json
{
  "status": "healthy"
}
```

**Response (unhealthy):**
```json
{
  "status": "unhealthy",
  "error": "database ping failed"
}
```

### GET /visits/progress/message
Returns a motivational progress message based on visits compared to goal.

**Response:**
```json
{
  "message": "💪 In the zone! 42 of 100 days - keep the momentum!"
}
```

**Message tiers:**
- 🏆 100%+: Champion! You crushed it
- 🔥 80-99%: Almost there! Finish strong
- 💪 50-79%: In the zone! Keep the momentum
- 🚀 20-49%: Building habits! You're on your way
- 🌱 0-19%: Every rep counts! Let's go

## Environment Variables

- `DATABASE_URL`: PostgreSQL connection string (libpq format)
  Example: `host=your-host port=5432 user=your-user password=your-password dbname=your-db sslmode=disable`
- `API_KEY`: API key for POST endpoint (default: "default-secret")
- `PORT`: Port to run on (default: 8080)

## Database

The API uses GORM ORM with PostgreSQL. The following tables are auto-migrated on startup:
- `entry`: id (primary key), date (timestamp), visited (boolean)
- `goal`: id (primary key), value (integer) - stores the visit goal target

## Running

1. Set environment variables
2. Run `go run .`

## Docker

Build and push the image to GitHub Container Registry:

```bash
# Build the image
docker build -t gym-api .

# Tag for GHCR
docker tag gym-api ghcr.io/s3nthilg0pal/gym-api:latest

# Login to GHCR (requires GitHub Personal Access Token with package permissions)
echo $GITHUB_TOKEN | docker login ghcr.io -u s3nthilg0pal --password-stdin

# Push the image
docker push ghcr.io/s3nthilg0pal/gym-api:latest
```

Then run locally with:

```bash
docker run -e DATABASE_URL="host=your-host port=5432 user=your-user password=your-password dbname=your-db sslmode=disable" -e API_KEY="your-api-key" -p 8080:8080 gym-api
```

## Kubernetes

Deploy the Docker image to your K8s cluster with appropriate environment variables and database connection.

Use the Kustomize files in the `k8s/` directory for deployment with ArgoCD:

```bash
kubectl apply -k k8s/
```

**Note:** Update the `secret.yaml` with your actual base64-encoded database URL and API key before deploying.

Or use ArgoCD to deploy from this repository.