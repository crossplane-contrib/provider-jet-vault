package generic

import "github.com/crossplane/terrajet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("vault_generic_secret", func(r *config.Resource) {

		// we need to override the default group that terrajet generated for
		// this resource, which would be "vault"
		r.ShortGroup = "generic"

		// we need to map data_json properly
		r.ExternalName = config.IdentifierFromProvider

	})
}
