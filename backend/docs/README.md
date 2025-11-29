# Backend Documentation

This directory contains comprehensive documentation for the Order System API.

## Documentation Files

- `API_DOCUMENTATION.md` - Complete API documentation with endpoints, models, and usage
- `swagger.json` - Auto-generated Swagger JSON specification
- `swagger.yaml` - Auto-generated Swagger YAML specification
- `docs.go` - Auto-generated Go documentation for Swagger

## Swagger Documentation

The API includes interactive Swagger documentation accessible at `/swagger/index.html` when the server is running. This provides a user-friendly interface to explore and test all API endpoints.

## Auto-Generated Documentation

The Swagger documentation is automatically generated using the `swag` tool. When API endpoints or models change, run:

```bash
cd backend
swag init --parseDependency --parseInternal --parseDepth 1
```

This regenerates the Swagger files with updated API information.

## Documentation Updates

When adding new endpoints or changing existing ones, remember to update the Swagger annotations in the handler files to keep the documentation accurate and comprehensive.