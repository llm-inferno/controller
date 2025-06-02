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
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infernov1beta1 "github.com/llm-inferno/controller/api/v1beta1"
)

// ServiceClassReconciler reconciles a ServiceClass object
type ServiceClassReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inferno.platform.ai,resources=serviceclasses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=inferno.platform.ai,resources=serviceclasses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inferno.platform.ai,resources=serviceclasses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceClass object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *ServiceClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// Fetch the object
	svc := &infernov1beta1.ServiceClass{}
	if err := r.Get(ctx, req.NamespacedName, svc); err != nil {
		logf.Log.Info("Error in getting service class object, may have been deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := handleFinalizer(ctx, svc, r.Update); err != nil {
		logf.Log.Error(err, "failed to update finalizer")
		return ctrl.Result{}, err
	}

	svcSpec := &svc.Spec
	svcName := svcSpec.Name
	svcPriority := svcSpec.Priority

	// Check if the object is being deleted
	if !svc.ObjectMeta.DeletionTimestamp.IsZero() {
		// delete the service class
		logf.Log.Info("Service class " + svcName + " is being deleted! Deleting from Optimizer ...")
		// delete service class model target data
		for _, data := range svcSpec.Data {
			modelName := data.Model
			args := "/" + svcName + "/" + modelName
			if _, err := GetAction(OptimizerURL, RemoveServiceClassModelTarget, args); err != nil {
				logf.Log.Error(err, "failed to delete service class model target data, retrying ..."+args)
			}
		}
		// remove service class
		if _, err := GetAction(OptimizerURL, RemoveServiceClass, "/"+svcName); err != nil {
			logf.Log.Error(err, "failed to delete service class "+svcName)
		}
		return reconcile.Result{}, nil
	}

	// Handle create/update
	logf.Log.Info("Service class " + svcName + " created/updated; adding to optimizer")
	// add service class
	args := "/" + svcName + "/" + strconv.Itoa(svcPriority)
	if _, err := GetAction(OptimizerURL, AddServiceClass, args); err != nil {
		logf.Log.Error(err, "failed to add service class, retrying ..."+svcName)
		return ctrl.Result{RequeueAfter: RetrialDuration}, nil
	}
	// add model target data
	for _, data := range svcSpec.Data {
		targetData := &infernov1beta1.ServiceClassDataItem{
			Name:                  svcName,
			Priority:              svcPriority,
			ServiceClassModelData: data,
		}
		if err := PostAction(OptimizerURL, AddServiceClassModelTarget, targetData, nil); err != nil {
			logf.Log.Error(err, "failed to add service class model target data, retrying ..."+svcName)
			return ctrl.Result{RequeueAfter: RetrialDuration}, nil
		}
	}
	logf.Log.Info("Service class " + svcName + " successfully processed")

	// Update status
	logf.Log.Info("Updating status of service class " + svcName)
	svc.Status.Active = true
	if err := r.Status().Update(ctx, svc); err != nil {
		logf.Log.Error(err, "failed to update service class status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infernov1beta1.ServiceClass{}).
		WithEventFilter(updatePredicate()).
		Named("serviceclass").
		Complete(r)
}
