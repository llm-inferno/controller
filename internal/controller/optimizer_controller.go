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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infernov1beta1 "github.com/llm-inferno/controller/api/v1beta1"
)

// OptimizerReconciler reconciles a Optimizer object
type OptimizerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inferno.platform.ai,resources=optimizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=inferno.platform.ai,resources=optimizers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inferno.platform.ai,resources=optimizers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Optimizer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *OptimizerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// Fetch the object
	optimizer := &infernov1beta1.Optimizer{}
	if err := r.Get(ctx, req.NamespacedName, optimizer); err != nil {
		logf.Log.Info("Error in getting optimizer object, may have been deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := handleFinalizer(ctx, optimizer, r.Update); err != nil {
		logf.Log.Error(err, "failed to update finalizer")
		return ctrl.Result{}, err
	}

	optimizerSpec := &optimizer.Spec
	optimizerName := optimizer.Name

	// Check if the object is being deleted
	if !optimizer.ObjectMeta.DeletionTimestamp.IsZero() {
		// delete the optimizer object
		logf.Log.Info("Optimizer " + optimizerName + " is being deleted!")
		return reconcile.Result{}, nil
	}

	// Handle create/update
	logf.Log.Info("Optimizer " + optimizerName + " created/updated")

	if optimizerSpec.Optimize {
		solution := infernov1beta1.AllocationSolution{}

		if StateLess {
			if systemData, err := r.readSystemData(ctx, req); err != nil {
				logf.Log.Error(err, "failed to read system data, retrying ...")
				return ctrl.Result{RequeueAfter: RetrialDuration}, nil
			} else if err := PostAction(OptimizerURL, OptimizeOne, systemData, &solution); err != nil {
				logf.Log.Error(err, "failed to optimize")
				return ctrl.Result{}, nil
			}
		} else {
			if err := PostAction(OptimizerURL, Optimize, optimizerSpec.Data.Spec, &solution); err != nil {
				logf.Log.Error(err, "failed to optimize")
				return ctrl.Result{}, nil
			}
		}

		logf.Log.Info("Optimizer " + optimizerName + " done")
		if len(solution.Spec) == 0 {
			err := fmt.Errorf("no feasible solution found")
			logf.Log.Error(err, "failed to optimize")
			return ctrl.Result{}, nil
		}

		serverList := &infernov1beta1.ServerList{}
		opts := []client.ListOption{
			client.InNamespace(req.NamespacedName.Namespace),
		}

		if err := r.List(ctx, serverList, opts...); err == nil {
			for _, server := range serverList.Items {
				serverName := server.Spec.Name
				if allocation, exists := solution.Spec[serverName]; exists {
					server.Spec.DesiredAlloc = allocation
					logf.Log.Info("Updating desired allocation for server: " + serverName)
					if err := r.Update(ctx, &server); err != nil {
						logf.Log.Error(err, "failed to update server "+serverName)
					}
				}
			}
		} else {
			logf.Log.Error(err, "failed to get server list")
		}

		// Update status
		logf.Log.Info("Updating status of optimizer " + optimizerName)

		optimizer.Spec.Optimize = false
		if err := r.Update(ctx, optimizer); err != nil {
			logf.Log.Error(err, "failed to update optimizer spec")
			return ctrl.Result{}, err
		}

		optimizer.Status.Done = true
		if err := r.Status().Update(ctx, optimizer); err != nil {
			logf.Log.Error(err, "failed to update optimizer status")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OptimizerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infernov1beta1.Optimizer{}).
		WithEventFilter(updatePredicate()).
		Named("optimizer").
		Complete(r)
}
