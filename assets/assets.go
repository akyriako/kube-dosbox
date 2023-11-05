package assets

import (
	"bytes"
	"embed"
	appsv1 "k8s.io/api/apps/v1"
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
}

func GetDeployment(namespace string, name string, port int) (*appsv1.Deployment, error) {
	deploymentBytes, err := manifests.ReadFile("manifests/deployment.yaml")
	if err != nil {
		return nil, err
	}

	tmp := template.New("deployment")
	parse, err := tmp.Parse(string(deploymentBytes))
	if err != nil {
		return nil, err
	}

	metadata := struct {
		Namespace string
		Name      string
		Port      int
	}{
		Namespace: namespace,
		Name:      name,
		Port:      port,
	}

	var deploymentParsedBytes bytes.Buffer
	err = parse.Execute(&deploymentParsedBytes, metadata)
	if err != nil {
		return nil, err
	}

	deploymentObject, err := runtime.Decode(
		appsCodecs.UniversalDecoder(appsv1.SchemeGroupVersion),
		deploymentParsedBytes.Bytes(),
	)
	if err != nil {
		return nil, err
	}

	return deploymentObject.(*appsv1.Deployment), nil
}
