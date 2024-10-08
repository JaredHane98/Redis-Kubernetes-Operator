//go:build !ignore_autogenerated

/*
Copyright 2024.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MapWrapper) DeepCopyInto(out *MapWrapper) {
	clone := in.DeepCopy()
	*out = *clone
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisConfigMapWrapper) DeepCopyInto(out *RedisConfigMapWrapper) {
	*out = *in
	in.MapWrapper.DeepCopyInto(&out.MapWrapper)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisConfigMapWrapper.
func (in *RedisConfigMapWrapper) DeepCopy() *RedisConfigMapWrapper {
	if in == nil {
		return nil
	}
	out := new(RedisConfigMapWrapper)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisConfigurationData) DeepCopyInto(out *RedisConfigurationData) {
	*out = *in
	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisConfigurationData.
func (in *RedisConfigurationData) DeepCopy() *RedisConfigurationData {
	if in == nil {
		return nil
	}
	out := new(RedisConfigurationData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisReplication) DeepCopyInto(out *RedisReplication) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisReplication.
func (in *RedisReplication) DeepCopy() *RedisReplication {
	if in == nil {
		return nil
	}
	out := new(RedisReplication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RedisReplication) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisReplicationConfiguration) DeepCopyInto(out *RedisReplicationConfiguration) {
	*out = *in
	in.RedisConfigurationData.DeepCopyInto(&out.RedisConfigurationData)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisReplicationConfiguration.
func (in *RedisReplicationConfiguration) DeepCopy() *RedisReplicationConfiguration {
	if in == nil {
		return nil
	}
	out := new(RedisReplicationConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisReplicationList) DeepCopyInto(out *RedisReplicationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RedisReplication, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisReplicationList.
func (in *RedisReplicationList) DeepCopy() *RedisReplicationList {
	if in == nil {
		return nil
	}
	out := new(RedisReplicationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RedisReplicationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisReplicationSentinelConfig) DeepCopyInto(out *RedisReplicationSentinelConfig) {
	*out = *in
	if in.RedisSentinelDowntime != nil {
		in, out := &in.RedisSentinelDowntime, &out.RedisSentinelDowntime
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisReplicationSentinelConfig.
func (in *RedisReplicationSentinelConfig) DeepCopy() *RedisReplicationSentinelConfig {
	if in == nil {
		return nil
	}
	out := new(RedisReplicationSentinelConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisReplicationSpec) DeepCopyInto(out *RedisReplicationSpec) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.VolumeMounts != nil {
		in, out := &in.VolumeMounts, &out.VolumeMounts
		*out = make([]corev1.VolumeMount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.StatefulsetConfig.DeepCopyInto(&out.StatefulsetConfig)
	in.RedisConfig.DeepCopyInto(&out.RedisConfig)
	if in.TLSConfig != nil {
		in, out := &in.TLSConfig, &out.TLSConfig
		*out = new(RedisTLSConfiguration)
		**out = **in
	}
	if in.RedisSentinelConfig != nil {
		in, out := &in.RedisSentinelConfig, &out.RedisSentinelConfig
		*out = new(RedisReplicationSentinelConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisReplicationSpec.
func (in *RedisReplicationSpec) DeepCopy() *RedisReplicationSpec {
	if in == nil {
		return nil
	}
	out := new(RedisReplicationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisReplicationStatus) DeepCopyInto(out *RedisReplicationStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisReplicationStatus.
func (in *RedisReplicationStatus) DeepCopy() *RedisReplicationStatus {
	if in == nil {
		return nil
	}
	out := new(RedisReplicationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisSentinel) DeepCopyInto(out *RedisSentinel) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisSentinel.
func (in *RedisSentinel) DeepCopy() *RedisSentinel {
	if in == nil {
		return nil
	}
	out := new(RedisSentinel)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RedisSentinel) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisSentinelConfiguration) DeepCopyInto(out *RedisSentinelConfiguration) {
	*out = *in
	in.RedisConfigurationData.DeepCopyInto(&out.RedisConfigurationData)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisSentinelConfiguration.
func (in *RedisSentinelConfiguration) DeepCopy() *RedisSentinelConfiguration {
	if in == nil {
		return nil
	}
	out := new(RedisSentinelConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisSentinelList) DeepCopyInto(out *RedisSentinelList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RedisSentinel, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisSentinelList.
func (in *RedisSentinelList) DeepCopy() *RedisSentinelList {
	if in == nil {
		return nil
	}
	out := new(RedisSentinelList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RedisSentinelList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisSentinelSpec) DeepCopyInto(out *RedisSentinelSpec) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	in.StatefulsetConfig.DeepCopyInto(&out.StatefulsetConfig)
	in.RedisConfig.DeepCopyInto(&out.RedisConfig)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisSentinelSpec.
func (in *RedisSentinelSpec) DeepCopy() *RedisSentinelSpec {
	if in == nil {
		return nil
	}
	out := new(RedisSentinelSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisSentinelStatus) DeepCopyInto(out *RedisSentinelStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisSentinelStatus.
func (in *RedisSentinelStatus) DeepCopy() *RedisSentinelStatus {
	if in == nil {
		return nil
	}
	out := new(RedisSentinelStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisTLSConfiguration) DeepCopyInto(out *RedisTLSConfiguration) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisTLSConfiguration.
func (in *RedisTLSConfiguration) DeepCopy() *RedisTLSConfiguration {
	if in == nil {
		return nil
	}
	out := new(RedisTLSConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StatefulSetConfiguration) DeepCopyInto(out *StatefulSetConfiguration) {
	*out = *in
	in.Wrapper.DeepCopyInto(&out.Wrapper)
	in.MetaData.DeepCopyInto(&out.MetaData)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StatefulSetConfiguration.
func (in *StatefulSetConfiguration) DeepCopy() *StatefulSetConfiguration {
	if in == nil {
		return nil
	}
	out := new(StatefulSetConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StatefulSetMetadataWrapper) DeepCopyInto(out *StatefulSetMetadataWrapper) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StatefulSetMetadataWrapper.
func (in *StatefulSetMetadataWrapper) DeepCopy() *StatefulSetMetadataWrapper {
	if in == nil {
		return nil
	}
	out := new(StatefulSetMetadataWrapper)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StatefulSpecWrapper) DeepCopyInto(out *StatefulSpecWrapper) {
	clone := in.DeepCopy()
	*out = *clone
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSConfig) DeepCopyInto(out *TLSConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSConfig.
func (in *TLSConfig) DeepCopy() *TLSConfig {
	if in == nil {
		return nil
	}
	out := new(TLSConfig)
	in.DeepCopyInto(out)
	return out
}
