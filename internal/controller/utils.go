package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
