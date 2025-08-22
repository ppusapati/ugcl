# Database Migrations with Atlas

This directory contains database migrations managed by [Atlas](https://atlasgo.io/).

## Prerequisites

1. Install the Atlas CLI:
   ```bash
   # macOS and Linux
   curl -sSf https://atlasgo.sh | sh
   
   # Windows with scoop
   scoop install atlas
   
   # or with chocolatey
   choco install atlas
   ```

## Getting Started

1. Set your database connection string in your environment:
   ```bash
   # Example for PostgreSQL
   export DB_DSN="postgres://user:pass@localhost:5432/yourdb?sslmode=disable"
   ```

2. Create a new migration:
   ```bash
   # Generate a new migration
   atlas migrate diff <migration_name> \
     --dir "file://migrations" \
     --to "postgres://user:pass@localhost:5432/yourdb?sslmode=disable" \
     --dev-url "docker://postgres/15"
   ```

3. Apply migrations:
   ```bash
   atlas migrate apply
   ```

4. Check migration status:
   ```bash
   atlas migrate status
   ```

## Development Workflow

1. Make changes to your database schema in your code
2. Generate a migration:
   ```bash
   atlas migrate diff <descriptive_name>
   ```
3. Review the generated migration file in the `migrations` directory
4. Apply the migration:
   ```bash
   atlas migrate apply
   ```

## CI/CD Integration

For CI/CD, you can use the following commands:

```bash
# Check for pending migrations
atlas migrate validate --dir "file://migrations"

# Apply migrations in non-interactive mode
atlas migrate apply --dir "file://migrations" --url $DB_DSN --auto-approve
```

## Rollback

To rollback the last migration:

```bash
atlas migrate down 1
```

## Documentation

- [Atlas CLI Reference](https://atlasgo.io/cli-reference)
- [Migration Directory Format](https://atlasgo.io/concepts/migration-directory)
- [Versioned Migrations](https://atlasgo.io/concepts/versioned-migrations)
