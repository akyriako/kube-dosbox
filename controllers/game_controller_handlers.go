package controllers

import (
	"context"
	"fmt"
	operatorv1alpha1 "github.com/akyriako/kube-dosbox/api/v1alpha1"
	"github.com/akyriako/kube-dosbox/assets"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *GameReconciler) CreateOrUpdateDeployment(
	ctx context.Context,
	req ctrl.Request,
	game *operatorv1alpha1.Game,
) (*appsv1.Deployment, error) {
	create := false

	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			create = true
		} else {
			logger.V(5).Error(err, "unable to fetch deployment")
			return nil, err
		}
	}

	if create {
		deployment, err = assets.GetDeployment(game.Namespace, game.Name, game.Spec.Port)
		if err != nil {
			logger.Error(err, "unable to parse deployment template")
			return nil, err
		}

		err = ctrl.SetControllerReference(game, deployment, r.Scheme)
		if err != nil {
			logger.Error(err, "unable to set controller reference")
			return nil, err
		}

		err = r.Create(ctx, deployment)
		if err != nil {
			logger.Error(err, "unable to create deployment")
			return nil, err
		}

		return deployment, nil
	}

	return deployment, nil
}

func (r *GameReconciler) CreateOrUpdateConfigMap(
	ctx context.Context,
	req ctrl.Request,
	game *operatorv1alpha1.Game,
	deployment *appsv1.Deployment,
) (*corev1.ConfigMap, error) {
	create := false

	cmap := &corev1.ConfigMap{}
	objectKey := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      fmt.Sprintf("%s-index-configmap", req.Name),
	}
	err := r.Get(ctx, objectKey, cmap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			create = true
		} else {
			logger.V(5).Error(err, "unable to fetch configmap")
			return nil, err
		}
	}

	if create {
		cmap, err = assets.GetConfigMap(game.Namespace, game.Name, filepath.Base(game.Spec.Url))
		if err != nil {
			logger.Error(err, "unable to parse configmap template")
			return nil, err
		}

		err = ctrl.SetControllerReference(deployment, cmap, r.Scheme)
		if err != nil {
			logger.Error(err, "unable to set controller reference")
			return nil, err
		}

		err = r.Create(ctx, cmap)
		if err != nil {
			logger.Error(err, "unable to create configmap")
			return nil, err
		}

		return cmap, nil
	}

	return cmap, nil
}

func (r *GameReconciler) CreateOrUpdatePersistentVolumeClaim(
	ctx context.Context,
	req ctrl.Request,
	game *operatorv1alpha1.Game,
	deployment *appsv1.Deployment,
) (*corev1.PersistentVolumeClaim, error) {
	create := false

	pvc := &corev1.PersistentVolumeClaim{}
	objectKey := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      fmt.Sprintf("%s-pvc", req.Name),
	}
	err := r.Get(ctx, objectKey, pvc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			create = true
		} else {
			logger.V(5).Error(err, "unable to fetch pvc")
			return nil, err
		}
	}

	if create {
		pvc, err = assets.GetPersistentVolumeClaim(game.Namespace, game.Name)
		if err != nil {
			logger.Error(err, "unable to parse pvc template")
			return nil, err
		}

		err = ctrl.SetControllerReference(deployment, pvc, r.Scheme)
		if err != nil {
			logger.Error(err, "unable to set controller reference")
			return nil, err
		}

		err = r.Create(ctx, pvc)
		if err != nil {
			logger.Error(err, "unable to create pvc")
			return nil, err
		}

		return pvc, nil
	}

	return pvc, nil
}

func (r *GameReconciler) CreateOrUpdateService(
	ctx context.Context,
	req ctrl.Request,
	game *operatorv1alpha1.Game,
	deployment *appsv1.Deployment,
) (*corev1.Service, error) {
	create := false

	svc := &corev1.Service{}
	err := r.Get(ctx, req.NamespacedName, svc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			create = true
		} else {
			logger.V(5).Error(err, "unable to fetch svc")
			return nil, err
		}
	}

	if create {
		svc, err = assets.GetService(game.Namespace, game.Name, game.Spec.Port)
		if err != nil {
			logger.Error(err, "unable to parse svc template")
			return nil, err
		}

		err = ctrl.SetControllerReference(deployment, svc, r.Scheme)
		if err != nil {
			logger.Error(err, "unable to set controller reference")
			return nil, err
		}

		err = r.Create(ctx, svc)
		if err != nil {
			logger.Error(err, "unable to create svc")
			return nil, err
		}

		return svc, nil
	}

	svcPort := svc.Spec.Ports[0].Port
	specPort := int32(game.Spec.Port)

	if svcPort != specPort {
		dc := svc.DeepCopy()
		dc.Spec.Ports[0].Port = specPort

		err = ctrl.SetControllerReference(game, deployment, r.Scheme)
		if err != nil {
			logger.Error(err, "unable to set controller reference")
			return nil, err
		}

		err = r.Update(ctx, dc)
		if err != nil {
			logger.Error(err, "unable to update svc")
			return nil, err
		}
	}

	return svc, nil
}
