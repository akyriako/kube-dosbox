package controllers

import (
	"context"
	"fmt"
	operatorv1alpha1 "github.com/akyriako/kube-dosbox/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

func (r *GameReconciler) GetStatus(ctx context.Context, req ctrl.Request, appLabel string) (bool, error) {
	pods := &corev1.PodList{}
	opts := []client.ListOption{
		client.MatchingLabels(map[string]string{"app": appLabel}),
		client.InNamespace(req.Namespace),
	}

	if err := r.List(ctx, pods, opts...); err != nil {
		return false, err
	}

	for _, pod := range pods.Items {
		if len(pod.Status.InitContainerStatuses) > 0 && len(pod.Status.ContainerStatuses) > 0 {
			init := pod.Status.InitContainerStatuses[0]
			engine := pod.Status.InitContainerStatuses[0]

			return init.Ready && engine.Ready, nil
		}

		return false, fmt.Errorf("unable to find containers in pod")
	}

	return false, nil
}

func (r *GameReconciler) SetStatus(ctx context.Context, req ctrl.Request, game *operatorv1alpha1.Game, ready bool) error {
	patch := client.MergeFrom(game.DeepCopy())
	game.Status.Ready = &ready

	err := r.Status().Patch(ctx, game, patch)
	if err != nil {
		logger.Error(err, "unable to patch game status")
		return err
	}

	logger.Info(fmt.Sprintf("%s is ready", strings.ToLower(game.Spec.GameName)))
	return nil
}

func (r *GameReconciler) RefreshStatus(ctx context.Context, req ctrl.Request, game *operatorv1alpha1.Game, appLabel string) (ctrl.Result, error) {
	ready, err := r.GetStatus(ctx, req, appLabel)
	if err != nil {
		logger.Error(err, "unable to fetch pod status")

		_ = r.SetStatus(ctx, req, game, false)

		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 15 * time.Second,
		}, err
	}

	if !ready {
		logger.Info("pod not ready, requeue in 15sec")

		_ = r.SetStatus(ctx, req, game, ready)

		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 15 * time.Second,
		}, nil
	}

	err = r.SetStatus(ctx, req, game, ready)
	if err != nil {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 15 * time.Second,
		}, nil
	}

	return ctrl.Result{}, nil
}
