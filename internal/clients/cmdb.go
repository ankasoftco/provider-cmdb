/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clients

import (
	"context"

	"github.com/go-openapi/runtime"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-cmdb/apis/v1alpha1"

	cmdb "github.com/anka-software/cmdb-sdk/pkg/client"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// Config for ServiceNow authentication struct
type Config struct {
	BaseURL  string
	Username string
	Password string
}

/*
// NewClient creates new ServiceNow with provided Configurations.
func NewClient(c Config) *cmdb.MainClient {
	transport := GetTransport(c)
	apiClient := cmdb.New(transport, strfmt.Default)

	return apiClient
}
*/

// NewClient creates new ServiceNow with provided Configurations and Credentials.
func NewClient(c Config) *cmdb.MainClient {
	transport := GetTransportWithAuthentication(c)
	apiClient := cmdb.New(transport, strfmt.Default)

	return apiClient
}

// GetTransportWithAuthentication returns the REST config with authentication header
func GetTransportWithAuthentication(c Config) runtime.ClientTransport {
	transport := httptransport.New(c.BaseURL, "/api/now", nil)
	transport.SetDebug(true)
	transport.DefaultAuthentication = httptransport.BasicAuth(c.Username, c.Password)

	return transport
}

// GetTransport returns the REST config
func GetTransport(c Config) runtime.ClientTransport {
	transport := httptransport.New(c.BaseURL, "", nil)
	transport.SetDebug(true)

	return transport
}

// GetConfig constructs a Config that can be used to authenticate to vRA
func GetConfig(ctx context.Context, c client.Client, mg resource.Managed) (*Config, error) {
	switch {
	case mg.GetProviderConfigReference() != nil:
		return UseProviderConfig(ctx, c, mg)
	default:
		return nil, errors.New("providerConfigRef is not given")
	}
}

// UseProviderConfig to produce a config that can be used to authenticate to AWS.
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.Managed) (*Config, error) {
	pc := &v1alpha1.ProviderConfig{}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced Provider")
	}

	t := resource.NewProviderConfigUsageTracker(c, &v1alpha1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	switch s := pc.Spec.Credentials.Source; s { //nolint:exhaustive
	case xpv1.CredentialsSourceSecret:
		csr := pc.Spec.Credentials.SecretRef
		if csr == nil {
			return nil, errors.New("no credentials secret referenced")
		}
		s := &corev1.Secret{}
		if err := c.Get(ctx, types.NamespacedName{Namespace: csr.Namespace, Name: csr.Name}, s); err != nil {
			return nil, errors.Wrap(err, "cannot get credentials secret")
		}
		return &Config{BaseURL: pc.Spec.BaseURL, Username: pc.Spec.Username, Password: string(s.Data[csr.Key])}, nil
	default:
		return nil, errors.Errorf("credentials source %s is not currently supported", s)
	}
}
