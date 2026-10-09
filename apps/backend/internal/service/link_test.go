//nolint:testpackage // Exercise the existing private destination policy before extraction.
package service

import (
	"testing"

	"github.com/6sLOGAN78/flux/internal/config"
)

func TestDestinationReservedRanges(t *testing.T) {
	policy := config.LinksConfig{ManagedHost: "go.flux.test", BlockedHosts: []string{"blocked.example"}}
	for _, value := range []string{
		"http://192.88.99.1", "http://[fec0::1]", "http://[2001:20::1]", "http://[3fff::1]",
		"http://[4000::1]", "http://[::ffff:192.88.99.1]",
		"https://nested.localhost", "https://host.local", "https://host.internal",
		"https://sub.blocked.example", "https://go.flux.test./loop",
	} {
		t.Run(value, func(t *testing.T) {
			if _, err := validateDestination(value, policy); err == nil {
				t.Fatal("unsafe destination accepted")
			}
		})
	}
}
