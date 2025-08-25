package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	apiv1alpha1 "github.com/llm-inferno/api/api/v1alpha1"
	infernov1alpha1 "github.com/llm-inferno/controller/api/v1alpha1"
)

// get URL of a REST server
func GetURL(hostEnvName, portEnvName string) string {
	host := "localhost"
	port := "8080"
	if h := os.Getenv(hostEnvName); h != "" {
		host = h
	}
	if p := os.Getenv(portEnvName); p != "" {
		port = p
	}
	return "http://" + host + ":" + port
}

// predicate to filter reconciliation events
func updatePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}

type updateOp func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error

// add or delete finalizer depending on object creation or deletion
func handleFinalizer(ctx context.Context, obj client.Object, op updateOp) error {
	name := FinalizerName
	if obj.GetDeletionTimestamp().IsZero() {
		// add finalizer in case of create/update
		if !controllerutil.ContainsFinalizer(obj, name) {
			ok := controllerutil.AddFinalizer(obj, name)
			logf.Log.Info("Add Finalizer", name, ok)
			return op(ctx, obj)
		}
	} else {
		// remove finalizer in case of deletion
		if controllerutil.ContainsFinalizer(obj, name) {
			ok := controllerutil.RemoveFinalizer(obj, name)
			logf.Log.Info("Remove Finalizer", name, ok)
			return op(ctx, obj)
		}
	}
	return nil
}

// send GET to optimizer REST API server
func GetAction(url string, verb string, args string) ([]byte, error) {
	endPoint := url + verb
	cmd := endPoint + args
	response, getErr := http.Get(cmd)
	if getErr != nil {
		return nil, getErr
	}
	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}
	return body, nil
}

// send POST to optimizer REST API server
func PostAction(url string, verb string, specIn any, specOut any) (err error) {
	var byteValue []byte
	var req *http.Request
	var res *http.Response
	if byteValue, err = json.Marshal(specIn); err == nil {
		endPoint := url + verb
		if req, err = http.NewRequest("POST", endPoint, bytes.NewBuffer(byteValue)); err == nil {
			req.Header.Add("Content-Type", "application/json")
			client := &http.Client{}
			if res, err = client.Do(req); err == nil {

				defer func() {
					if closeErr := res.Body.Close(); closeErr != nil && err == nil {
						err = closeErr
					}
				}()

				if res.StatusCode == http.StatusOK {
					if specOut != nil {
						err = json.NewDecoder(res.Body).Decode(specOut)
					}
				} else {
					err = fmt.Errorf("%s", res.Status)
				}
			}
		}
	}
	return err
}

// read all system data from specifications of resources
func (r *OptimizerReconciler) readSystemData(ctx context.Context,
	req ctrl.Request) (systemData *apiv1alpha1.SystemData, err error) {

	systemSpec := apiv1alpha1.SystemSpec{
		Accelerators: apiv1alpha1.AcceleratorData{
			Spec: make([]apiv1alpha1.AcceleratorSpec, 0),
		},
		Models: apiv1alpha1.ModelData{
			PerfData: make([]apiv1alpha1.ModelAcceleratorPerfData, 0),
		},
		ServiceClasses: apiv1alpha1.ServiceClassData{
			Spec: make([]apiv1alpha1.ServiceClassSpec, 0),
		},
		Servers: apiv1alpha1.ServerData{
			Spec: make([]apiv1alpha1.ServerSpec, 0),
		},
		Optimizer: apiv1alpha1.OptimizerData{
			Spec: apiv1alpha1.OptimizerSpec{},
		},
		Capacity: apiv1alpha1.CapacityData{
			Count: make([]apiv1alpha1.AcceleratorCount, 0),
		},
	}

	opts := []client.ListOption{
		client.InNamespace(req.NamespacedName.Namespace),
	}

	// get optimizer data
	optimizerList := &infernov1alpha1.OptimizerList{}
	if err = r.List(ctx, optimizerList, opts...); err != nil {
		logf.Log.Error(err, "failed to get optimizer list")
		return nil, err
	}
	if len(optimizerList.Items) == 0 {
		err = fmt.Errorf("no optimizer specs found")
		logf.Log.Error(err, "failed to get optimizer")
		return nil, err
	}
	optimizer := optimizerList.Items[0] // only one is needed
	systemSpec.Optimizer.Spec = optimizer.Spec.Data.Spec

	// get accelerator data
	acceleratorList := &infernov1alpha1.AcceleratorList{}
	if err = r.List(ctx, acceleratorList, opts...); err != nil {
		logf.Log.Error(err, "failed to get accelerator list")
		return nil, err
	}
	for _, accelerator := range acceleratorList.Items {
		systemSpec.Accelerators.Spec = append(systemSpec.Accelerators.Spec,
			apiv1alpha1.AcceleratorSpec(accelerator.Spec))
	}

	// get model data
	modelList := &infernov1alpha1.ModelList{}
	if err = r.List(ctx, modelList, opts...); err != nil {
		logf.Log.Error(err, "failed to get model list")
		return nil, err
	}
	for _, model := range modelList.Items {
		for _, data := range model.Spec.Data {
			perfData := &apiv1alpha1.ModelAcceleratorPerfData{
				Name:         model.Spec.Name,
				Acc:          data.Acc,
				AccCount:     data.AccCount,
				Alpha:        data.Alpha,
				Beta:         data.Beta,
				MaxBatchSize: data.MaxBatchSize,
				AtTokens:     data.AtTokens,
			}
			systemSpec.Models.PerfData = append(systemSpec.Models.PerfData,
				*perfData)
		}
	}

	// get service class data
	svcList := &infernov1alpha1.ServiceClassList{}
	if err = r.List(ctx, svcList, opts...); err != nil {
		logf.Log.Error(err, "failed to get service class list")
		return nil, err
	}
	for _, svc := range svcList.Items {
		svcSpec := &apiv1alpha1.ServiceClassSpec{
			Name:         svc.Spec.Name,
			Priority:     svc.Spec.Priority,
			ModelTargets: make([]apiv1alpha1.ModelTarget, len(svc.Spec.Data)),
		}
		for i, data := range svc.Spec.Data {
			svcSpec.ModelTargets[i] = apiv1alpha1.ModelTarget{
				Model:   data.Model,
				SLO_ITL: data.SLO_ITL,
				SLO_TTW: data.SLO_TTW,
				SLO_TPS: data.SLO_TPS,
			}
		}
		systemSpec.ServiceClasses.Spec = append(systemSpec.ServiceClasses.Spec,
			*svcSpec)
	}

	// get server data
	serverList := &infernov1alpha1.ServerList{}
	if err = r.List(ctx, serverList, opts...); err != nil {
		logf.Log.Error(err, "failed to get server list")
		return nil, err
	}
	for _, server := range serverList.Items {
		systemSpec.Servers.Spec = append(systemSpec.Servers.Spec,
			apiv1alpha1.ServerSpec(server.Spec))
	}

	// get capacity data
	capacityList := &infernov1alpha1.CapacityList{}
	if err := r.List(ctx, capacityList, opts...); err != nil {
		logf.Log.Error(err, "failed to get capacity list")
		return nil, err
	}
	for _, capacity := range capacityList.Items {
		systemSpec.Capacity.Count = append(systemSpec.Capacity.Count,
			capacity.Spec.Count...)
	}

	systemData = &apiv1alpha1.SystemData{
		Spec: systemSpec,
	}
	return systemData, nil
}

// remove a string item from a slice of strings
func RemoveFromSlice(slice []string, item string) (out []string) {
	out = make([]string, 0)
	for _, s := range slice {
		if s != item {
			out = append(out, s)
		}
	}
	return out
}
