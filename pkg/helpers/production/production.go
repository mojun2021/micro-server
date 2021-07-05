// Package production gives useful helpers about the current environment.
package production

import (
	"fmt"
	"net"
	"os"
)

// A global flag that indicates whether the code is running in a production
// environment.
//
// Note that in this context, "production" does not necessarily mean "player
// facing". It mainly means the code is not running on a developer machine.
var inProduction bool

// InProduction checks whether the code currently runs in production.
func InProduction() bool { return inProduction }

// SetInProduction sets the production state. This should be used only for
// testing. You have been warned.
func SetInProduction(v bool) func() {
	previous := inProduction
	inProduction = v

	return func() { inProduction = previous }
}

// Hostname get the hostname of the current host, or "localhost" in case of
// error.
func Hostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}

	return "localhost"
}

// EndpointToHostname converts an endpoint to a compatible hostname
// representation suitable for being used in a HTTP request URI.
//
// If useLocal is set, resolution will favor local addresses.
func EndpointToHostname(endpoint string, useLocal bool) string {
	defaultHost := "localhost"

	if !useLocal {
		defaultHost = Hostname()
	}

	if host, port, err := net.SplitHostPort(endpoint); err == nil {
		switch host {
		case "::", "0.0.0.0", "":
			host = defaultHost
		}

		return fmt.Sprintf("%s:%s", host, port)
	}

	if ip := net.ParseIP(endpoint); ip != nil {
		if !ip.IsUnspecified() {
			return ip.String()
		}
	}

	return defaultHost
}

func init() {
	// TODO: Reevaluate the logic here. Running in Kubernetes does not
	// necessarily mean "in production".
	const envVarKubernetes = "KUBERNETES_PORT"

	SetInProduction(os.Getenv(envVarKubernetes) != "")
}
