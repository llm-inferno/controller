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
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infernov1beta1 "github.ibm.com/inferno/controller/api/v1beta1"
)

// ServerReconciler reconciles a Server object
type ServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inferno.platform.ai,resources=servers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=inferno.platform.ai,resources=servers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inferno.platform.ai,resources=servers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Server object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *ServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// Fetch the object
	server := &infernov1beta1.Server{}
	if err := r.Get(ctx, req.NamespacedName, server); err != nil {
		logf.Log.Info("Error in getting server object, may have been deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := handleFinalizer(ctx, server, r.Update); err != nil {
		logf.Log.Error(err, "failed to update finalizer")
		return ctrl.Result{}, err
	}

	serverSpec := &server.Spec
	serverName := serverSpec.Name

	// Check if the object is being deleted
	if !server.ObjectMeta.DeletionTimestamp.IsZero() {
		// delete the server
		logf.Log.Info("Server " + serverName + " is being deleted! Deleting from Optimizer ...")
		endPoint := OptimizerURL + RemoveServer
		cmd := endPoint + "/" + serverName
		if _, err := http.Get(cmd); err != nil {
			logf.Log.Error(err, "failed to delete server "+serverName)
		}
		return reconcile.Result{}, nil
	}

	// Handle create/update
	logf.Log.Info("Server " + serverName + " created/updated; adding to optimizer")
	if err := PostAction(OptimizerURL, AddServer, serverSpec, nil); err != nil {
		logf.Log.Error(err, "failed to add server, retrying ..."+serverName)
		return ctrl.Result{RequeueAfter: RetrialDuration}, nil
	}
	logf.Log.Info("Server " + serverName + " successfully processed")

	// Update status
	logf.Log.Info("Updating status of server " + serverName)
	server.Status.Active = true
	if err := r.Status().Update(ctx, server); err != nil {
		logf.Log.Error(err, "failed to update server status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infernov1beta1.Server{}).
		WithEventFilter(updatePredicate()).
		Named("server").
		Complete(r)
}
