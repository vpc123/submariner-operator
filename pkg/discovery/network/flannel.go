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

package network

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// nolint:nilnil // Intentional as the purpose is to discover.
func discoverFlannelNetwork(clientSet kubernetes.Interface) (*ClusterNetwork, error) {
	daemonsets, err := clientSet.AppsV1().DaemonSets(metav1.NamespaceSystem).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, errors.WithMessage(err, "error obtaining the \"flannel\" Daemonset")
	}

	volumes := make([]v1.Volume, 0)
	// look for a daemonset matching "flannel"
	for k := range daemonsets.Items {
		if strings.Contains(daemonsets.Items[k].Name, "flannel") {
			volumes = daemonsets.Items[k].Spec.Template.Spec.Volumes
		}
	}

	if len(volumes) < 1 {
		return nil, nil
	}

	var flannelConfigMap string
	// look for the associated confimap to the flannel daemonset
	for k := range volumes {
		if strings.Contains(volumes[k].Name, "flannel") {
			if volumes[k].ConfigMap.Name != "" {
				flannelConfigMap = volumes[k].ConfigMap.Name
			}
		}
	}

	if flannelConfigMap == "" {
		return nil, nil
	}

	// look for the configmap details using the configmap name discovered from the daemonset
	cm, err := clientSet.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(context.TODO(), flannelConfigMap, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, errors.WithMessage(err, "error obtaining the \"flannel\" ConfigMap")
	}

	podCIDR := extractPodCIDRFromNetConfigJSON(cm)
	if podCIDR == nil {
		return nil, nil
	}

	clusterNetwork := &ClusterNetwork{
		NetworkPlugin: "flannel",
		PodCIDRs:      []string{*podCIDR},
	}

	// Try to detect the service CIDRs using the generic functions
	clusterIPRange, err := findClusterIPRange(clientSet)
	if err != nil {
		return nil, err
	}

	if clusterIPRange != "" {
		clusterNetwork.ServiceCIDRs = []string{clusterIPRange}
	}

	return clusterNetwork, nil
}
