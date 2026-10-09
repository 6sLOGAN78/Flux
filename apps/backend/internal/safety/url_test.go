package safety_test

import (
	"testing"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/safety"
)

func TestDestinationAdversarialCorpus(t *testing.T) {
	t.Parallel()
	policy := config.LinksConfig{ManagedHost: "go.flux.test", BlockedHosts: []string{"blocked.example"}}
	for _, value := range []string{
		"", "example.com", "//example.com", "https:example.com", "ftp://example.com", "javascript:alert(1)",
		"https://user:password@example.com", "https://example.com\\evil", "https://example.com/%5cpath",
		"https://example.com/%00", "https://example.com/%0d%0a", "https://example.com/\n", "https://example.com/ space",
		"https://%65xample.com", "https://example.com:", "https://example.com:0", "https://example.com:65536",
		"https://example.com:abc", "https://[example.com]", "https://[::1%25eth0]", "https://example..com",
		"https://localhost", "https://child.localhost.", "https://host.local", "https://host.internal",
		"https://metadata.google.internal", "https://metadata.google", "https://instance-data.ec2.internal",
		"https://go.flux.test/path", "https://child.go.flux.test", "https://sub.blocked.example",
		"http://0.0.0.0", "http://0.1.2.3", "http://10.255.255.255", "http://100.64.0.0", "http://100.127.255.255",
		"http://127.255.255.255", "http://169.254.169.254", "http://172.16.0.0", "http://172.31.255.255",
		"http://192.168.255.255", "http://192.0.0.1", "http://192.0.2.1", "http://192.88.99.1",
		"http://198.18.0.0", "http://198.19.255.255", "http://198.51.100.1", "http://203.0.113.1",
		"http://192.31.196.1", "http://192.52.193.1", "http://192.175.48.1", "http://[2620:4f:8000::1]",
		"http://224.0.0.1", "http://239.255.255.255", "http://240.0.0.1", "http://255.255.255.255",
		"http://[::]", "http://[::1]", "http://[fc00::1]", "http://[fdff::1]", "http://[fe80::1]",
		"http://[febf::1]", "http://[fec0::1]", "http://[ff02::1]", "http://[2001:db8::1]",
		"http://[2001::1]", "http://[2001:20::1]", "http://[2002::1]", "http://[3fff::1]", "http://[4000::1]",
		"http://[64:ff9b::7f00:1]", "http://[64:ff9b:1::1]", "http://[::ffff:127.0.0.1]",
		"http://[::ffff:192.88.99.1]", "http://2130706433", "http://127.1", "http://0177.0.0.1", "http://0x7f000001",
	} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			if _, err := safety.ValidateDestination(value, policy); err == nil {
				t.Fatal("unsafe destination accepted")
			}
		})
	}
}

func TestDestinationCanonicalMeaning(t *testing.T) {
	t.Parallel()
	policy := config.LinksConfig{ManagedHost: "go.flux.test"}
	for _, item := range []struct{ input, expected string }{
		{"https://EXAMPLE.COM./a%2Fb?q=a%2Bb&x=1#section", "https://example.com/a%2Fb?q=a%2Bb&x=1#section"},
		{"https://BÜCHER.example/path?x=%20", "https://xn--bcher-kva.example/path?x=%20"},
		{"http://8.8.8.8:8080/a", "http://8.8.8.8:8080/a"},
		{"https://[2606:4700:4700::1111]/", "https://[2606:4700:4700::1111]/"},
		{"https://[::ffff:8.8.8.8]/", "https://8.8.8.8/"},
	} {
		actual, err := safety.ValidateDestination(item.input, policy)
		if err != nil || actual != item.expected {
			t.Fatal("destination path or query meaning changed")
		}
	}
	if _, err := safety.ValidateDestination("https://example.com", config.LinksConfig{}); err == nil {
		t.Fatal("missing operator policy accepted")
	}
}
