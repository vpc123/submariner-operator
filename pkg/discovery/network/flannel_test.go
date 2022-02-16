/*
SPDX-License-Identifier: Apache-2.0

Copyright Contributors to the Submariner project.

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

package network_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/submariner-io/submariner-operator/pkg/discovery/network"
	v1apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	NetworkPluginFlannel = "flannel"
	testFlannelPodCIDR   = "10.0.0.0/8"
)

var _ = Describe("Flannel Network", func() {
	When("There are no generic k8s pods to look at", func() {
		It("Should return the ClusterNetwork structure with the pod CIDR and the service CIDR", func() {
			clusterNet := testDiscoverFlannelWith(&flannelDaemonSet, &flannelCfgMap)
			Expect(clusterNet).NotTo(BeNil())
			Expect(clusterNet.NetworkPlugin).To(Equal(NetworkPluginFlannel))
			Expect(clusterNet.PodCIDRs).To(Equal([]string{testFlannelPodCIDR}))
			Expect(clusterNet.ServiceCIDRs).To(Equal([]string{testServiceCIDRFromService}))
		})
	})
})

func testDiscoverFlannelWith(objects ...runtime.Object) *network.ClusterNetwork {
	clientSet := newTestClient(objects...)
	clusterNet, err := network.Discover(nil, clientSet, nil, "")
	Expect(err).NotTo(HaveOccurred())

	return clusterNet
}

var flannelDaemonSet = v1apps.DaemonSet{
	ObjectMeta: v1meta.ObjectMeta{
		Name:      "kube-flannel-ds",
		Namespace: "kube-system",
	},
	Spec: v1apps.DaemonSetSpec{
		Template: v1.PodTemplateSpec{
			ObjectMeta: v1meta.ObjectMeta{},
			Spec: v1.PodSpec{
				Volumes: volumes,
			},
		},
	},
}

var volumes = []v1.Volume{
	{
		Name: "flannel-cfg",
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{Name: "kube-flannel-cfg"},
			},
		},
	},
}

var flannelCfgMap = v1.ConfigMap{
	ObjectMeta: v1meta.ObjectMeta{
		Name:      "kube-flannel-cfg",
		Namespace: "kube-system",
	},
	Data: map[string]string{
		"net-conf.json": `{
			"Network": "10.0.0.0/8",
			"SubnetLen": 20,
			"SubnetMin": "10.10.0.0",
			"SubnetMax": "10.99.0.0",
			"Backend": {
				"Type": "udp",
				"Port": 7890
			}
		}`,
	},
}
