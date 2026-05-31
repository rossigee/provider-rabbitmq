package v1beta1

func (in *PermissionSpec) DeepCopyInto(out *PermissionSpec) {
	*out = *in
	in.ManagedResourceSpec.DeepCopyInto(&out.ManagedResourceSpec)
}

func (in *PermissionStatus) DeepCopyInto(out *PermissionStatus) {
	*out = *in
	in.ConditionedStatus.DeepCopyInto(&out.ConditionedStatus)
}
