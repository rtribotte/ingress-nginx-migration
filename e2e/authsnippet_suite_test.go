package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	authMethodGetIngressName = "auth-method-get-test"
	authMethodGetTraefikHost = authMethodGetIngressName + ".traefik.local"
	authMethodGetNginxHost   = authMethodGetIngressName + ".nginx.local"

	authMethodPostIngressName = "auth-method-post-test"
	authMethodPostTraefikHost = authMethodPostIngressName + ".traefik.local"
	authMethodPostNginxHost   = authMethodPostIngressName + ".nginx.local"

	authSnippetProxySetIngressName = "auth-snippet-proxyset-test"
	authSnippetProxySetTraefikHost = authSnippetProxySetIngressName + ".traefik.local"
	authSnippetProxySetNginxHost   = authSnippetProxySetIngressName + ".nginx.local"

	authSnippetAddHeaderIngressName = "auth-snippet-addheader-test"
	authSnippetAddHeaderTraefikHost = authSnippetAddHeaderIngressName + ".traefik.local"
	authSnippetAddHeaderNginxHost   = authSnippetAddHeaderIngressName + ".nginx.local"

	authSnippetProxyMethodIngressName = "auth-snippet-proxymethod-test"
	authSnippetProxyMethodTraefikHost = authSnippetProxyMethodIngressName + ".traefik.local"
	authSnippetProxyMethodNginxHost   = authSnippetProxyMethodIngressName + ".nginx.local"

	authSnippetIfCondIngressName = "auth-snippet-ifcond-test"
	authSnippetIfCondTraefikHost = authSnippetIfCondIngressName + ".traefik.local"
	authSnippetIfCondNginxHost   = authSnippetIfCondIngressName + ".nginx.local"

	authSnippetWithConfigIngressName = "auth-snippet-withconfig-test"
	authSnippetWithConfigTraefikHost = authSnippetWithConfigIngressName + ".traefik.local"
	authSnippetWithConfigNginxHost   = authSnippetWithConfigIngressName + ".nginx.local"

	authSnippetMoreSetIngressName = "auth-snippet-moreset-test"
	authSnippetMoreSetTraefikHost = authSnippetMoreSetIngressName + ".traefik.local"
	authSnippetMoreSetNginxHost   = authSnippetMoreSetIngressName + ".nginx.local"

	authMethodWithSnippetIngressName = "auth-method-withsnippet-test"
	authMethodWithSnippetTraefikHost = authMethodWithSnippetIngressName + ".traefik.local"
	authMethodWithSnippetNginxHost   = authMethodWithSnippetIngressName + ".nginx.local"

	authSnippetServerURL = "http://auth-server.default.svc.cluster.local"
)

type AuthSnippetSuite struct {
	BaseSuite
}

func TestAuthSnippetSuite(t *testing.T) {
	suite.Run(t, new(AuthSnippetSuite))
}

func (s *AuthSnippetSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	// Deploy auth server to both clusters.
	err := s.traefik.ApplyFixture("auth-server.yaml")
	require.NoError(s.T(), err, "deploy auth-server to traefik cluster")

	err = s.nginx.ApplyFixture("auth-server.yaml")
	require.NoError(s.T(), err, "deploy auth-server to nginx cluster")

	// Wait for auth server to be ready.
	err = waitForDeployment(s.traefik, s.traefik.TestNamespace, "auth-server")
	require.NoError(s.T(), err, "auth-server not ready in traefik cluster")

	err = waitForDeployment(s.nginx, s.nginx.TestNamespace, "auth-server")
	require.NoError(s.T(), err, "auth-server not ready in nginx cluster")

	// 1. auth-method GET
	authMethodGetAnnotations := map[string]string{
		"nginx.ingress.kubernetes.io/auth-url":              authSnippetServerURL + "/",
		"nginx.ingress.kubernetes.io/auth-method":           "GET",
		"nginx.ingress.kubernetes.io/auth-response-headers": "X-Echo-Authorization,X-Echo-Cookie,X-Echo-User-Agent,X-Echo-Accept,X-Echo-Test-Header,X-Echo-Proxy-Authorization,X-Echo-Upgrade,X-Echo-TE,X-Echo-Forwarded-For,X-Echo-Forwarded-Host,X-Echo-Forwarded-Proto,X-Echo-Forwarded-Port,X-Echo-Forwarded-Method,X-Echo-Forwarded-Uri,X-Echo-Forwarded-Prefix,X-Echo-Forwarded-Server,X-Echo-Forwarded-Tls-Client-Cert,X-Echo-Forwarded-Tls-Client-Cert-Info,X-Echo-Real-Ip",
	}

	err = s.traefik.DeployIngress(authMethodGetIngressName, authMethodGetTraefikHost, authMethodGetAnnotations)
	require.NoError(s.T(), err, "deploy auth-method-get ingress to traefik cluster")

	err = s.nginx.DeployIngress(authMethodGetIngressName, authMethodGetNginxHost, authMethodGetAnnotations)
	require.NoError(s.T(), err, "deploy auth-method-get ingress to nginx cluster")

	// 2. auth-method POST
	authMethodPostAnnotations := map[string]string{
		"nginx.ingress.kubernetes.io/auth-url":    authSnippetServerURL + "/",
		"nginx.ingress.kubernetes.io/auth-method": "POST",
	}

	err = s.traefik.DeployIngress(authMethodPostIngressName, authMethodPostTraefikHost, authMethodPostAnnotations)
	require.NoError(s.T(), err, "deploy auth-method-post ingress to traefik cluster")

	err = s.nginx.DeployIngress(authMethodPostIngressName, authMethodPostNginxHost, authMethodPostAnnotations)
	require.NoError(s.T(), err, "deploy auth-method-post ingress to nginx cluster")

	// 3. auth-snippet with proxy_set_header
	authSnippetProxySetAnnotations := map[string]string{
		"nginx.ingress.kubernetes.io/auth-url":     authSnippetServerURL + "/",
		"nginx.ingress.kubernetes.io/auth-snippet": `proxy_set_header X-Custom-Auth "auth-value";`,
	}

	err = s.traefik.DeployIngress(authSnippetProxySetIngressName, authSnippetProxySetTraefikHost, authSnippetProxySetAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-proxyset ingress to traefik cluster")

	err = s.nginx.DeployIngress(authSnippetProxySetIngressName, authSnippetProxySetNginxHost, authSnippetProxySetAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-proxyset ingress to nginx cluster")

	// 4. auth-snippet with add_header
	authSnippetAddHeaderAnnotations := map[string]string{
		"nginx.ingress.kubernetes.io/auth-url":     authSnippetServerURL + "/",
		"nginx.ingress.kubernetes.io/auth-snippet": `add_header X-Auth-Debug "debug-value" always;`,
	}

	err = s.traefik.DeployIngress(authSnippetAddHeaderIngressName, authSnippetAddHeaderTraefikHost, authSnippetAddHeaderAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-addheader ingress to traefik cluster")

	err = s.nginx.DeployIngress(authSnippetAddHeaderIngressName, authSnippetAddHeaderNginxHost, authSnippetAddHeaderAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-addheader ingress to nginx cluster")

	// 5. auth-snippet with proxy_method
	authSnippetProxyMethodAnnotations := map[string]string{
		"nginx.ingress.kubernetes.io/auth-url":     authSnippetServerURL + "/",
		"nginx.ingress.kubernetes.io/auth-snippet": `proxy_method GET;`,
	}

	err = s.traefik.DeployIngress(authSnippetProxyMethodIngressName, authSnippetProxyMethodTraefikHost, authSnippetProxyMethodAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-proxymethod ingress to traefik cluster")

	err = s.nginx.DeployIngress(authSnippetProxyMethodIngressName, authSnippetProxyMethodNginxHost, authSnippetProxyMethodAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-proxymethod ingress to nginx cluster")

	// 6. auth-snippet with if condition
	authSnippetIfCondAnnotations := map[string]string{
		"nginx.ingress.kubernetes.io/auth-url": authSnippetServerURL + "/",
		"nginx.ingress.kubernetes.io/auth-snippet": `
if ($request_method = POST) {
    return 403;
}`,
	}

	err = s.traefik.DeployIngress(authSnippetIfCondIngressName, authSnippetIfCondTraefikHost, authSnippetIfCondAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-ifcond ingress to traefik cluster")

	err = s.nginx.DeployIngress(authSnippetIfCondIngressName, authSnippetIfCondNginxHost, authSnippetIfCondAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-ifcond ingress to nginx cluster")

	// 7. auth-snippet + configuration-snippet combined
	authSnippetWithConfigAnnotations := map[string]string{
		"nginx.ingress.kubernetes.io/auth-url":              authSnippetServerURL + "/",
		"nginx.ingress.kubernetes.io/auth-snippet":          `add_header X-Auth-Extra "from-auth" always;`,
		"nginx.ingress.kubernetes.io/configuration-snippet": `add_header X-Config-Extra "from-config" always;`,
	}

	err = s.traefik.DeployIngress(authSnippetWithConfigIngressName, authSnippetWithConfigTraefikHost, authSnippetWithConfigAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-withconfig ingress to traefik cluster")

	err = s.nginx.DeployIngress(authSnippetWithConfigIngressName, authSnippetWithConfigNginxHost, authSnippetWithConfigAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-withconfig ingress to nginx cluster")

	// 8. auth-snippet with more_set_input_headers
	authSnippetMoreSetAnnotations := map[string]string{
		"nginx.ingress.kubernetes.io/auth-url":     authSnippetServerURL + "/",
		"nginx.ingress.kubernetes.io/auth-snippet": `more_set_input_headers "X-Injected: injected-value";`,
	}

	err = s.traefik.DeployIngress(authSnippetMoreSetIngressName, authSnippetMoreSetTraefikHost, authSnippetMoreSetAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-moreset ingress to traefik cluster")

	err = s.nginx.DeployIngress(authSnippetMoreSetIngressName, authSnippetMoreSetNginxHost, authSnippetMoreSetAnnotations)
	require.NoError(s.T(), err, "deploy auth-snippet-moreset ingress to nginx cluster")

	// 9. auth-method + auth-snippet without proxy_method (compatible)
	authMethodWithSnippetAnnotations := map[string]string{
		"nginx.ingress.kubernetes.io/auth-url":     authSnippetServerURL + "/",
		"nginx.ingress.kubernetes.io/auth-method":  "GET",
		"nginx.ingress.kubernetes.io/auth-snippet": `add_header X-Auth-Check "checked" always;`,
	}

	err = s.traefik.DeployIngress(authMethodWithSnippetIngressName, authMethodWithSnippetTraefikHost, authMethodWithSnippetAnnotations)
	require.NoError(s.T(), err, "deploy auth-method-withsnippet ingress to traefik cluster")

	err = s.nginx.DeployIngress(authMethodWithSnippetIngressName, authMethodWithSnippetNginxHost, authMethodWithSnippetAnnotations)
	require.NoError(s.T(), err, "deploy auth-method-withsnippet ingress to nginx cluster")

	// Wait for all ingresses to be ready.
	s.traefik.WaitForIngressReady(s.T(), authMethodGetTraefikHost, 20, 1*time.Second)
	s.nginx.WaitForIngressReady(s.T(), authMethodGetNginxHost, 20, 1*time.Second)
	s.traefik.WaitForIngressReady(s.T(), authMethodPostTraefikHost, 20, 1*time.Second)
	s.nginx.WaitForIngressReady(s.T(), authMethodPostNginxHost, 20, 1*time.Second)
	s.traefik.WaitForIngressReady(s.T(), authSnippetProxySetTraefikHost, 20, 1*time.Second)
	s.nginx.WaitForIngressReady(s.T(), authSnippetProxySetNginxHost, 20, 1*time.Second)
	s.traefik.WaitForIngressReady(s.T(), authSnippetAddHeaderTraefikHost, 20, 1*time.Second)
	s.nginx.WaitForIngressReady(s.T(), authSnippetAddHeaderNginxHost, 20, 1*time.Second)
	s.traefik.WaitForIngressReady(s.T(), authSnippetProxyMethodTraefikHost, 20, 1*time.Second)
	s.nginx.WaitForIngressReady(s.T(), authSnippetProxyMethodNginxHost, 20, 1*time.Second)
	s.traefik.WaitForIngressReady(s.T(), authSnippetIfCondTraefikHost, 20, 1*time.Second)
	s.nginx.WaitForIngressReady(s.T(), authSnippetIfCondNginxHost, 20, 1*time.Second)
	s.traefik.WaitForIngressReady(s.T(), authSnippetWithConfigTraefikHost, 20, 1*time.Second)
	s.nginx.WaitForIngressReady(s.T(), authSnippetWithConfigNginxHost, 20, 1*time.Second)
	s.traefik.WaitForIngressReady(s.T(), authSnippetMoreSetTraefikHost, 20, 1*time.Second)
	s.nginx.WaitForIngressReady(s.T(), authSnippetMoreSetNginxHost, 20, 1*time.Second)
	s.traefik.WaitForIngressReady(s.T(), authMethodWithSnippetTraefikHost, 20, 1*time.Second)
	s.nginx.WaitForIngressReady(s.T(), authMethodWithSnippetNginxHost, 20, 1*time.Second)
}

func (s *AuthSnippetSuite) TearDownSuite() {
	_ = s.traefik.DeleteIngress(authMethodGetIngressName)
	_ = s.nginx.DeleteIngress(authMethodGetIngressName)
	_ = s.traefik.DeleteIngress(authMethodPostIngressName)
	_ = s.nginx.DeleteIngress(authMethodPostIngressName)
	_ = s.traefik.DeleteIngress(authSnippetProxySetIngressName)
	_ = s.nginx.DeleteIngress(authSnippetProxySetIngressName)
	_ = s.traefik.DeleteIngress(authSnippetAddHeaderIngressName)
	_ = s.nginx.DeleteIngress(authSnippetAddHeaderIngressName)
	_ = s.traefik.DeleteIngress(authSnippetProxyMethodIngressName)
	_ = s.nginx.DeleteIngress(authSnippetProxyMethodIngressName)
	_ = s.traefik.DeleteIngress(authSnippetIfCondIngressName)
	_ = s.nginx.DeleteIngress(authSnippetIfCondIngressName)
	_ = s.traefik.DeleteIngress(authSnippetWithConfigIngressName)
	_ = s.nginx.DeleteIngress(authSnippetWithConfigIngressName)
	_ = s.traefik.DeleteIngress(authSnippetMoreSetIngressName)
	_ = s.nginx.DeleteIngress(authSnippetMoreSetIngressName)
	_ = s.traefik.DeleteIngress(authMethodWithSnippetIngressName)
	_ = s.nginx.DeleteIngress(authMethodWithSnippetIngressName)

	// Clean up auth server.
	_ = s.traefik.Kubectl("delete", "-f", fmt.Sprintf("%s/auth-server.yaml", fixturesDir), "-n", s.traefik.TestNamespace, "--ignore-not-found")
	_ = s.nginx.Kubectl("delete", "-f", fmt.Sprintf("%s/auth-server.yaml", fixturesDir), "-n", s.nginx.TestNamespace, "--ignore-not-found")
}

// TestAuthMethodGET verifies that auth-method GET works with auth-url and
// compares which client request headers each controller forwards into the
// auth subrequest by default. Traefik's ForwardAuth copies all original
// request headers except RFC 7230 hop-by-hop headers; this test asserts
// parity with nginx-ingress on a representative set.
func (s *AuthSnippetSuite) TestAuthMethodGET() {
	clientHeaders := map[string]string{
		"Authorization":       "Bearer test-token",
		"Cookie":              "session=abc",
		"User-Agent":          "auth-snippet-e2e",
		"Accept":              "application/json",
		"X-Test-Header":       "test-value",
		"Proxy-Authorization": "Basic dGVzdA==",
		"Upgrade":             "websocket",
		"TE":                  "trailers",
		// Client-supplied X-Forwarded-* — used to observe each controller's
		// trust/overwrite behavior on the auth subrequest.
		"X-Forwarded-For":    "203.0.113.10",
		"X-Forwarded-Host":   "client-set.example.com",
		"X-Forwarded-Proto":  "https",
		"X-Forwarded-Port":   "4443",
		"X-Forwarded-Method": http.MethodDelete,
		"X-Forwarded-Uri":    "/client/uri",
		// Security-advisory header set: members of Traefik's XHeadersSet that
		// the snippet writeHeader does NOT re-set. The maintainers' rationale
		// is that the entrypoint-level XForwarded middleware strips these from
		// untrusted sources before the snippet path ever runs. Markers below
		// must never reach the auth server via the Traefik path.
		"X-Forwarded-Prefix":               "/spoofed-admin",
		"X-Real-Ip":                        "10.0.0.66",
		"X-Forwarded-Server":               "evil.example.com",
		"X-Forwarded-Tls-Client-Cert":      "ATTACKER-CN=admin,O=Evil",
		"X-Forwarded-Tls-Client-Cert-Info": "spoofed-info",
	}

	traefikResp := s.traefik.MakeRequest(s.T(), authMethodGetTraefikHost, http.MethodGet, "/", clientHeaders, 3, 1*time.Second)
	require.NotNil(s.T(), traefikResp, "traefik response should not be nil")

	nginxResp := s.nginx.MakeRequest(s.T(), authMethodGetNginxHost, http.MethodGet, "/", clientHeaders, 3, 1*time.Second)
	require.NotNil(s.T(), nginxResp, "nginx response should not be nil")

	assert.Equal(s.T(), nginxResp.StatusCode, traefikResp.StatusCode, "status code mismatch")
	assert.Equal(s.T(), http.StatusOK, traefikResp.StatusCode,
		"expected 200 when auth-method is GET and auth server allows")
	assert.Equal(s.T(), http.StatusOK, nginxResp.StatusCode,
		"expected 200 when auth-method is GET and auth server allows")

	// Client headers both controllers forward into the auth subrequest by default.
	// Proxy-Authorization is intentionally forwarded by both (despite being hop-by-hop in RFC 7230)
	// because forward-auth servers commonly need it to validate proxy credentials.
	forwardedEchoHeaders := []string{
		"X-Echo-Authorization",
		"X-Echo-Cookie",
		"X-Echo-User-Agent",
		"X-Echo-Accept",
		"X-Echo-Test-Header",
		"X-Echo-Proxy-Authorization",
	}
	for _, h := range forwardedEchoHeaders {
		assert.Equal(s.T(), nginxResp.RequestHeaders[h], traefikResp.RequestHeaders[h],
			"%s: nginx-ingress and traefik should propagate the same value to the auth subrequest", h)
		assert.NotEmpty(s.T(), traefikResp.RequestHeaders[h],
			"traefik should forward %s into the auth subrequest by default", h)
		assert.NotEmpty(s.T(), nginxResp.RequestHeaders[h],
			"nginx-ingress should forward %s into the auth subrequest by default", h)
	}

	// Hop-by-hop headers (RFC 7230) — Traefik strips them from the auth subrequest. Expect nginx-ingress to do the same.
	hopByHopEchoHeaders := []string{
		"X-Echo-Upgrade",
		"X-Echo-TE",
	}
	for _, h := range hopByHopEchoHeaders {
		assert.Equal(s.T(), nginxResp.RequestHeaders[h], traefikResp.RequestHeaders[h],
			"%s: nginx-ingress and traefik should handle the hop-by-hop header identically", h)
		assert.Empty(s.T(), traefikResp.RequestHeaders[h],
			"traefik should strip hop-by-hop %s from the auth subrequest", h)
		assert.Empty(s.T(), nginxResp.RequestHeaders[h],
			"nginx-ingress should strip hop-by-hop %s from the auth subrequest", h)
	}

	// X-Forwarded-* propagation on the auth subrequest. The client sends marker values to
	// exercise each controller's behavior; the two diverge sharply on what reaches the auth
	// server. Traefik strips untrusted client X-Forwarded-* and synthesizes its own values;
	// nginx-ingress only explicitly rewrites a subset (X-Forwarded-For, X-Forwarded-Proto)
	// and lets the rest of the X-Forwarded-* family pass through verbatim from the client
	// via the default proxy_pass_request_headers behaviour.

	// X-Forwarded-Host: Traefik synthesizes from the actual request and overrides the client
	// value. nginx-ingress does not explicitly set this header on the auth subrequest, so the
	// client-supplied value is forwarded as-is — an auth server behind nginx-ingress that
	// trusts this header is vulnerable to spoofing from untrusted clients.
	assert.NotEmpty(s.T(), traefikResp.RequestHeaders["X-Echo-Forwarded-Host"],
		"traefik should populate X-Forwarded-Host on the auth subrequest")
	assert.NotEqual(s.T(), "client-set.example.com", traefikResp.RequestHeaders["X-Echo-Forwarded-Host"],
		"traefik should overwrite client-supplied X-Forwarded-Host")
	assert.Equal(s.T(), "client-set.example.com", nginxResp.RequestHeaders["X-Echo-Forwarded-Host"],
		"nginx-ingress passes the client-supplied X-Forwarded-Host through to the auth subrequest")

	// X-Forwarded-Proto: Traefik synthesizes "http". nginx-ingress explicitly clears the
	// header on the auth subrequest (proxy_set_header X-Forwarded-Proto "") so it is dropped.
	assert.Equal(s.T(), "http", traefikResp.RequestHeaders["X-Echo-Forwarded-Proto"],
		"traefik should set X-Forwarded-Proto to the actual request scheme")
	assert.Empty(s.T(), nginxResp.RequestHeaders["X-Echo-Forwarded-Proto"],
		"nginx-ingress clears X-Forwarded-Proto on the auth subrequest")

	// X-Forwarded-Port: same as Host. Traefik synthesizes; nginx-ingress passes through.
	assert.NotEmpty(s.T(), traefikResp.RequestHeaders["X-Echo-Forwarded-Port"],
		"traefik should populate X-Forwarded-Port on the auth subrequest")
	assert.NotEqual(s.T(), "4443", traefikResp.RequestHeaders["X-Echo-Forwarded-Port"],
		"traefik should overwrite client-supplied X-Forwarded-Port")
	assert.Equal(s.T(), "4443", nginxResp.RequestHeaders["X-Echo-Forwarded-Port"],
		"nginx-ingress passes the client-supplied X-Forwarded-Port through to the auth subrequest")

	// X-Forwarded-For: nginx-ingress explicitly sets it to $remote_addr on the auth subrequest,
	// dropping the client chain entirely (not appending). Traefik also drops the client value
	// because the source is untrusted by default.
	assert.NotEmpty(s.T(), traefikResp.RequestHeaders["X-Echo-Forwarded-For"],
		"traefik should populate X-Forwarded-For on the auth subrequest")
	assert.NotEmpty(s.T(), nginxResp.RequestHeaders["X-Echo-Forwarded-For"],
		"nginx-ingress should populate X-Forwarded-For on the auth subrequest")
	assert.NotContains(s.T(), traefikResp.RequestHeaders["X-Echo-Forwarded-For"], "203.0.113.10",
		"traefik should drop the client-supplied X-Forwarded-For from an untrusted source")
	assert.NotContains(s.T(), nginxResp.RequestHeaders["X-Echo-Forwarded-For"], "203.0.113.10",
		"nginx-ingress should replace X-Forwarded-For with $remote_addr, not append")

	// X-Forwarded-Method / X-Forwarded-Uri: Traefik synthesizes and overwrites; nginx-ingressa
	// does not synthesize these names (it uses X-Original-Method / X-Original-URL) so the
	// client-supplied value is forwarded as-is.
	assert.Equal(s.T(), http.MethodGet, traefikResp.RequestHeaders["X-Echo-Forwarded-Method"],
		"traefik should set X-Forwarded-Method to the actual request method, overriding the client value")
	assert.Equal(s.T(), http.MethodDelete, nginxResp.RequestHeaders["X-Echo-Forwarded-Method"],
		"nginx-ingress does not synthesize X-Forwarded-Method and should forward the client value as-is")
	assert.NotEqual(s.T(), "/client/uri", traefikResp.RequestHeaders["X-Echo-Forwarded-Uri"],
		"traefik should set X-Forwarded-Uri to the actual URI, overriding the client value")
	assert.Equal(s.T(), "/client/uri", nginxResp.RequestHeaders["X-Echo-Forwarded-Uri"],
		"nginx-ingress does not synthesize X-Forwarded-Uri and should forward the client value as-is")

	// Security-advisory header set — entrypoint-level XForwarded must strip every member
	// of XHeadersSet (forwarded_header.go) before the snippet auth subrequest runs.
	// The snippet writeHeader does not re-set these five headers; the safety claim is that
	// they're already gone from req.Header by the time the snippet sees the request, for
	// any untrusted source. If any of the markers below shows up on the Traefik side, the
	// rationale is broken and the advisories' attack model is reproducible.
	advisoryMarkers := []struct{ echo, clientMarker string }{
		{"X-Echo-Forwarded-Prefix", "/spoofed-admin"},
		{"X-Echo-Real-Ip", "10.0.0.66"},
		{"X-Echo-Forwarded-Server", "evil.example.com"},
		{"X-Echo-Forwarded-Tls-Client-Cert", "ATTACKER-CN=admin,O=Evil"},
		{"X-Echo-Forwarded-Tls-Client-Cert-Info", "spoofed-info"},
	}
	for _, c := range advisoryMarkers {
		assert.NotContains(s.T(), traefikResp.RequestHeaders[c.echo], c.clientMarker,
			"traefik: untrusted-source entry should strip client-supplied %s before the snippet auth subrequest", c.echo)
	}

	// nginx-ingress side: X-Real-IP is overridden by proxy_set_header X-Real-IP $remote_addr
	// in the auth subrequest template; the other four are not touched and pass through.
	assert.NotEqual(s.T(), "10.0.0.66", nginxResp.RequestHeaders["X-Echo-Real-Ip"],
		"nginx-ingress should override X-Real-Ip with $remote_addr on the auth subrequest")
	assert.NotEmpty(s.T(), nginxResp.RequestHeaders["X-Echo-Real-Ip"],
		"nginx-ingress should populate X-Real-Ip on the auth subrequest")
	assert.Equal(s.T(), "/spoofed-admin", nginxResp.RequestHeaders["X-Echo-Forwarded-Prefix"],
		"nginx-ingress passes client-supplied X-Forwarded-Prefix through to the auth subrequest")
	assert.Equal(s.T(), "evil.example.com", nginxResp.RequestHeaders["X-Echo-Forwarded-Server"],
		"nginx-ingress passes client-supplied X-Forwarded-Server through to the auth subrequest")
	assert.Equal(s.T(), "ATTACKER-CN=admin,O=Evil", nginxResp.RequestHeaders["X-Echo-Forwarded-Tls-Client-Cert"],
		"nginx-ingress passes client-supplied X-Forwarded-Tls-Client-Cert through to the auth subrequest")
	assert.Equal(s.T(), "spoofed-info", nginxResp.RequestHeaders["X-Echo-Forwarded-Tls-Client-Cert-Info"],
		"nginx-ingress passes client-supplied X-Forwarded-Tls-Client-Cert-Info through to the auth subrequest")

	// Underscore-alias regression check (GHSA-5m6w-wvh7-57vm). Send X_Forwarded_Proto in a
	// second request and verify Traefik's entrypoint maps the underscore form back to the
	// canonical X-Forwarded-Proto via isManagedXHeader (forwarded_header.go:53-66) and
	// strips it, so the snippet's own X-Forwarded-Proto value ("http") is what reaches the
	// auth server — not the client's injected scheme.
	underscoreHeaders := map[string]string{
		"X_Forwarded_Proto": "client-injected-scheme",
	}
	traefikUnderscoreResp := s.traefik.MakeRequest(s.T(), authMethodGetTraefikHost, http.MethodGet, "/", underscoreHeaders, 3, 1*time.Second)
	require.NotNil(s.T(), traefikUnderscoreResp, "traefik underscore-alias response should not be nil")
	assert.Equal(s.T(), http.StatusOK, traefikUnderscoreResp.StatusCode,
		"underscore-alias request should still pass auth")
	assert.Equal(s.T(), "http", traefikUnderscoreResp.RequestHeaders["X-Echo-Forwarded-Proto"],
		"traefik entrypoint must strip X_Forwarded_Proto underscore alias; snippet should set X-Forwarded-Proto=http")
}

// TestAuthMethodPOST verifies that auth-method POST works with auth-url.
// The auth server allows all methods on /, so the request should succeed.
func (s *AuthSnippetSuite) TestAuthMethodPOST() {
	traefikResp := s.traefik.MakeRequest(s.T(), authMethodPostTraefikHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), traefikResp, "traefik response should not be nil")

	nginxResp := s.nginx.MakeRequest(s.T(), authMethodPostNginxHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), nginxResp, "nginx response should not be nil")

	assert.Equal(s.T(), nginxResp.StatusCode, traefikResp.StatusCode, "status code mismatch")
	assert.Equal(s.T(), http.StatusOK, traefikResp.StatusCode,
		"expected 200 when auth-method is POST and auth server allows")
	assert.Equal(s.T(), http.StatusOK, nginxResp.StatusCode,
		"expected 200 when auth-method is POST and auth server allows")
}

// TestAuthSnippetProxySetHeader verifies that auth-snippet with proxy_set_header
// sends the custom header to the auth server in the auth subrequest.
func (s *AuthSnippetSuite) TestAuthSnippetProxySetHeader() {
	traefikResp := s.traefik.MakeRequest(s.T(), authSnippetProxySetTraefikHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), traefikResp, "traefik response should not be nil")

	nginxResp := s.nginx.MakeRequest(s.T(), authSnippetProxySetNginxHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), nginxResp, "nginx response should not be nil")

	// The auth server returns 200 on /, so the request should pass through
	// to the backend. The proxy_set_header applies to the auth subrequest,
	// so we verify the overall request succeeds.
	assert.Equal(s.T(), nginxResp.StatusCode, traefikResp.StatusCode, "status code mismatch")
	assert.Equal(s.T(), http.StatusOK, traefikResp.StatusCode,
		"expected 200 when auth-snippet sets proxy header on auth subrequest")
	assert.Equal(s.T(), http.StatusOK, nginxResp.StatusCode,
		"expected 200 when auth-snippet sets proxy header on auth subrequest")
}

// TestAuthSnippetAddHeader verifies that auth-snippet with add_header
// does NOT add headers to the client response (add_header in auth-snippet
// applies to the auth subrequest context, not the main response).
// We verify auth passes successfully and coexists with the snippet.
func (s *AuthSnippetSuite) TestAuthSnippetAddHeader() {
	traefikResp := s.traefik.MakeRequest(s.T(), authSnippetAddHeaderTraefikHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), traefikResp, "traefik response should not be nil")

	nginxResp := s.nginx.MakeRequest(s.T(), authSnippetAddHeaderNginxHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), nginxResp, "nginx response should not be nil")

	assert.Equal(s.T(), nginxResp.StatusCode, traefikResp.StatusCode, "status code mismatch")
	assert.Equal(s.T(), http.StatusOK, traefikResp.StatusCode,
		"expected 200 when auth passes with add_header snippet")
}

// TestAuthSnippetProxyMethod verifies that auth-snippet with proxy_method
// overrides the method used for the auth subrequest.
func (s *AuthSnippetSuite) TestAuthSnippetProxyMethod() {
	traefikResp := s.traefik.MakeRequest(s.T(), authSnippetProxyMethodTraefikHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), traefikResp, "traefik response should not be nil")

	nginxResp := s.nginx.MakeRequest(s.T(), authSnippetProxyMethodNginxHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), nginxResp, "nginx response should not be nil")

	assert.Equal(s.T(), nginxResp.StatusCode, traefikResp.StatusCode, "status code mismatch")
	assert.Equal(s.T(), http.StatusOK, traefikResp.StatusCode,
		"expected 200 when auth-snippet sets proxy_method GET")
	assert.Equal(s.T(), http.StatusOK, nginxResp.StatusCode,
		"expected 200 when auth-snippet sets proxy_method GET")
}

// TestAuthSnippetIfCondition verifies that auth-snippet with an if block
// can conditionally alter behavior. GET should pass through (200),
// POST should be blocked (403) by the if condition in the auth-snippet.
func (s *AuthSnippetSuite) TestAuthSnippetIfCondition() {
	// GET request should pass through auth and reach the backend.
	traefikRespGet := s.traefik.MakeRequest(s.T(), authSnippetIfCondTraefikHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), traefikRespGet, "traefik GET response should not be nil")

	nginxRespGet := s.nginx.MakeRequest(s.T(), authSnippetIfCondNginxHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), nginxRespGet, "nginx GET response should not be nil")

	assert.Equal(s.T(), nginxRespGet.StatusCode, traefikRespGet.StatusCode, "GET status code mismatch")
	assert.Equal(s.T(), http.StatusOK, traefikRespGet.StatusCode,
		"expected 200 for GET when auth-snippet if condition does not match")
	assert.Equal(s.T(), http.StatusOK, nginxRespGet.StatusCode,
		"expected 200 for GET when auth-snippet if condition does not match")

	// POST request should be blocked by the if condition.
	traefikRespPost := s.traefik.MakeRequest(s.T(), authSnippetIfCondTraefikHost, http.MethodPost, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), traefikRespPost, "traefik POST response should not be nil")

	nginxRespPost := s.nginx.MakeRequest(s.T(), authSnippetIfCondNginxHost, http.MethodPost, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), nginxRespPost, "nginx POST response should not be nil")

	assert.Equal(s.T(), nginxRespPost.StatusCode, traefikRespPost.StatusCode, "POST status code mismatch")
	assert.Equal(s.T(), http.StatusForbidden, traefikRespPost.StatusCode,
		"expected 403 for POST when auth-snippet if condition matches")
	assert.Equal(s.T(), http.StatusForbidden, nginxRespPost.StatusCode,
		"expected 403 for POST when auth-snippet if condition matches")
}

// TestAuthSnippetWithConfigSnippet verifies that auth-snippet and
// configuration-snippet can be used together. Both should add their
// respective headers to the response.
func (s *AuthSnippetSuite) TestAuthSnippetWithConfigSnippet() {
	traefikResp := s.traefik.MakeRequest(s.T(), authSnippetWithConfigTraefikHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), traefikResp, "traefik response should not be nil")

	nginxResp := s.nginx.MakeRequest(s.T(), authSnippetWithConfigNginxHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), nginxResp, "nginx response should not be nil")

	assert.Equal(s.T(), nginxResp.StatusCode, traefikResp.StatusCode, "status code mismatch")
	assert.Equal(s.T(), http.StatusOK, traefikResp.StatusCode,
		"expected 200 when both auth-snippet and configuration-snippet are set")

	// auth-snippet add_header does NOT affect client response (auth subrequest context only).
	// Only configuration-snippet add_header affects the client response.
	assert.Equal(s.T(), "from-config", traefikResp.ResponseHeaders.Get("X-Config-Extra"),
		"traefik should include X-Config-Extra header from configuration-snippet")
	assert.Equal(s.T(), "from-config", nginxResp.ResponseHeaders.Get("X-Config-Extra"),
		"nginx should include X-Config-Extra header from configuration-snippet")
}

// TestAuthSnippetMoreSetInputHeaders verifies that auth-snippet with
// more_set_input_headers does not break auth. The directive modifies the auth
// subrequest headers, not the upstream request, so we can only verify auth succeeds.
func (s *AuthSnippetSuite) TestAuthSnippetMoreSetInputHeaders() {
	traefikResp := s.traefik.MakeRequest(s.T(), authSnippetMoreSetTraefikHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), traefikResp, "traefik response should not be nil")

	nginxResp := s.nginx.MakeRequest(s.T(), authSnippetMoreSetNginxHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), nginxResp, "nginx response should not be nil")

	assert.Equal(s.T(), nginxResp.StatusCode, traefikResp.StatusCode, "status code mismatch")
	assert.Equal(s.T(), http.StatusOK, traefikResp.StatusCode,
		"expected 200 when auth-snippet uses more_set_input_headers")
}

// TestAuthMethodWithSnippet verifies that auth-method and auth-snippet can
// be used together when auth-snippet does not contain proxy_method.
// add_header in auth-snippet applies to the auth subrequest, not the client response.
func (s *AuthSnippetSuite) TestAuthMethodWithSnippet() {
	traefikResp := s.traefik.MakeRequest(s.T(), authMethodWithSnippetTraefikHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), traefikResp, "traefik response should not be nil")

	nginxResp := s.nginx.MakeRequest(s.T(), authMethodWithSnippetNginxHost, http.MethodGet, "/", nil, 3, 1*time.Second)
	require.NotNil(s.T(), nginxResp, "nginx response should not be nil")

	assert.Equal(s.T(), nginxResp.StatusCode, traefikResp.StatusCode, "status code mismatch")
	assert.Equal(s.T(), http.StatusOK, traefikResp.StatusCode,
		"expected 200 when auth-method GET and auth-snippet add_header are combined")
	assert.Equal(s.T(), http.StatusOK, nginxResp.StatusCode,
		"expected 200 when auth-method GET and auth-snippet add_header are combined")
}
