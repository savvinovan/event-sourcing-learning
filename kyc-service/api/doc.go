// Package api contains the OpenAPI 3.0 specification and generator config for kyc-service.
//
// To regenerate the server interface and models:
//
//	go generate ./api/...
package api

//go:generate go tool oapi-codegen -config oapi-codegen.yaml openapi.yaml
