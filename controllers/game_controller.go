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
	"github.com/akyriako/kube-dosbox/assets"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	logger logr.Logger
)

// GameReconciler reconciles a Game object
type GameReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operator.contrib.dosbox.com,resources=games,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.contrib.dosbox.com,resources=games/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.contrib.dosbox.com,resources=games/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pvc,verbs=get;list;watch;create;update;patch;delete

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

	var exists = true
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			exists = false
		} else {
			logger.V(5).Error(err, "unable to fetch deployment")
			return ctrl.Result{}, err
		}
	}

	if (exists && game.Spec.ForceRedeploy) || !exists {
		return r.create(ctx, game)
	} else {
		return r.update(ctx, game)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GameReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Game{}).
		WithEventFilter(setupEventFilterPredicates()).
		Complete(r)
}

func (r *GameReconciler) create(ctx context.Context, game *operatorv1alpha1.Game) (ctrl.Result, error) {
	//index, err := assets.GetIndex(filepath.Base(game.Spec.Url))
	//if err != nil {
	//	return ctrl.Result{}, err
	//}

	deployment, err := assets.GetDeployment(game.Namespace, game.Name, game.Spec.Port)
	if err != nil {
		logger.Error(err, "unable to parse deployment template")
		return ctrl.Result{}, err
	}

	cmap, err := assets.GetConfigMap(game.Namespace, game.Name, filepath.Base(game.Spec.Url))
	if err != nil {
		logger.Error(err, "unable to parse configmap template")
		return ctrl.Result{}, err
	}

	pvc, err := assets.GetPersistentVolumeClaim(game.Namespace, game.Name)
	if err != nil {
		logger.Error(err, "unable to parse pvc template")
		return ctrl.Result{}, err
	}

	svc, err := assets.GetService(game.Namespace, game.Name, game.Spec.Port)
	if err != nil {
		logger.Error(err, "unable to parse svc template")
		return ctrl.Result{}, err
	}

	err = ctrl.SetControllerReference(game, deployment, r.Scheme)
	if err != nil {
		logger.Error(err, "unable to set controller reference")
		return ctrl.Result{}, err
	}

	err = r.Create(ctx, deployment)
	if err != nil {
		logger.Error(err, "unable to create deployment")
		return ctrl.Result{}, err
	}

	err = ctrl.SetControllerReference(deployment, cmap, r.Scheme)
	if err != nil {
		logger.Error(err, "unable to set controller reference")
		return ctrl.Result{}, err
	}

	err = r.Create(ctx, cmap)
	if err != nil {
		logger.Error(err, "unable to create configmap")
		return ctrl.Result{}, err
	}

	err = ctrl.SetControllerReference(deployment, pvc, r.Scheme)
	if err != nil {
		logger.Error(err, "unable to set controller reference")
		return ctrl.Result{}, err
	}

	err = r.Create(ctx, pvc)
	if err != nil {
		logger.Error(err, "unable to create pvc")
		return ctrl.Result{}, err
	}

	err = ctrl.SetControllerReference(deployment, svc, r.Scheme)
	if err != nil {
		logger.Error(err, "unable to set controller reference")
		return ctrl.Result{}, err
	}

	err = r.Create(ctx, svc)
	if err != nil {
		logger.Error(err, "unable to create svc")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GameReconciler) update(ctx context.Context, game *operatorv1alpha1.Game) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func setupEventFilterPredicates() predicate.Predicate {
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
