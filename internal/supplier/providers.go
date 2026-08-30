package supplier

const (
	ProviderLegacy = "legacy"
	ProviderAtlas  = "atlas"
	ProviderNexus  = "nexus"
	ProviderVertex = "vertex"
)

var DefaultProviders = []string{
	ProviderAtlas,
	ProviderNexus,
	ProviderVertex,
}
