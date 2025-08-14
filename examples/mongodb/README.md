# OpenFGA with MongoDB Example

This example demonstrates how to use OpenFGA with MongoDB as the storage backend.

## Quick Start

1. Start MongoDB and OpenFGA:
   ```bash
   make up
   ```

2. Run the example client:
   ```bash
   make run
   ```

3. Clean up:
   ```bash
   make down
   ```

## What This Example Does

The example client program:
1. ✅ Creates an OpenFGA store 
2. ✅ Writes an authorization model for a simple document sharing system
3. ⚠️ Attempts to write relationship tuples (API formatting issue)
4. ✅ Demonstrates MongoDB storage integration
5. ✅ Validates that data is persisted in MongoDB

## Architecture

- **MongoDB**: Document database storing OpenFGA data (stores, authorization models, tuples, changelog)
- **OpenFGA Server**: Authorization service with MongoDB storage backend (built from source)
- **Client**: Go program demonstrating OpenFGA operations via HTTP API

## Verification

You can verify that data is being stored in MongoDB:

```bash
# Check stores
docker exec mongo mongosh mongodb://localhost:27017/openfga --eval "db.stores.find().count()"

# Check authorization models  
docker exec mongo mongosh mongodb://localhost:27017/openfga --eval "db.authorization_models.find().count()"

# List all collections
docker exec mongo mongosh mongodb://localhost:27017/openfga --eval "db.getCollectionNames()"
```

## Known Issues

- The tuple write API has a formatting issue that needs to be resolved
- This doesn't affect the core MongoDB storage functionality
- The playground can be used for interactive testing: http://localhost:3000/playground

## Files

- `docker-compose.yaml` - MongoDB service configuration
- `Makefile` - Commands to run the example
- `client/main.go` - Example client application
- `client/go.mod` - Go module for the client