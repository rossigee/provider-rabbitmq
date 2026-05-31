package v1beta1

func (in *BindingParameters) DeepCopyInto(out *BindingParameters) {
	*out = *in
	if in.Arguments != nil {
		in, out := &in.Arguments, &out.Arguments
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

func (in *BindingObservation) DeepCopyInto(out *BindingObservation) {
	*out = *in
	if in.Arguments != nil {
		in, out := &in.Arguments, &out.Arguments
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

func (in *BindingSpec) DeepCopyInto(out *BindingSpec) {
	*out = *in
	in.ManagedResourceSpec.DeepCopyInto(&out.ManagedResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

func (in *BindingStatus) DeepCopyInto(out *BindingStatus) {
	*out = *in
	in.ConditionedStatus.DeepCopyInto(&out.ConditionedStatus)
	in.AtProvider.DeepCopyInto(&out.AtProvider)
}
