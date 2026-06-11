package cache

import "fmt"

// Key builds a namespaced cache key: service:entity:identifier.
// Example: Key("wallet", "user", "123") => "wallet:user:123"
func Key(service, entity, identifier string) string {
	return fmt.Sprintf("%s:%s:%s", service, entity, identifier)
}

// Prefix returns a scan/delete prefix for a service and entity.
func Prefix(service, entity string) string {
	return fmt.Sprintf("%s:%s:", service, entity)
}

// ServicePrefix returns a scan/delete prefix for an entire service namespace.
func ServicePrefix(service string) string {
	return service + ":"
}
