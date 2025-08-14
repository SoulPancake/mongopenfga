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
1. Creates an OpenFGA store
2. Writes an authorization model for a simple document sharing system
3. Writes relationship tuples (user permissions)
4. Performs authorization checks
5. Validates the setup works correctly

## Architecture

- **MongoDB**: Document database storing OpenFGA data (tuples, models, stores)
- **OpenFGA Server**: Authorization service with MongoDB storage backend
- **Client**: Go program demonstrating OpenFGA operations

## Files

- `docker-compose.yaml` - MongoDB and OpenFGA services
- `Makefile` - Commands to run the example
- `client/main.go` - Example client application
- `client/go.mod` - Go module for the client