/*
Copyright 2021 The Crossplane Authors.

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
	"encoding/json"
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/terrajet/pkg/terraform"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-jet-vault/apis/v1alpha1"
)

const (
	keyVaultAddr          = "address"
	keyToken              = "token"
	keyTokenName          = "token_name"
	keyCaCertFile         = "ca_cert_file"
	keyCaCertDir          = "ca_cert_dir"
	keySkipTLSVerify      = "skip_tls_verify"
	keySkipChildToken     = "skip_child_token"
	keyMaxLeaseTTLSeconds = "max_lease_ttl_seconds"
	keyMaxRetries         = "max_retries"
	keyMaxRetriesCcc      = "max_retries_ccc"
	keyNamespace          = "namespace"

	// TODO(@aaronme) These should only be added to the configuration if they
	// are supplied
	// keyAuthLogin          = "auth_login"
	// keyClientAuth         = "client_auth"
	// keyHeaders            = "headers"

	// Vault credentials environment variable names
	envVaultAddr          = "VAULT_ADDR"
	envToken              = "VAULT_TOKEN"
	envTokenName          = "VAULT_TOKEN_NAME"
	envCaCertFile         = "VAULT_CACERT"
	envCaCertDir          = "VAULT_CAPATH"
	envSkipTLSVerify      = "VAULT_SKIP_VERIFY"
	envSkipChildToken     = "TERRAFORM_VAULT_SKIP_CHILD_TOKEN"
	envMaxLeaseTTLSeconds = "TERRAFORM_VAULT_MAX_TTL"
	envMaxRetries         = "VAULT_MAX_RETRIES"
	envMaxRetriesCcc      = "VAULT_MAX_RETRIES_CCC"
	envNamespace          = "VAULT_NAMESPACE"
)

const (
	fmtEnvVar = "%s=%s"

	// error messages
	errNoProviderConfig     = "no providerConfigRef provided"
	errGetProviderConfig    = "cannot get referenced ProviderConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal vault credentials as JSON"
)

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		configRef := mg.GetProviderConfigReference()
		if configRef == nil {
			return ps, errors.New(errNoProviderConfig)
		}
		pc := &v1alpha1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
			return ps, errors.Wrap(err, errGetProviderConfig)
		}

		t := resource.NewProviderConfigUsageTracker(client, &v1alpha1.ProviderConfigUsage{})
		if err := t.Track(ctx, mg); err != nil {
			return ps, errors.Wrap(err, errTrackUsage)
		}

		data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, client, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}
		vaultCreds := map[string]string{}
		if err := json.Unmarshal(data, &vaultCreds); err != nil {
			return ps, errors.Wrap(err, errUnmarshalCredentials)
		}

		// set provider configuration
		ps.Configuration = map[string]interface{}{
			"address": vaultCreds[keyVaultAddr],
		}
		// set environment variables for sensitive provider configuration
		ps.Env = []string{
			fmt.Sprintf(fmtEnvVar, envVaultAddr, vaultCreds[keyVaultAddr]),
			fmt.Sprintf(fmtEnvVar, envToken, vaultCreds[keyToken]),
			fmt.Sprintf(fmtEnvVar, envTokenName, vaultCreds[keyTokenName]),
			fmt.Sprintf(fmtEnvVar, envToken, vaultCreds[keyToken]),
			fmt.Sprintf(fmtEnvVar, envCaCertFile, vaultCreds[keyCaCertFile]),
			fmt.Sprintf(fmtEnvVar, envCaCertDir, vaultCreds[keyCaCertDir]),
			fmt.Sprintf(fmtEnvVar, envSkipTLSVerify, vaultCreds[keySkipTLSVerify]),
			fmt.Sprintf(fmtEnvVar, envSkipChildToken, vaultCreds[keySkipChildToken]),
			fmt.Sprintf(fmtEnvVar, envMaxLeaseTTLSeconds, vaultCreds[keyMaxLeaseTTLSeconds]),
			fmt.Sprintf(fmtEnvVar, envMaxRetries, vaultCreds[keyMaxRetries]),
			fmt.Sprintf(fmtEnvVar, envMaxRetriesCcc, vaultCreds[keyMaxRetriesCcc]),
			fmt.Sprintf(fmtEnvVar, envNamespace, vaultCreds[keyNamespace]),
		}
		return ps, nil
	}
}
