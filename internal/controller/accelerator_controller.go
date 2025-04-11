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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infernov1beta1 "github.ibm.com/inferno/controller/api/v1beta1"
)

// AcceleratorReconciler reconciles a Accelerator object
type AcceleratorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inferno.platform.ai,resources=accelerators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=inferno.platform.ai,resources=accelerators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inferno.platform.ai,resources=accelerators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Accelerator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *AcceleratorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// Fetch the object
	accelerator := &infernov1beta1.Accelerator{}
	if err := r.Get(ctx, req.NamespacedName, accelerator); err != nil {
		logf.Log.Info("Error in getting accelerator object, may have been deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := handleFinalizer(ctx, accelerator, r.Update); err != nil {
		logf.Log.Error(err, "failed to update finalizer")
		return ctrl.Result{}, err
	}

	acceleratorSpec := &accelerator.Spec
	acceleratorName := acceleratorSpec.Name

	// Check if the object is being deleted
	if !accelerator.ObjectMeta.DeletionTimestamp.IsZero() {
		// delete the accelerator
		logf.Log.Info("Accelerator " + acceleratorName + " is being deleted! Deleting from Optimizer ...")
		if _, err := GetAction(OptimizerURL, RemoveAccelerator, "/"+acceleratorName); err != nil {
			logf.Log.Error(err, "failed to delete accelerator "+acceleratorName)
		}
		return reconcile.Result{}, nil
	}

	// Handle create/update
	logf.Log.Info("Accelerator " + acceleratorName + " created/updated; adding to optimizer")
	if err := PostAction(OptimizerURL, AddAccelerator, acceleratorSpec, nil); err != nil {
		logf.Log.Error(err, "failed to add accelerator, retrying ..."+acceleratorName)
		return ctrl.Result{RequeueAfter: RetrialDuration}, nil
	}
	logf.Log.Info("Accelerator " + acceleratorName + " successfully processed")

	// Update status
	logf.Log.Info("Updating status of accelerator " + acceleratorName)
	accelerator.Status.Active = true
	if err := r.Status().Update(ctx, accelerator); err != nil {
		logf.Log.Error(err, "failed to update accelerator status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AcceleratorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infernov1beta1.Accelerator{}).
		WithEventFilter(updatePredicate()).
		Named("accelerator").
		Complete(r)
}
