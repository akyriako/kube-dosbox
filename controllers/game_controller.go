/*
Copyright 2023.

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
	operatorv1alpha1 "github.com/akyriako/kube-dosbox/api/v1alpha1"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	logger logr.Logger
)

var (
	gameEventFilters = builder.WithPredicates(predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// We only need to check generation changes here, because it is only
			// updated on spec changes. On the other hand RevisionVersion
			// changes also on status changes. We want to omit reconciliation
			// for status updates.
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// DeleteStateUnknown evaluates to false only if the object
			// has been confirmed as deleted by the api server.
			return !e.DeleteStateUnknown
		},
		CreateFunc: func(e event.CreateEvent) bool {
			switch object := e.Object.(type) {
			case *operatorv1alpha1.Game:
				return object.Spec.Deploy
			default:
				return false
			}
		},
	})
)

// GameReconciler reconciles a Game object
type GameReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operator.contrib.dosbox.com,resources=games,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.contrib.dosbox.com,resources=games/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.contrib.dosbox.com,resources=games/finalizers,verbs=update
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps;persistentvolumeclaims;services;pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Game object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *GameReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger = log.FromContext(ctx).WithName("controller")

	game := &operatorv1alpha1.Game{}
	if err := r.Get(ctx, req.NamespacedName, game); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		logger.V(5).Error(err, "unable to fetch game")
		return ctrl.Result{}, err
	}

	//if game.Status.Ready == nil {
	//	//_ = r.SetStatus(ctx, req, game, false)
	//}

	if !game.Spec.Deploy {
		err := r.DeleteDeployment(ctx, req, game)
		if err != nil {
			return ctrl.Result{}, err
		}

		_ = r.SetStatus(ctx, req, game, false)

		return ctrl.Result{}, nil
	}

	_, err := r.CreateOrUpdatePersistentVolumeClaimAssets(ctx, req, game)
	if err != nil {
		return ctrl.Result{}, err
	}

	deployment, err := r.CreateOrUpdateDeployment(ctx, req, game)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = r.CreateOrUpdateConfigMap(ctx, req, game, deployment)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = r.CreateOrUpdatePersistentVolumeClaim(ctx, req, game, deployment)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = r.CreateOrUpdateService(ctx, req, game, deployment)
	if err != nil {
		return ctrl.Result{}, err
	}

	return r.RefreshStatus(ctx, req, game, deployment.Labels["app"])
}

// SetupWithManager sets up the controller with the Manager.
func (r *GameReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Game{}, gameEventFilters).
		Complete(r)
}
