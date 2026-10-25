// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package ec

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elastic/cloud-sdk-go/pkg/client/deployments"
	"github.com/elastic/terraform-provider-ec/ec/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Certificates are generated for each test, so no private keys or expiring fixtures are committed.
func tlsTestFiles(t *testing.T) (caFile, certFile, keyFile string, serverCert tls.Certificate, roots *x509.CertPool) {
	t.Helper()
	dir := t.TempDir()
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour),
		NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, ca, ca, &caKey.PublicKey, caKey)
	require.NoError(t, err)
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	caFile = filepath.Join(dir, "ca.pem")
	require.NoError(t, os.WriteFile(caFile, caPEM, 0600))
	roots = x509.NewCertPool()
	require.True(t, roots.AppendCertsFromPEM(caPEM))
	for i, name := range []string{"client", "server"} {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		cert := &x509.Certificate{
			SerialNumber: big.NewInt(int64(i + 2)), NotBefore: ca.NotBefore, NotAfter: ca.NotAfter,
			KeyUsage:    x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}
		if name == "server" {
			cert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
			cert.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
		}
		der, err := x509.CreateCertificate(rand.Reader, cert, ca, &key.PublicKey, caKey)
		require.NoError(t, err)
		keyDER, err := x509.MarshalPKCS8PrivateKey(key)
		require.NoError(t, err)
		certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
		keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
		if name == "server" {
			serverCert, err = tls.X509KeyPair(certPEM, keyPEM)
			require.NoError(t, err)
		} else {
			certFile, keyFile = filepath.Join(dir, "client.pem"), filepath.Join(dir, "client-key.pem")
			require.NoError(t, os.WriteFile(certFile, certPEM, 0600))
			require.NoError(t, os.WriteFile(keyFile, keyPEM, 0600))
		}
	}
	return
}

func configureTLSProvider(t *testing.T, config providerConfig) (*Provider, provider.ConfigureResponse) {
	t.Helper()
	ctx := context.Background()
	p := &Provider{}
	schemaResp := provider.SchemaResponse{}
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)
	var value types.Object
	require.False(t, tfsdk.ValueFrom(ctx, &config, schemaResp.Schema.Type(), &value).HasError())
	raw, err := value.ToTerraformValue(ctx)
	require.NoError(t, err)
	resp := provider.ConfigureResponse{}
	p.Configure(ctx, provider.ConfigureRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}, &resp)
	return p, resp
}

func TestProviderMutualTLS(t *testing.T) {
	ca, cert, key, serverCert, roots := tlsTestFiles(t)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.TLS.VerifiedChains)
		assert.Equal(t, "ApiKey secret", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"deployments":[],"items":[]}`)
	}))
	server.TLS = &tls.Config{Certificates: []tls.Certificate{serverCert}, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: roots}
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.StartTLS()
	defer server.Close()
	oldGetenv := util.GetEnv
	t.Cleanup(func() { util.GetEnv = oldGetenv })
	for _, tc := range []struct {
		name      string
		env       bool
		override  bool
		noCA      bool
		noCert    bool
		insecure  bool
		emptyCA   bool
		emptyCert bool
	}{
		{name: "provider configuration"},
		{name: "environment defaults", env: true},
		{name: "provider overrides environment", override: true},
		{name: "empty CA overrides environment", env: true, emptyCA: true},
		{name: "empty client certificate and key override environment", env: true, emptyCert: true},
		{name: "untrusted server is rejected", noCA: true},
		{name: "client certificate is required", noCert: true},
		{name: "insecure still sends client certificate", noCA: true, insecure: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := map[string]string{}
			config := providerConfig{Endpoint: types.StringValue(server.URL), ApiKey: types.StringValue("secret"),
				CACert: types.StringValue(ca), TLSCert: types.StringValue(cert), TLSKey: types.StringValue(key), Insecure: types.BoolValue(tc.insecure)}
			if tc.env {
				env = map[string]string{"EC_CA_CERT": ca, "EC_TLS_CERT": cert, "EC_TLS_KEY": key}
				config.CACert, config.TLSCert, config.TLSKey = types.StringNull(), types.StringNull(), types.StringNull()
			}
			if tc.override {
				env = map[string]string{"EC_CA_CERT": "missing", "EC_TLS_CERT": "missing", "EC_TLS_KEY": "missing"}
			}
			if tc.noCA {
				config.CACert = types.StringNull()
			}
			if tc.noCert {
				config.TLSCert, config.TLSKey = types.StringNull(), types.StringNull()
			}
			if tc.emptyCA {
				config.CACert = types.StringValue("")
			}
			if tc.emptyCert {
				config.TLSCert, config.TLSKey = types.StringValue(""), types.StringValue("")
			}
			util.GetEnv = func(name string) string { return env[name] }
			p, resp := configureTLSProvider(t, config)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			_, hostedErr := p.client.V1API.Deployments.ListDeployments(deployments.NewListDeploymentsParams(), p.client.AuthWriter)
			_, serverlessErr := p.slsClient.ListElasticsearchProjectsWithResponse(context.Background(), nil)
			if tc.noCert || tc.emptyCert || ((tc.noCA || tc.emptyCA) && !tc.insecure) {
				require.Error(t, hostedErr)
				require.Error(t, serverlessErr)
			} else {
				require.NoError(t, hostedErr)
				require.NoError(t, serverlessErr)
			}
		})
	}
}

func TestHTTPClientTLSValidation(t *testing.T) {
	ca, cert, key, _, _ := tlsTestFiles(t)
	_, _, otherKey, _, _ := tlsTestFiles(t)
	invalid := filepath.Join(t.TempDir(), "invalid.pem")
	require.NoError(t, os.WriteFile(invalid, []byte("invalid PEM"), 0600))
	for _, tc := range []struct {
		name  string
		setup apiSetup
		want  string
	}{
		{"certificate without key", apiSetup{tlsCert: cert}, "must be configured together"},
		{"key without certificate", apiSetup{tlsKey: key}, "must be configured together"},
		{"missing CA", apiSetup{caCert: filepath.Join(t.TempDir(), "missing")}, "failed reading ca_cert"},
		{"invalid CA", apiSetup{caCert: invalid}, "valid PEM certificates"},
		{"missing certificate", apiSetup{tlsCert: "missing", tlsKey: key}, "failed loading tls_cert and tls_key"},
		{"missing key", apiSetup{tlsCert: cert, tlsKey: "missing"}, "failed loading tls_cert and tls_key"},
		{"invalid certificate", apiSetup{tlsCert: invalid, tlsKey: key}, "failed loading tls_cert and tls_key"},
		{"invalid key", apiSetup{tlsCert: cert, tlsKey: invalid}, "failed loading tls_cert and tls_key"},
		{"mismatched key", apiSetup{tlsCert: cert, tlsKey: otherKey}, "failed loading tls_cert and tls_key"},
		{"CA only", apiSetup{caCert: ca}, ""},
		{"client certificate only", apiSetup{tlsCert: cert, tlsKey: key}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup.apikey = "secret"
			cfg, err := newAPIConfig(tc.setup)
			if tc.want != "" {
				require.ErrorContains(t, err, tc.want)
				return
			}
			require.NoError(t, err)
			transport := cfg.Client.Transport.(*http.Transport)
			defer transport.CloseIdleConnections()
			assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
			assert.NotNil(t, transport.Proxy)
			defaultTLS := http.DefaultTransport.(*http.Transport).TLSClientConfig
			if defaultTLS != nil {
				assert.Nil(t, defaultTLS.RootCAs)
				assert.Empty(t, defaultTLS.Certificates)
			}
		})
	}
}
