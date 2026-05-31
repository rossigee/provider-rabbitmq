package v1beta1

func (in *UserParameters) DeepCopyInto(out *UserParameters) {
	*out = *in
	if in.PasswordSecretRef != nil {
		x := *in.PasswordSecretRef
		out.PasswordSecretRef = &x
	}
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

func (in *UserObservation) DeepCopyInto(out *UserObservation) {
	*out = *in
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

func (in *UserSpec) DeepCopyInto(out *UserSpec) {
	*out = *in
	in.ManagedResourceSpec.DeepCopyInto(&out.ManagedResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

func (in *UserStatus) DeepCopyInto(out *UserStatus) {
	*out = *in
	in.ConditionedStatus.DeepCopyInto(&out.ConditionedStatus)
	in.AtProvider.DeepCopyInto(&out.AtProvider)
}
