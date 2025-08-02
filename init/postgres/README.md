# PostgreSQL Initialization Script

This directory contains the PostgreSQL initialization script that runs when the PostgreSQL container starts for the first time.

## File: `01-init.sql`

This script automatically sets up the database schema for the Go-Next backend application.

### What it does:

1. **Enables UUID extension** - For generating UUID primary keys
2. **Creates database tables**:
   - `users` - User accounts and profiles
   - `roles` - User roles and permissions
   - `user_roles` - Many-to-many relationship between users and roles
   - `categories` - Content categories (hierarchical)
   - `posts` - Blog posts and articles
   - `comments` - User comments on posts
   - `media` - File uploads and media
   - `mediable` - Polymorphic associations for media
   - `notifications` - User notifications
   - `refresh_tokens` - JWT refresh tokens
   - `verification_tokens` - Email verification and password reset tokens

3. **Inserts default data**:
   - Default roles: admin, moderator, author, user
   - Default admin user: admin@example.com (password: admin123)
   - Default categories: Technology, Business, Lifestyle, News, Tutorials

4. **Creates indexes** for better query performance
5. **Sets up triggers** to automatically update `updated_at` timestamps
6. **Grants necessary permissions** to the postgres user

### Default Admin Credentials

- **Email**: admin@example.com
- **Password**: admin123
- **Role**: admin (full access)

⚠️ **Important**: Change the default admin password in production!

### Database Schema Features

- **UUID Primary Keys**: All tables use UUID primary keys for better security
- **Timestamps**: Automatic `created_at` and `updated_at` timestamps
- **Soft Deletes**: Support for soft delete patterns
- **Polymorphic Associations**: Media can be attached to any model
- **Role-Based Access Control**: Flexible permission system
- **Hierarchical Categories**: Categories can have parent-child relationships
- **Comment Threading**: Comments support nested replies

### Running the Script

The script runs automatically when the PostgreSQL container starts for the first time. It's mounted in the Docker Compose configuration:

```yaml
volumes:
  - ./init/postgres:/docker-entrypoint-initdb.d
```

### Manual Execution

If you need to run the script manually:

```bash
# Connect to PostgreSQL container
docker exec -it go-next-postgres psql -U postgres -d go_next

# Or run the script directly
docker exec -i go-next-postgres psql -U postgres -d go_next < init/postgres/01-init.sql
```

### Notes

- The script uses `IF NOT EXISTS` clauses to prevent errors on re-runs
- All inserts use `ON CONFLICT` clauses to prevent duplicate data
- The script is idempotent and can be run multiple times safely
- Indexes are created for commonly queried columns
- Triggers automatically update `updated_at` timestamps 