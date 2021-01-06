// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package utils

import (
	"fmt"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var clientSet *kubernetes.Clientset
var kubeConfig *rest.Config

func initKubeConfig() {
	kubeConfig = GetConfig()
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		panic(err)
	}

	clientSet = clientset
}

// GetConfig gets a kubernetes rest config
func GetConfig() *rest.Config {
	if kubeConfig != nil {
		return kubeConfig
	}

	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	conf, err := rest.InClusterConfig()
	if err != nil {
		conf, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err)
		}
	}

	return conf
}

// GetKubeClient gets a kubernetes client
func GetKubeClient() *kubernetes.Clientset {
	if clientSet == nil {
		initKubeConfig()
	}

	return clientSet
}

// ToISO8601DateTimeString converts dateTime to ISO8601 Format
// ISO8601 Format: 2020-01-01T01:01:01.10101Z
func ToISO8601DateTimeString(dateTime time.Time) string {
	t := dateTime.UTC()

	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	micros := t.Nanosecond() / 1000

	return fmt.Sprintf(
		"%04d-%02d-%02dT%02d:%02d:%02d.%03dZ",
		year, month, day, hour, minute, second, micros)
}
