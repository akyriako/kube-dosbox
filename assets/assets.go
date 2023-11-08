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
	manifests embed.FS

	//go:embed static/*
	static embed.FS

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

func GetService(namespace string, name string, port int) (*corev1.Service, error) {
	metadata := struct {
		Namespace string
		Name      string
		Port      int
	}{
		Namespace: namespace,
		Name:      name,
		Port:      port,
	}

	object, err := getObject("service", corev1.SchemeGroupVersion, metadata)
	if err != nil {
		return nil, err
	}

	return object.(*corev1.Service), nil
}

func GetPersistentVolumeClaim(namespace string, name string, storage uint64) (*corev1.PersistentVolumeClaim, error) {
	metadata := struct {
		Namespace string
		Name      string
		Storage   uint64
	}{
		Namespace: namespace,
		Name:      name,
		Storage:   storage,
	}

	object, err := getObject("pvc", corev1.SchemeGroupVersion, metadata)
	if err != nil {
		return nil, err
	}

	return object.(*corev1.PersistentVolumeClaim), nil
}

func GetConfigMap(namespace string, name string, bundle string) (*corev1.ConfigMap, error) {
	metadata := struct {
		Namespace string
		Name      string
		Bundle    string
	}{
		Namespace: namespace,
		Name:      name,
		Bundle:    bundle,
	}

	object, err := getObject("configmap", corev1.SchemeGroupVersion, metadata)
	if err != nil {
		return nil, err
	}

	return object.(*corev1.ConfigMap), nil
}

func GetIndex(bundle string) ([]byte, error) {
	staticBytes, err := static.ReadFile("static/index.html")
	if err != nil {
		return nil, err
	}

	tmp := template.New("index")
	parse, err := tmp.Parse(string(staticBytes))
	if err != nil {
		return nil, err
	}

	metadata := struct {
		Bundle string
	}{
		Bundle: bundle,
	}

	var buffer bytes.Buffer
	err = parse.Execute(&buffer, metadata)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
