package v1beta1

import xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"

func (in *ProviderCredentials) DeepCopyInto(out *ProviderCredentials) {
	*out = *in
	in.CommonCredentialSelectors.DeepCopyInto(&out.CommonCredentialSelectors)
}

func (in *TLSConfig) DeepCopyInto(out *TLSConfig) {
	*out = *in
	if in.CABundleSecretRef != nil {
		in, out := &in.CABundleSecretRef, &out.CABundleSecretRef
		*out = new(xpv1.SecretKeySelector)
		**out = **in
	}
}

func (in *ProviderConfigSpec) DeepCopyInto(out *ProviderConfigSpec) {
	*out = *in
	in.Credentials.DeepCopyInto(&out.Credentials)
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(TLSConfig)
		(*in).DeepCopyInto(*out)
	}
}

func (in *ProviderConfigStatus) DeepCopyInto(out *ProviderConfigStatus) {
	*out = *in
	in.ProviderConfigStatus.DeepCopyInto(&out.ProviderConfigStatus)
}
