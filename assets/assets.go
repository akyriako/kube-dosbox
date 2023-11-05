package assets

import (
	"bytes"
	"embed"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"text/template"
)

var (
	//go:embed manifests/*
	manifests  embed.FS
	appsScheme = runtime.NewScheme()
	appsCodecs = serializer.NewCodecFactory(appsScheme)
)

func init() {
	if err := appsv1.AddToScheme(appsScheme); err != nil {
		panic(err)
	}

	if err := corev1.AddToScheme(appsScheme); err != nil {
		panic(err)
	}
}

func getTemplate(name string) (*template.Template, error) {
	manifestBytes, err := manifests.ReadFile(fmt.Sprintf("manifests/%s.yaml", name))
	if err != nil {
		return nil, err
	}

	tmp := template.New(name)
	parse, err := tmp.Parse(string(manifestBytes))
	if err != nil {
		return nil, err
	}

	return parse, nil
}

func getObject(name string, gv schema.GroupVersion, metadata any) (runtime.Object, error) {
	parse, err := getTemplate(name)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = parse.Execute(&buffer, metadata)
	if err != nil {
		return nil, err
	}

	object, err := runtime.Decode(
		appsCodecs.UniversalDecoder(gv),
		buffer.Bytes(),
	)

	return object, nil
}

func GetDeployment(namespace string, name string, port int) (*appsv1.Deployment, error) {
	metadata := struct {
		Namespace string
		Name      string
		Port      int
	}{
		Namespace: namespace,
		Name:      name,
		Port:      port,
	}

	object, err := getObject("deployment", appsv1.SchemeGroupVersion, metadata)
	if err != nil {
		return nil, err
	}

	return object.(*appsv1.Deployment), nil
}

func GetPersistentVolumeClaim(namespace string, name string) (*corev1.PersistentVolumeClaim, error) {
	metadata := struct {
		Namespace string
		Name      string
	}{
		Namespace: namespace,
		Name:      name,
	}

	object, err := getObject("pvc", corev1.SchemeGroupVersion, metadata)
	if err != nil {
		return nil, err
	}

	return object.(*corev1.PersistentVolumeClaim), nil
}
