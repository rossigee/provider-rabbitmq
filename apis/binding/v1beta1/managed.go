package v1beta1

import xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"

func (in *Binding) GetProviderConfigReference() *xpv1.ProviderConfigReference {
	return in.Spec.ProviderConfigReference
}
