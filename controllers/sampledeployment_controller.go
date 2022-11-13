/*
Copyright 2022.

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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	samplev1 "github.com/gosarami/sample-deployment/api/v1"
)

// SampleDeploymentReconciler reconciles a SampleDeployment object
type SampleDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sample.gosarami.github.io,resources=sampledeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sample.gosarami.github.io,resources=sampledeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sample.gosarami.github.io,resources=sampledeployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the sample closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SampleDeployment object against the actual sample state, and then
// perform operations to make the sample state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SampleDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = log.FromContext(ctx)
	logger := log.FromContext(ctx)

	var sampleDeployment samplev1.SampleDeployment

	err := r.Get(ctx, req.NamespacedName, &sampleDeployment)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "unable to get SampleDeployment", "name", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if err := r.reconcileSample(ctx, sampleDeployment); err != nil {
		logger.Error(err, "unable to reconsile Cluster", "name", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if !sampleDeployment.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	return r.updateStatus(ctx, sampleDeployment)
}

func (r *SampleDeploymentReconciler) reconcileSample(ctx context.Context, sampleDeployment samplev1.SampleDeployment) error {
	logger := log.FromContext(ctx)

	sample := &samplev1.Sample{}
	sample.SetNamespace(sampleDeployment.Namespace)
	sample.SetName("sample-1")

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, sample, func() error {
		if sample.Spec.Version == "" {
			sample.Spec.Version = sampleDeployment.Spec.Version
		}
		return ctrl.SetControllerReference(&sampleDeployment, sample, r.Scheme)
	})

	if err != nil {
		logger.Error(err, "unable to create or update Sample")
		return err
	}
	if op != controllerutil.OperationResultNone {
		logger.Info("reconcile Sample successfully", "op", op)
	}
	return nil
}

func (r *SampleDeploymentReconciler) updateStatus(ctx context.Context, sampleDeployment samplev1.SampleDeployment) (ctrl.Result, error) {
	var sample samplev1.Sample
	if err := r.Get(
		ctx,
		client.ObjectKey{
			Namespace: sampleDeployment.Namespace,
			//			Name:      getSampleName(sampleDeployment),
		},
		&sample,
	); err != nil {
		return ctrl.Result{}, err
	}

	/*
		var status samplev1.SampleDeploymentStatus
		status.Sample.Status = "TBD"

		if sampleDeployment.Status != status {
			sampleDeployment.Status = status
			if err := r.Status().Update(ctx, &sampleDeployment); err != nil {
				return ctrl.Result{}, err
			}
		}
	*/

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SampleDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&samplev1.SampleDeployment{}).
		Owns(&samplev1.Sample{}).
		Complete(r)
}
