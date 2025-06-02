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

	infernov1alpha1 "github.com/llm-inferno/controller/api/v1alpha1"
)

// ModelReconciler reconciles a Model object
type ModelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inferno.platform.ai,resources=models,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=inferno.platform.ai,resources=models/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inferno.platform.ai,resources=models/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Model object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *ModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// Fetch the object
	model := &infernov1alpha1.Model{}
	if err := r.Get(ctx, req.NamespacedName, model); err != nil {
		logf.Log.Info("Error in getting model object, may have been deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := handleFinalizer(ctx, model, r.Update); err != nil {
		logf.Log.Error(err, "failed to update finalizer")
		return ctrl.Result{}, err
	}

	modelSpec := &model.Spec
	modelName := modelSpec.Name

	// Check if the object is being deleted
	if !model.ObjectMeta.DeletionTimestamp.IsZero() {
		// delete the model
		logf.Log.Info("Model " + modelName + " is being deleted! Deleting from Optimizer ...")
		// delete accelerator perf data
		for _, data := range modelSpec.Data {
			accName := data.Acc
			args := "/" + modelName + "/" + accName
			if _, err := GetAction(OptimizerURL, RemoveModelAcceleratorPerf, args); err != nil {
				logf.Log.Error(err, "failed to delete model accelerator perf data, retrying ..."+args)
			}
		}
		// remove model
		if _, err := GetAction(OptimizerURL, RemoveModel, "/"+modelName); err != nil {
			logf.Log.Error(err, "failed to delete model "+modelName)
		}
		return reconcile.Result{}, nil
	}

	// Handle create/update
	logf.Log.Info("Model " + modelName + " created/updated; adding to optimizer")
	// add model
	if _, err := GetAction(OptimizerURL, AddModel, "/"+modelName); err != nil {
		logf.Log.Error(err, "failed to add model, retrying ..."+modelName)
		return ctrl.Result{RequeueAfter: RetrialDuration}, nil
	}
	// add accelerator perf data
	for _, data := range modelSpec.Data {
		perfData := &infernov1alpha1.ModelAcceleratorPerfData{
			Name:                modelName,
			AcceleratorPerfData: data,
		}
		if err := PostAction(OptimizerURL, AddModelAcceleratorPerf, perfData, nil); err != nil {
			logf.Log.Error(err, "failed to add model accelerator perf data, retrying ..."+modelName)
			return ctrl.Result{RequeueAfter: RetrialDuration}, nil
		}
	}
	logf.Log.Info("Model " + modelName + " successfully processed")

	// Update status
	logf.Log.Info("Updating status of model " + modelName)
	model.Status.Active = true
	if err := r.Status().Update(ctx, model); err != nil {
		logf.Log.Error(err, "failed to update model status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infernov1alpha1.Model{}).
		WithEventFilter(updatePredicate()).
		Named("model").
		Complete(r)
}
