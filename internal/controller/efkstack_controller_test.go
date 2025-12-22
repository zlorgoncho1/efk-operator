/*
Copyright 2025.

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

package controller

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	loggingv1 "github.com/zlorgoncho1/efk-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEFKStackController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EFKStack Controller Suite")
}

var _ = Describe("EFKStack Controller", func() {
	var (
		ctx        context.Context
		fakeClient client.Client
		reconciler *EFKStackReconciler
		scheme     *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		_ = loggingv1.AddToScheme(scheme)
		_ = corev1.AddToScheme(scheme)

		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			Build()

		reconciler = &EFKStackReconciler{
			Client: fakeClient,
			Scheme: scheme,
		}
	})

	Context("When creating an EFKStack", func() {
		It("Should set default phase to Pending", func() {
			efkStack := &loggingv1.EFKStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-efk",
					Namespace: "default",
				},
				Spec: loggingv1.EFKStackSpec{
					Elasticsearch: loggingv1.ElasticsearchSpec{
						Version:  "8.11.0",
						Replicas: 3,
					},
					FluentBit: loggingv1.FluentBitSpec{
						Version: "2.2.0",
					},
					Kibana: loggingv1.KibanaSpec{
						Version:  "8.11.0",
						Replicas: 2,
					},
				},
			}

			err := fakeClient.Create(ctx, efkStack)
			Expect(err).NotTo(HaveOccurred())

			// Verify the object was created
			created := &loggingv1.EFKStack{}
			err = fakeClient.Get(ctx, client.ObjectKeyFromObject(efkStack), created)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Spec.Elasticsearch.Version).To(Equal("8.11.0"))
		})
	})
})
