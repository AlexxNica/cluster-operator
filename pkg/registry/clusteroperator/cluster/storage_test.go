/*
Copyright 2017 The Kubernetes Authors.

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

package cluster

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	genericregistrytest "k8s.io/apiserver/pkg/registry/generic/testing"
	etcdtesting "k8s.io/apiserver/pkg/storage/etcd/testing"

	clusteroperatorapi "github.com/openshift/cluster-operator/pkg/apis/clusteroperator"
	"github.com/openshift/cluster-operator/pkg/registry/registrytest"
)

func newStorage(t *testing.T) (*genericregistry.Store, *etcdtesting.EtcdTestServer) {
	etcdStorage, server := registrytest.NewEtcdStorage(t, clusteroperatorapi.GroupName)
	restOptions := generic.RESTOptions{
		StorageConfig:           etcdStorage,
		Decorator:               generic.UndecoratedStorage,
		DeleteCollectionWorkers: 1,
		ResourcePrefix:          "clusters",
	}
	clusterStorage, _ := NewStorage(restOptions)
	return clusterStorage, server
}

func validNewCluster(name string) *clusteroperatorapi.Cluster {
	return &clusteroperatorapi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: clusteroperatorapi.ClusterSpec{
			MachineSets: []clusteroperatorapi.ClusterMachineSet{
				{
					Name: "master",
					MachineSetConfig: clusteroperatorapi.MachineSetConfig{
						NodeType: clusteroperatorapi.NodeTypeMaster,
						Infra:    true,
						Size:     1,
					},
				},
			},
		},
	}
}

func validChangedCluster() *clusteroperatorapi.Cluster {
	return validNewCluster("foo")
}

func TestCreate(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	defer storage.DestroyFunc()
	test := genericregistrytest.New(t, storage)
	cluster := validNewCluster("foo")
	cluster.ObjectMeta = metav1.ObjectMeta{GenerateName: "foo"}
	test.TestCreate(
		// valid
		cluster,
		// invalid
		&clusteroperatorapi.Cluster{
			ObjectMeta: metav1.ObjectMeta{Name: "*BadName!"},
		},
	)
}

func TestUpdate(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	defer storage.DestroyFunc()
	test := genericregistrytest.New(t, storage)
	test.TestUpdate(
		// valid
		validNewCluster("foo"),
		// updateFunc
		func(obj runtime.Object) runtime.Object {
			object := obj.(*clusteroperatorapi.Cluster)
			return object
		},
		//invalid update
		func(obj runtime.Object) runtime.Object {
			object := obj.(*clusteroperatorapi.Cluster)
			return object
		},
	)
}

func TestDelete(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	defer storage.DestroyFunc()
	test := genericregistrytest.New(t, storage).ReturnDeletedObject()
	test.TestDelete(validNewCluster("foo"))
}

func TestGet(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	defer storage.DestroyFunc()
	test := genericregistrytest.New(t, storage)
	test.TestGet(validNewCluster("foo"))
}

func TestList(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	defer storage.DestroyFunc()
	test := genericregistrytest.New(t, storage)
	test.TestList(validNewCluster("foo"))
}

// TODO: Figure out how to make this test pass.
//func TestWatch(t *testing.T) {
//	storage, server := newStorage(t)
//	defer server.Terminate(t)
//	defer storage.DestroyFunc()
//	test := genericregistrytest.New(t, storage)
//	test.TestWatch(
//		validNewCluster("foo"),
//		// matching labels
//		[]labels.Set{},
//		// not matching labels
//		[]labels.Set{
//			{"foo": "bar"},
//		},
//		// matching fields
//		[]fields.Set{
//			{"metadata.name": "foo"},
//		},
//		// not matching fields
//		[]fields.Set{
//			{"metadata.name": "bar"},
//		},
//	)
//}
