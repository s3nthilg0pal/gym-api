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