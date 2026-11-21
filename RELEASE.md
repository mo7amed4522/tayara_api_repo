# Release Notes

## Version 0.2.0 (2026-11-21)

### Overview
This release includes significant improvements to user management, database query optimization, and API documentation. The focus has been on refactoring database operations for better performance and adding comprehensive user profile features.

### 🎉 New Features

#### User Management
- **User Profile API**: Complete user profile management system
  - Get user profile details
  - Update user information
  - Delete user account functionality
- **Cash Tracking**: Added cash tracking in database for API call monitoring

#### Explore Features
- **Server-side Explore Tests**: Migrated explore test functionality to server-side implementation
- **Enhanced Target Data**: Improved explore target data handling and validation

### 🔧 Improvements

#### Database Optimization
- **Pagination Refactor**: Refactored all database-backed lists to use `utils.PaginateWithCustomQuery`
- **Query Optimization**: Eliminated handwritten Raw queries in business-service reads
- **Cleaner Code**: More maintainable and consistent database access patterns

#### API Documentation
- **Swagger Updates**: Regenerated all swagger documentation with swag v1.8.12
- **Pagination Documentation**: Added comprehensive pagination parameters to:
  - Discover destinations endpoint
  - Home ads endpoint
  - All list endpoints

### 🐛 Bug Fixes
- Resolved security data issues
- Fixed explore target data inconsistencies

### 📦 Installation

```bash
# Clone the repository
git clone https://github.com/your-repo/tayara_api_repo.git
cd tayara_api_repo

# Checkout v0.2.0
git checkout v0.2.0

# Install dependencies
go mod download

# Run the application
go run main.go
```

### 🔄 Upgrade Guide

If you're upgrading from v0.1.0:

1. **Database Changes**: Ensure your database schema is up to date
2. **Configuration**: Review `config.yaml` for any new settings
3. **API Calls**: Check the updated Swagger docs for pagination parameter changes

### 📚 Documentation
- Full API documentation available at `/swagger/index.html`
- See [CHANGELOG.md](CHANGELOG.md) for detailed change list

### 🙏 Contributors
Special thanks to all contributors who made this release possible!

---

## Version 0.1.0 (2026-08-22)

### Initial Release
- Core API functionality
- Authentication and authorization
- Database models and migrations
- Docker support
- Swagger API documentation
