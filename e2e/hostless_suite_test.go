package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	// An Ingress declares a path on a rule without host.
	// The rule belongs to the ingress-nginx default server, which only receives the requests for the hosts no other server matches.
	hostlessIngress = "hostless"
	hostlessPath    = "/hostless"

	// Another Ingress declares a path on a host of its own.
	hostlessKnownHost    = "hostless-known"
	hostlessKnownIngress = "hostless-known"
	hostlessKnownPath    = "/known"

	// No Ingress declares this host.
	hostlessUnknownHost = "hostless-unknown"
)

// HostlessSuite checks which hosts the paths of a rule without host apply to.
type HostlessSuite struct {
	BaseSuite
}

func TestHostlessSuite(t *testing.T) {
	suite.Run(t, new(HostlessSuite))
}

func (s *HostlessSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	// TearDownSuite does not run when SetupSuite fails, and a remaining Ingress would break the suites running afterward.
	s.T().Cleanup(s.cleanup)

	for _, c := range []*Cluster{s.traefik, s.nginx} {
		err := c.DeployIngressWith(ingressTemplateData{
			Name: hostlessKnownIngress,
			Host: hostlessKnownHost + "." + c.Name + ".local",
			Path: hostlessKnownPath,
		})
		require.NoError(s.T(), err, "deploy %s ingress to %s cluster", hostlessKnownIngress, c.Name)
	}

	// A single Ingress serves both controllers: they both see it, and the ingress-nginx admission webhook refuses a second
	// Ingress declaring the same path without host.
	err := s.traefik.DeployIngressWith(ingressTemplateData{
		Name: hostlessIngress,
		Path: hostlessPath,
	})
	require.NoError(s.T(), err, "deploy %s ingress", hostlessIngress)

	for _, c := range []*Cluster{s.traefik, s.nginx} {
		s.waitForStatus(c, hostlessKnownHost, hostlessKnownPath, http.StatusOK)
		s.waitForStatus(c, hostlessUnknownHost, hostlessPath, http.StatusOK)
	}
}

func (s *HostlessSuite) cleanup() {
	if s.T().Failed() {
		s.T().Log(s.traefik.GetIngressControllerLogs(500))
		s.T().Log(s.nginx.GetIngressControllerLogs(500))
	}

	_ = s.traefik.DeleteIngress(hostlessIngress)
	for _, c := range []*Cluster{s.traefik, s.nginx} {
		_ = c.DeleteIngress(hostlessKnownIngress)
	}
}

// waitForStatus waits until the controller answers the given status over HTTP for <host>.<cluster>.local.
// It fails the suite when it never does, as the assertions would then prove nothing.
func (s *HostlessSuite) waitForStatus(c *Cluster, host, path string, status int) {
	s.T().Helper()

	for range 30 {
		resp := c.MakeRequest(s.T(), host+"."+c.Name+".local", http.MethodGet, path, nil, 1, time.Second)
		if resp != nil && resp.StatusCode == status {
			return
		}
		time.Sleep(time.Second)
	}

	s.T().Fatalf("[%s] path %s on host %s does not answer %d after 30 retries", c.Name, path, host, status)
}

// statuses sends the same request to both controllers, over HTTP and HTTPS, and returns the status codes they answered.
func (s *HostlessSuite) statuses(host, path string) (traefikStatuses, nginxStatuses []int) {
	s.T().Helper()

	for _, c := range []*Cluster{s.traefik, s.nginx} {
		httpResp := c.MakeRequest(s.T(), host+"."+c.Name+".local", http.MethodGet, path, nil, 3, time.Second)
		require.NotNil(s.T(), httpResp, "%s: HTTP %s%s", c.Name, host, path)
		httpsResp := c.MakeTLSRequest(s.T(), host+"."+c.Name+".local", http.MethodGet, path, nil, nil, 3, time.Second)
		require.NotNil(s.T(), httpsResp, "%s: HTTPS %s%s", c.Name, host, path)

		statuses := []int{httpResp.StatusCode, httpsResp.StatusCode}
		if c == s.traefik {
			traefikStatuses = statuses
		} else {
			nginxStatuses = statuses
		}
	}

	return traefikStatuses, nginxStatuses
}

// TestUnknownHostServesHostlessPath checks that a host no Ingress declares serves the path of the rule without host.
func (s *HostlessSuite) TestUnknownHostServesHostlessPath() {
	traefikStatuses, nginxStatuses := s.statuses(hostlessUnknownHost, hostlessPath)

	assert.Equal(s.T(), []int{http.StatusOK, http.StatusOK}, nginxStatuses, "nginx")
	assert.Equal(s.T(), nginxStatuses, traefikStatuses, "traefik")
}

// TestKnownHostDoesNotServeHostlessPath checks that a host declared by another Ingress keeps serving its own paths,
// and does not serve the path of the rule without host.
func (s *HostlessSuite) TestKnownHostDoesNotServeHostlessPath() {
	traefikStatuses, nginxStatuses := s.statuses(hostlessKnownHost, hostlessKnownPath)

	assert.Equal(s.T(), []int{http.StatusOK, http.StatusOK}, nginxStatuses, "nginx: %s", hostlessKnownPath)
	assert.Equal(s.T(), nginxStatuses, traefikStatuses, "traefik: %s", hostlessKnownPath)

	traefikStatuses, nginxStatuses = s.statuses(hostlessKnownHost, hostlessPath)

	assert.Equal(s.T(), []int{http.StatusNotFound, http.StatusNotFound}, nginxStatuses, "nginx: %s", hostlessPath)
	assert.Equal(s.T(), nginxStatuses, traefikStatuses, "traefik: %s", hostlessPath)
}
