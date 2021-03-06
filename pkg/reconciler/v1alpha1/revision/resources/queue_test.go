/*
Copyright 2018 The Knative Authors

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

package resources

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/knative/pkg/logging"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/autoscaler"
	"github.com/knative/serving/pkg/reconciler/v1alpha1/revision/config"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var boolTrue = true

func TestMakeQueueContainer(t *testing.T) {
	tests := []struct {
		name     string
		rev      *v1alpha1.Revision
		lc       *logging.Config
		ac       *autoscaler.Config
		cc       *config.Controller
		userport *corev1.ContainerPort
		want     *corev1.Container
	}{{
		name: "no owner no autoscaler single",
		rev: &v1alpha1.Revision{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "foo",
				Name:      "bar",
				UID:       "1234",
			},
			Spec: v1alpha1.RevisionSpec{
				ContainerConcurrency: 1,
				TimeoutSeconds:       45,
			},
		},
		lc: &logging.Config{},
		ac: &autoscaler.Config{},
		cc: &config.Controller{},
		userport: &corev1.ContainerPort{
			Name:          userPortEnvName,
			ContainerPort: v1alpha1.DefaultUserPort,
		},
		want: &corev1.Container{
			// These are effectively constant
			Name:           QueueContainerName,
			Resources:      queueResources,
			Ports:          queuePorts,
			Lifecycle:      queueLifecycle,
			ReadinessProbe: queueReadinessProbe,
			// These changed based on the Revision and configs passed in.
			Env: []corev1.EnvVar{{
				Name:  "SERVING_NAMESPACE",
				Value: "foo", // matches namespace
			}, {
				Name: "SERVING_CONFIGURATION",
				// No OwnerReference
			}, {
				Name:  "SERVING_REVISION",
				Value: "bar", // matches name
			}, {
				Name:  "SERVING_AUTOSCALER",
				Value: "autoscaler", // no autoscaler configured.
			}, {
				Name:  "SERVING_AUTOSCALER_PORT",
				Value: "8080",
			}, {
				Name:  "CONTAINER_CONCURRENCY",
				Value: "1",
			}, {
				Name:  "REVISION_TIMEOUT_SECONDS",
				Value: "45",
			}, {
				Name: "SERVING_POD",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
				},
			}, {
				Name: "SERVING_LOGGING_CONFIG",
				// No logging configuration
			}, {
				Name: "SERVING_LOGGING_LEVEL",
				// No logging level
			}, {
				Name:  "USER_PORT",
				Value: strconv.Itoa(v1alpha1.DefaultUserPort),
			}},
		},
	}, {
		name: "no owner no autoscaler single",
		rev: &v1alpha1.Revision{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "foo",
				Name:      "bar",
				UID:       "1234",
			},
			Spec: v1alpha1.RevisionSpec{
				ContainerConcurrency: 1,
				TimeoutSeconds:       45,
			},
		},
		lc: &logging.Config{},
		ac: &autoscaler.Config{},
		cc: &config.Controller{
			QueueSidecarImage: "alpine",
		},
		userport: &corev1.ContainerPort{
			Name:          userPortEnvName,
			ContainerPort: v1alpha1.DefaultUserPort,
		},
		want: &corev1.Container{
			// These are effectively constant
			Name:           QueueContainerName,
			Resources:      queueResources,
			Ports:          queuePorts,
			Lifecycle:      queueLifecycle,
			ReadinessProbe: queueReadinessProbe,
			// These changed based on the Revision and configs passed in.
			Image: "alpine",
			Env: []corev1.EnvVar{{
				Name:  "SERVING_NAMESPACE",
				Value: "foo", // matches namespace
			}, {
				Name: "SERVING_CONFIGURATION",
				// No OwnerReference
			}, {
				Name:  "SERVING_REVISION",
				Value: "bar", // matches name
			}, {
				Name:  "SERVING_AUTOSCALER",
				Value: "autoscaler", // no autoscaler configured.
			}, {
				Name:  "SERVING_AUTOSCALER_PORT",
				Value: "8080",
			}, {
				Name:  "CONTAINER_CONCURRENCY",
				Value: "1",
			}, {
				Name:  "REVISION_TIMEOUT_SECONDS",
				Value: "45",
			}, {
				Name: "SERVING_POD",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
				},
			}, {
				Name: "SERVING_LOGGING_CONFIG",
				// No logging configuration
			}, {
				Name: "SERVING_LOGGING_LEVEL",
				// No logging level
			}, {
				Name:  "USER_PORT",
				Value: strconv.Itoa(v1alpha1.DefaultUserPort),
			}},
		},
	}, {
		name: "config owner as env var, multi-concurrency",
		rev: &v1alpha1.Revision{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "baz",
				Name:      "blah",
				UID:       "1234",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion:         v1alpha1.SchemeGroupVersion.String(),
					Kind:               "Configuration",
					Name:               "the-parent-config-name",
					Controller:         &boolTrue,
					BlockOwnerDeletion: &boolTrue,
				}},
			},
			Spec: v1alpha1.RevisionSpec{
				ContainerConcurrency: 0,
				TimeoutSeconds:       45,
			},
		},
		lc: &logging.Config{},
		ac: &autoscaler.Config{},
		cc: &config.Controller{},
		userport: &corev1.ContainerPort{
			Name:          userPortEnvName,
			ContainerPort: v1alpha1.DefaultUserPort,
		},
		want: &corev1.Container{
			// These are effectively constant
			Name:           QueueContainerName,
			Resources:      queueResources,
			Ports:          queuePorts,
			Lifecycle:      queueLifecycle,
			ReadinessProbe: queueReadinessProbe,
			// These changed based on the Revision and configs passed in.
			Env: []corev1.EnvVar{{
				Name:  "SERVING_NAMESPACE",
				Value: "baz", // matches namespace
			}, {
				Name:  "SERVING_CONFIGURATION",
				Value: "the-parent-config-name",
			}, {
				Name:  "SERVING_REVISION",
				Value: "blah", // matches name
			}, {
				Name:  "SERVING_AUTOSCALER",
				Value: "autoscaler", // no autoscaler configured.
			}, {
				Name:  "SERVING_AUTOSCALER_PORT",
				Value: "8080",
			}, {
				Name:  "CONTAINER_CONCURRENCY",
				Value: "0",
			}, {
				Name:  "REVISION_TIMEOUT_SECONDS",
				Value: "45",
			}, {
				Name: "SERVING_POD",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
				},
			}, {
				Name: "SERVING_LOGGING_CONFIG",
				// No logging configuration
			}, {
				Name: "SERVING_LOGGING_LEVEL",
				// No logging level
			}, {
				Name:  "USER_PORT",
				Value: strconv.Itoa(v1alpha1.DefaultUserPort),
			}},
		},
	}, {
		name: "logging configuration as env var",
		rev: &v1alpha1.Revision{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "log",
				Name:      "this",
				UID:       "1234",
			},
			Spec: v1alpha1.RevisionSpec{
				ContainerConcurrency: 0,
				TimeoutSeconds:       45,
			},
		},
		lc: &logging.Config{
			LoggingConfig: "The logging configuration goes here",
			LoggingLevel: map[string]zapcore.Level{
				"queueproxy": zapcore.ErrorLevel,
			},
		},
		ac: &autoscaler.Config{},
		cc: &config.Controller{},
		userport: &corev1.ContainerPort{
			Name:          userPortEnvName,
			ContainerPort: v1alpha1.DefaultUserPort,
		},
		want: &corev1.Container{
			// These are effectively constant
			Name:           QueueContainerName,
			Resources:      queueResources,
			Ports:          queuePorts,
			Lifecycle:      queueLifecycle,
			ReadinessProbe: queueReadinessProbe,
			// These changed based on the Revision and configs passed in.
			Env: []corev1.EnvVar{{
				Name:  "SERVING_NAMESPACE",
				Value: "log", // matches namespace
			}, {
				Name: "SERVING_CONFIGURATION",
				// No Configuration owner.
			}, {
				Name:  "SERVING_REVISION",
				Value: "this", // matches name
			}, {
				Name:  "SERVING_AUTOSCALER",
				Value: "autoscaler", // no autoscaler configured.
			}, {
				Name:  "SERVING_AUTOSCALER_PORT",
				Value: "8080",
			}, {
				Name:  "CONTAINER_CONCURRENCY",
				Value: "0",
			}, {
				Name:  "REVISION_TIMEOUT_SECONDS",
				Value: "45",
			}, {
				Name: "SERVING_POD",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
				},
			}, {
				Name:  "SERVING_LOGGING_CONFIG",
				Value: "The logging configuration goes here", // from logging config
			}, {
				Name:  "SERVING_LOGGING_LEVEL",
				Value: "error", // from logging config
			}, {
				Name:  "USER_PORT",
				Value: strconv.Itoa(v1alpha1.DefaultUserPort),
			}},
		},
	}, {
		name: "container concurrency 10",
		rev: &v1alpha1.Revision{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "foo",
				Name:      "bar",
				UID:       "1234",
			},
			Spec: v1alpha1.RevisionSpec{
				ContainerConcurrency: 10,
				TimeoutSeconds:       45,
			},
		},
		lc: &logging.Config{},
		ac: &autoscaler.Config{},
		cc: &config.Controller{},
		userport: &corev1.ContainerPort{
			Name:          userPortEnvName,
			ContainerPort: v1alpha1.DefaultUserPort,
		},
		want: &corev1.Container{
			// These are effectively constant
			Name:           QueueContainerName,
			Resources:      queueResources,
			Ports:          queuePorts,
			Lifecycle:      queueLifecycle,
			ReadinessProbe: queueReadinessProbe,
			// These changed based on the Revision and configs passed in.
			Env: []corev1.EnvVar{{
				Name:  "SERVING_NAMESPACE",
				Value: "foo", // matches namespace
			}, {
				Name: "SERVING_CONFIGURATION",
				// No OwnerReference
			}, {
				Name:  "SERVING_REVISION",
				Value: "bar", // matches name
			}, {
				Name:  "SERVING_AUTOSCALER",
				Value: "autoscaler", // no autoscaler configured.
			}, {
				Name:  "SERVING_AUTOSCALER_PORT",
				Value: "8080",
			}, {
				Name:  "CONTAINER_CONCURRENCY",
				Value: "10",
			}, {
				Name:  "REVISION_TIMEOUT_SECONDS",
				Value: "45",
			}, {
				Name: "SERVING_POD",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
				},
			}, {
				Name: "SERVING_LOGGING_CONFIG",
				// No logging configuration
			}, {
				Name: "SERVING_LOGGING_LEVEL",
				// No logging level
			}, {
				Name:  "USER_PORT",
				Value: strconv.Itoa(v1alpha1.DefaultUserPort),
			}},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := makeQueueContainer(test.rev, test.lc, test.ac, test.cc)
			if diff := cmp.Diff(test.want, got, cmpopts.IgnoreUnexported(resource.Quantity{})); diff != "" {
				t.Errorf("makeQueueContainer (-want, +got) = %v", diff)
			}
		})
	}
}
