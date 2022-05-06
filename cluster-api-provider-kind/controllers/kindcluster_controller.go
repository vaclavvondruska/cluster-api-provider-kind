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
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	infrastructurev1alpha1 "cluster-api-provider-kind/api/v1alpha1"
)

// KindClusterReconciler reconciles a KindCluster object
type KindClusterReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	KindClient *KindClient
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *KindClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	clusterName := req.NamespacedName.String()

	// Retrieve the cluster
	var kindCluster infrastructurev1alpha1.KindCluster
	if err := r.Get(ctx, req.NamespacedName, &kindCluster); err != nil {
		if errors.IsNotFound(err) {
			err = r.KindClient.DeleteCluster(req.Namespace, req.Name)
			return ctrl.Result{}, err
		}
		logger.Error(err, fmt.Sprintf("Failed to retrieve cluster %s", clusterName))
		return ctrl.Result{}, err
	}

	// Retrieve the owner cluster
	ownerCluster, err := util.GetOwnerCluster(ctx, r.Client, kindCluster.ObjectMeta)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to retrieve owner cluster %s", clusterName))
		return ctrl.Result{}, err
	}

	// Get Patch helper to update the cluster
	helper, err := patch.NewHelper(&kindCluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Retrieve observed cluster state
	observedStatus, err := r.KindClient.GetClusterStatus(req.Namespace, req.Name)
	clusterNotFound := err == KindClusterNotFoundError
	if err != nil && !clusterNotFound {
		return ctrl.Result{}, err
	}

	// If the resource is being deleted, handle the finalizer
	if kindCluster.IsBeingDeleted() {
		if kindCluster.HasFinalizer(infrastructurev1alpha1.KindClusterFinalizerName) {
			// Make sure the Kind cluster is deleted
			if !clusterNotFound {
				err := r.KindClient.DeleteCluster(kindCluster.Namespace, kindCluster.Name)
				if err != nil {
					logger.Error(err, fmt.Sprintf("Failed to delete Kind cluster %s", clusterName))
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: time.Second}, nil
			}
			// Delete the owner cluster, unless it is already being deleted
			if ownerCluster != nil && ownerCluster.ObjectMeta.DeletionTimestamp.IsZero() {
				err := r.Client.Delete(ctx, ownerCluster)
				if err != nil {
					logger.Error(err, fmt.Sprintf("Failed to handle finalizer in cluster %s", clusterName))
					return ctrl.Result{}, err
				}
			}
			kindCluster.RemoveFinalizer(infrastructurev1alpha1.KindClusterFinalizerName)
		}
		err = helper.Patch(ctx, &kindCluster)
		return ctrl.Result{}, err
	}

	// Add finalizer if needed
	if !kindCluster.HasFinalizer(infrastructurev1alpha1.KindClusterFinalizerName) {
		kindCluster.AddFinalizer(infrastructurev1alpha1.KindClusterFinalizerName)
	}

	// In case the cluster is in the Failed state, stop the reconciliation
	if kindCluster.Status.State == infrastructurev1alpha1.KindClusterStateFailed {
		return ctrl.Result{}, nil
	}

	// Update cluster state
	result := ctrl.Result{}
	if observedStatus.State == KindStateRunning {
		if !kindCluster.HasControlPlaneEndpoint() {
			kindCluster.AddControlPlaneEndpoint(observedStatus.Host, observedStatus.Port)
		}
		kindCluster.Status.State = infrastructurev1alpha1.KindClusterStateRunning
		kindCluster.Status.Ready = true
	} else if observedStatus.State == KindStateFailed {
		kindCluster.Status.State = infrastructurev1alpha1.KindClusterStateFailed
		kindCluster.Status.Ready = false
	} else {
		if clusterNotFound && observedStatus.State != KindStatePending {
			err = r.KindClient.CreateCluster(req.Namespace, req.Name, kindCluster.Spec)
			if err != nil {
				logger.Error(err, fmt.Sprintf("Failed to start cluster %s", clusterName))
				return ctrl.Result{}, err
			}
		}
		kindCluster.Status.State = infrastructurev1alpha1.KindClusterStatePending
		kindCluster.Status.Ready = false
		result.RequeueAfter = 5 * time.Second
	}

	err = helper.Patch(ctx, &kindCluster)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to update cluster %s", clusterName))
		return ctrl.Result{}, err
	}

	return result, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *KindClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1alpha1.KindCluster{}).
		Complete(r)
}
