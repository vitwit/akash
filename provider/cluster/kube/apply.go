package kube

import (
	"context"

	akashv1 "github.com/ovrclk/akash/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func applyNS(kc kubernetes.Interface, b *nsBuilder) error {
	// TODO: accept context as parameter
	ctx := context.Background()
	obj, err := kc.CoreV1().Namespaces().Get(ctx, b.name(), metav1.GetOptions{})
	switch {
	case err == nil:
		obj, err = b.update(obj)
		if err == nil {
			_, err = kc.CoreV1().Namespaces().Update(ctx, obj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		obj, err = b.create()
		if err == nil {
			_, err = kc.CoreV1().Namespaces().Create(ctx, obj, metav1.CreateOptions{})
		}
	}
	return err
}

func applyDeployment(kc kubernetes.Interface, b *deploymentBuilder) error {
	// TODO: accept context as parameter
	ctx := context.Background()
	obj, err := kc.AppsV1().Deployments(b.ns()).Get(ctx, b.name(), metav1.GetOptions{})
	switch {
	case err == nil:
		obj, err = b.update(obj)
		if err == nil {
			_, err = kc.AppsV1().Deployments(b.ns()).Update(ctx, obj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		obj, err = b.create()
		if err == nil {
			_, err = kc.AppsV1().Deployments(b.ns()).Create(ctx, obj, metav1.CreateOptions{})
		}
	}
	return err
}

func applyService(kc kubernetes.Interface, b *serviceBuilder) error {
	// TODO: accept context as parameter
	ctx := context.Background()
	obj, err := kc.CoreV1().Services(b.ns()).Get(ctx, b.name(), metav1.GetOptions{})
	switch {
	case err == nil:
		obj, err = b.update(obj)
		if err == nil {
			_, err = kc.CoreV1().Services(b.ns()).Update(ctx, obj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		obj, err = b.create()
		if err == nil {
			_, err = kc.CoreV1().Services(b.ns()).Create(ctx, obj, metav1.CreateOptions{})
		}
	}
	return err
}

func applyIngress(kc kubernetes.Interface, b *ingressBuilder) error {
	// TODO: accept context as parameter
	ctx := context.Background()
	obj, err := kc.ExtensionsV1beta1().Ingresses(b.ns()).Get(ctx, b.name(), metav1.GetOptions{})
	switch {
	case err == nil:
		obj, err = b.update(obj)
		if err == nil {
			_, err = kc.ExtensionsV1beta1().Ingresses(b.ns()).Update(ctx, obj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		obj, err = b.create()
		if err == nil {
			_, err = kc.ExtensionsV1beta1().Ingresses(b.ns()).Create(ctx, obj, metav1.CreateOptions{})
		}
	}
	return err
}

func prepareEnvironment(kc kubernetes.Interface, ns string) error {
	// TODO: accept context as parameter
	ctx := context.Background()
	_, err := kc.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		obj := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}
		_, err = kc.CoreV1().Namespaces().Create(ctx, obj, metav1.CreateOptions{})
	}
	return err
}

func applyManifest(kc akashv1.Interface, b *manifestBuilder) error {
	// TODO: accept context as parameter
	ctx := context.Background()
	obj, err := kc.AkashV1().Manifests(b.ns()).Get(ctx, b.name(), metav1.GetOptions{})
	switch {
	case err == nil:
		obj, err = b.update(obj)
		if err == nil {
			_, err = kc.AkashV1().Manifests(b.ns()).Update(ctx, obj, metav1.UpdateOptions{})
		}
	case errors.IsNotFound(err):
		obj, err = b.create()
		if err == nil {
			_, err = kc.AkashV1().Manifests(b.ns()).Create(ctx, obj, metav1.CreateOptions{})
		}
	}
	return err
}
