package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/pragkent/aliyun-disk-provisioner/cloud"
	"github.com/pragkent/aliyun-disk-provisioner/volume"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	id          = flag.String("id", "default", "Unique provisioner identity")
	provisioner = flag.String("provisioner", "pragkent.me/aliyun-disk", "Name of the provisioner")
	master      = flag.String("master", "", "Master URL")
	kubeconfig  = flag.String("kubeconfig", "", "Absolute path to the kubeconfig")
	cluster     = flag.String("cluster", "", "Cluster name")
	version     = flag.Bool("version", false, "Show version")
)

func main() {
	flag.Parse()
	flag.Set("logtostderr", "true")

	if *version {
		os.Exit(printVersion())
	}

	os.Exit(run())
}

func printVersion() int {
	var versionString bytes.Buffer

	fmt.Fprintf(&versionString, "%s version %s", Name, Version)
	if GitCommit != "" {
		fmt.Fprintf(&versionString, " (%s)", GitCommit)
	}

	fmt.Println(versionString.String())
	return 0
}

func run() int {
	if errs := validateProvisioner(*provisioner, field.NewPath("provisioner")); len(errs) != 0 {
		glog.Errorf("Invalid provisioner specified: %v", errs)
		return 1
	}
	glog.Infof("Provisioner %s specified", *provisioner)

	clientset, err := newClientSet()
	if err != nil {
		return 1
	}

	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		glog.Errorf("Error getting server version: %v", err)
		return 1
	}

	provider, err := newCloudProvider()
	if err != nil {
		glog.Error(err)
		return 1
	}

	pr := volume.NewProvisioner(*id, getClusterName(provider), clientset, provider)
	pc := controller.NewProvisionController(
		clientset,
		*provisioner,
		pr,
		serverVersion.GitVersion)
	pc.Run(wait.NeverStop)

	return 0
}

func validateProvisioner(provisioner string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(provisioner) == 0 {
		allErrs = append(allErrs, field.Required(fldPath, provisioner))
	}

	if len(provisioner) > 0 {
		for _, msg := range validation.IsQualifiedName(strings.ToLower(provisioner)) {
			allErrs = append(allErrs, field.Invalid(fldPath, provisioner, msg))
		}
	}

	return allErrs
}

func newClientSet() (kubernetes.Interface, error) {
	var config *rest.Config
	var err error

	if *master != "" || *kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		glog.Errorf("Failed to get client config: %v", err)
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorf("Failed to create client: %v", err)
		return nil, err
	}

	return clientset, err
}

func newCloudProvider() (cloud.Provider, error) {
	region := os.Getenv("ALIYUN_REGION")
	accessKey := os.Getenv("ALIYUN_ACCESS_KEY")
	secretKey := os.Getenv("ALIYUN_SECRET_KEY")

	if region == "" || accessKey == "" || secretKey == "" {
		return nil, errors.New("ALIYUN_REGION, ALIYUN_ACCESS_KEY or ALIYUN_SECRET_KEY is missing")
	}

	return cloud.NewProvider(accessKey, secretKey, region), nil
}

func getClusterName(provider cloud.Provider) string {
	if *cluster != "" {
		return *cluster
	}

	return provider.Cluster()
}
