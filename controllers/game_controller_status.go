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

func (r *GameReconciler) GetStatus(
	ctx context.Context,
	req ctrl.Request,
	appLabel string,
) (bool, error) {
	pods := &corev1.PodList{}
	opts := []client.ListOption{
		client.MatchingLabels(map[string]string{"app": appLabel}),
		client.InNamespace(req.Namespace),
	}

	if err := r.List(ctx, pods, opts...); err != nil {
		return false, err
	}

	ready := false
	for _, pod := range pods.Items {
		if len(pod.Status.InitContainerStatuses) > 0 && len(pod.Status.ContainerStatuses) > 0 {
			init := pod.Status.InitContainerStatuses[0]
			engine := pod.Status.InitContainerStatuses[0]

			ready = init.Ready && engine.Ready
		}

		// If at *least one* of the Pods in the Deployment is Ready
		// declare the whole Game as Ready to be played.
		if ready {
			break
		}
	}

	return ready, nil
}

func (r *GameReconciler) SetStatus(
	ctx context.Context,
	req ctrl.Request,
	game *operatorv1alpha1.Game,
	ready bool,
) error {
	patch := client.MergeFrom(game.DeepCopy())
	if ready == true {
		game.Status.Ready = &ready
	} else {
		game.Status.Ready = nil
	}

	err := r.Status().Patch(ctx, game, patch)
	if err != nil {
		logger.V(5).Error(err, "unable to patch game status")
		return err
	}

	if ready {
		logger.Info(fmt.Sprintf("%s is ready", strings.ToLower(game.Name)))
	}

	return nil
}

func (r *GameReconciler) RefreshStatus(
	ctx context.Context,
	req ctrl.Request,
	game *operatorv1alpha1.Game,
	appLabel string,
) (ctrl.Result, error) {
	ready, err := r.GetStatus(ctx, req, appLabel)
	if err != nil {
		logger.V(5).Error(err, "unable to fetch pod status")

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
