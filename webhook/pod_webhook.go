/*
Copyright 2021.

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

package v1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

//+kubebuilder:webhook:path=/mutate--v1-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups=core,resources=pods,verbs=create;update,versions=v1,name=mutate.pod.hitsumabushi845.github.io,admissionReviewVersions=v1

var (
	AnnotationKey string = "hitsumabushi845.github.io/sample-annotation"
)

type PodWebhook struct{}

var _ admission.CustomDefaulter = &PodWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (*PodWebhook) Default(ctx context.Context, obj runtime.Object) error {

	pod := obj.(*corev1.Pod)

	if pod.Annotations == nil {
		return nil
	}

	annotationValue, found := pod.Annotations[AnnotationKey]
	if !found {
		podlog.Info("Annotation not found. Skip defaulting.")
		return nil
	} else {
		podlog.Info("Annotation found:", "value", annotationValue)
	}

	if len(pod.Spec.Containers) > 1 {
		podlog.Info("Multiple containers found. Skip defaulting.")
		return nil
	}

	pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, corev1.EnvVar{Name: "SAMPLE_ENV", Value: annotationValue})

	return nil
}

//+kubebuilder:webhook:path=/validate--v1-pod,mutating=false,failurePolicy=fail,sideEffects=None,groups=core,resources=pods,verbs=create;update,versions=v1,name=validate.pod.hitsumabushi845.github.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &PodWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *PodWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	return r.ValidateAnnotation(obj.(*corev1.Pod))
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *PodWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	return r.ValidateAnnotation(newObj.(*corev1.Pod))
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (*PodWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}

func (*PodWebhook) ValidateAnnotation(obj *corev1.Pod) error {
	var errs field.ErrorList
	_, found := obj.Annotations[AnnotationKey]

	if !found {
		errs = append(errs, field.Required(field.NewPath("annotations"), "Annotation hitsumabushi845.github.io/sample-annotation must be defined."))
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: "core.v1", Kind: "Pod"}, obj.Name, errs)
		podlog.Error(err, "validation error", "name", obj.Name)
		return err
	}
	return nil
}
