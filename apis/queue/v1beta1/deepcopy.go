package v1beta1

func (in *QueueParameters) DeepCopyInto(out *QueueParameters) {
	*out = *in
	if in.Arguments != nil {
		in, out := &in.Arguments, &out.Arguments
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

func (in *QueueObservation) DeepCopyInto(out *QueueObservation) {
	*out = *in
	if in.Arguments != nil {
		in, out := &in.Arguments, &out.Arguments
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

func (in *QueueSpec) DeepCopyInto(out *QueueSpec) {
	*out = *in
	in.ManagedResourceSpec.DeepCopyInto(&out.ManagedResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

func (in *QueueStatus) DeepCopyInto(out *QueueStatus) {
	*out = *in
	in.ConditionedStatus.DeepCopyInto(&out.ConditionedStatus)
	in.AtProvider.DeepCopyInto(&out.AtProvider)
}
