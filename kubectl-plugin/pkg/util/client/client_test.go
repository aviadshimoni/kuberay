package client

import (
	"context"
	"testing"

	"github.com/ray-project/kuberay/kubectl-plugin/pkg/util"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeFake "k8s.io/client-go/kubernetes/fake"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	rayClientFake "github.com/ray-project/kuberay/ray-operator/pkg/client/clientset/versioned/fake"
)

func TestGetKubeRayOperatorVersion(t *testing.T) {
	helmKubeObjects := []runtime.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kuberay-operator-helm-chart",
				Namespace: "default",
				Labels: map[string]string{
					"app.kubernetes.io/name": "kuberay-operator",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "kuberay/operator:v0.5.0",
							},
						},
					},
				},
			},
		},
	}
	kustomizeObjects := []runtime.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kuberay-operator-kustomize",
				Namespace: "test",
				Labels: map[string]string{
					"app.kubernetes.io/name": "kuberay",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "kuberay/operator:v0.6.0",
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name            string
		expectedVersion string
		expectedError   string
		kubeObjects     []runtime.Object
	}{
		{
			name:            "kubeRay operator not found",
			expectedVersion: "",
			expectedError:   "no KubeRay operator deployments found in any namespace",
			kubeObjects:     nil,
		},
		{
			name:            "find kubeRay operator version for helm chart",
			expectedVersion: "v0.5.0",
			expectedError:   "",
			kubeObjects:     helmKubeObjects,
		},
		{
			name:            "find kubeRay operator version for Kustomize",
			expectedVersion: "v0.6.0",
			expectedError:   "",
			kubeObjects:     kustomizeObjects,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kubeClientSet := kubeFake.NewClientset(tc.kubeObjects...)
			client := NewClientForTesting(kubeClientSet, nil)

			version, err := client.GetKubeRayOperatorVersion(context.Background())

			if tc.expectedVersion != "" {
				assert.Equal(t, version, tc.expectedVersion)
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestGetRayHeadSvcNameByRayCluster(t *testing.T) {
	kubeObjects := []runtime.Object{}

	rayObjects := []runtime.Object{
		&rayv1.RayCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "raycluster-default",
				Namespace: "default",
			},
			Status: rayv1.RayClusterStatus{
				Head: rayv1.HeadInfo{
					ServiceName: "raycluster-default-head-svc",
				},
			},
		},
		&rayv1.RayCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "raycluster-test",
				Namespace: "test",
			},
			Status: rayv1.RayClusterStatus{
				Head: rayv1.HeadInfo{
					ServiceName: "raycluster-test-head-svc",
				},
			},
		},
	}

	kubeClientSet := kubeFake.NewClientset(kubeObjects...)
	rayClient := rayClientFake.NewSimpleClientset(rayObjects...)
	client := NewClientForTesting(kubeClientSet, rayClient)

	tests := []struct {
		name         string
		namespace    string
		resourceName string
		serviceName  string
	}{
		{
			name:         "find service name in default namespace",
			namespace:    "default",
			resourceName: "raycluster-default",
			serviceName:  "raycluster-default-head-svc",
		},
		{
			name:         "find service name in test namespace",
			namespace:    "test",
			resourceName: "raycluster-test",
			serviceName:  "raycluster-test-head-svc",
		},
		{
			name:         "resource not found",
			namespace:    "default",
			resourceName: "raycluster-not-found",
			serviceName:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svcName, err := client.GetRayHeadSvcName(context.Background(), tc.namespace, util.RayCluster, tc.resourceName)
			if tc.serviceName == "" {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.serviceName, svcName)
			}
		})
	}
}

func TestGetRayHeadSvcNameByRayJob(t *testing.T) {
	kubeObjects := []runtime.Object{}

	rayObjects := []runtime.Object{
		&rayv1.RayJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rayjob-default",
				Namespace: "default",
			},
			Status: rayv1.RayJobStatus{
				RayClusterStatus: rayv1.RayClusterStatus{
					Head: rayv1.HeadInfo{
						ServiceName: "rayjob-default-raycluster-xxxxx-head-svc",
					},
				},
			},
		},
		&rayv1.RayJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rayjob-test",
				Namespace: "test",
			},
			Status: rayv1.RayJobStatus{
				RayClusterStatus: rayv1.RayClusterStatus{
					Head: rayv1.HeadInfo{
						ServiceName: "rayjob-test-raycluster-xxxxx-head-svc",
					},
				},
			},
		},
	}

	kubeClientSet := kubeFake.NewClientset(kubeObjects...)
	rayClient := rayClientFake.NewSimpleClientset(rayObjects...)
	client := NewClientForTesting(kubeClientSet, rayClient)

	tests := []struct {
		name         string
		namespace    string
		resourceName string
		serviceName  string
	}{
		{
			name:         "find service name in default namespace",
			namespace:    "default",
			resourceName: "rayjob-default",
			serviceName:  "rayjob-default-raycluster-xxxxx-head-svc",
		},
		{
			name:         "find service name in test namespace",
			namespace:    "test",
			resourceName: "rayjob-test",
			serviceName:  "rayjob-test-raycluster-xxxxx-head-svc",
		},
		{
			name:         "resource not found",
			namespace:    "default",
			resourceName: "rayjob-not-found",
			serviceName:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svcName, err := client.GetRayHeadSvcName(context.Background(), tc.namespace, util.RayJob, tc.resourceName)
			if tc.serviceName == "" {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.serviceName, svcName)
			}
		})
	}
}

func TestGetRayHeadSvcNameByRayService(t *testing.T) {
	kubeObjects := []runtime.Object{}

	rayObjects := []runtime.Object{
		&rayv1.RayService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rayservice-default",
				Namespace: "default",
			},
			Status: rayv1.RayServiceStatuses{
				ActiveServiceStatus: rayv1.RayServiceStatus{
					RayClusterStatus: rayv1.RayClusterStatus{
						Head: rayv1.HeadInfo{
							ServiceName: "rayservice-default-raycluster-xxxxx-head-svc",
						},
					},
				},
			},
		},
		&rayv1.RayService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rayservice-test",
				Namespace: "test",
			},
			Status: rayv1.RayServiceStatuses{
				ActiveServiceStatus: rayv1.RayServiceStatus{
					RayClusterStatus: rayv1.RayClusterStatus{
						Head: rayv1.HeadInfo{
							ServiceName: "rayservice-test-raycluster-xxxxx-head-svc",
						},
					},
				},
			},
		},
	}

	kubeClientSet := kubeFake.NewClientset(kubeObjects...)
	rayClient := rayClientFake.NewSimpleClientset(rayObjects...)
	client := NewClientForTesting(kubeClientSet, rayClient)

	tests := []struct {
		name         string
		namespace    string
		resourceName string
		serviceName  string
	}{
		{
			name:         "find service name in default namespace",
			namespace:    "default",
			resourceName: "rayservice-default",
			serviceName:  "rayservice-default-raycluster-xxxxx-head-svc",
		},
		{
			name:         "find service name in test namespace",
			namespace:    "test",
			resourceName: "rayservice-test",
			serviceName:  "rayservice-test-raycluster-xxxxx-head-svc",
		},
		{
			name:         "resource not found",
			namespace:    "default",
			resourceName: "rayservice-not-found",
			serviceName:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svcName, err := client.GetRayHeadSvcName(context.Background(), tc.namespace, util.RayService, tc.resourceName)
			if tc.serviceName == "" {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.serviceName, svcName)
			}
		})
	}
}
