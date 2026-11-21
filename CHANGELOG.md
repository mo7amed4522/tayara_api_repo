# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.2.0] - 2026-11-21

### Added
- User profile management endpoints (get, update, delete user account)
- Cash tracking in database for API calls
- Explore target data improvements
- Explore test server-side implementation

### Changed
- Refactored database-backed lists to use `utils.PaginateWithCustomQuery`
- Normal business-service reads no longer use handwritten Raw queries
- Updated pagination parameters for discover destinations endpoint
- Updated pagination parameters for home ads endpoint

### Fixed
- Security data issues resolved
- Explore target data corrections

### Documentation
- Regenerated swagger docs with swag v1.8.12
- Synchronized swagger docs format across all endpoints
- Added comprehensive pagination parameter documentation

## [v0.1.0] - 2026-08-22

### Added
- Initial release
- Core API functionality
- Basic authentication and authorization
- Database models and migrations
- Docker support
- API documentation with Swagger
