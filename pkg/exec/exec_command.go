package exec

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// Options passed to WithOptions
type Options struct {
	Command []string

	Namespace     string
	PodName       string
	ContainerName string

	Stdin         io.Reader
	CaptureStdout bool
	CaptureStderr bool
}

// Command run a command in a pod
func Command(podName, namespaceName string, cmd ...string) (string, error) {

	pod, err := getPod(podName, namespaceName)
	if err != nil {
		return "", fmt.Errorf("could not get pod info: %v", err)
	}

	if len(pod.Spec.Containers) != 1 {
		return "", fmt.Errorf("could not determine which container to use")
	}

	if pod.Status.ContainerStatuses[0].Ready != true {
		return "", fmt.Errorf("container is not ready")
	}

	execOut, execErr, err := CommandInContainer(podName, pod.Spec.Containers[0].Name, namespaceName, cmd...)

	if err != nil {
		return "", fmt.Errorf("could not execute: %v", err)
	}

	if len(execErr) > 0 {
		return "", fmt.Errorf("stderr: %v", execErr)
	}

	return execOut, nil
}

// CommandInContainer command in the
// specified container and return stdout, stderr and error
func CommandInContainer(podName, containerName, namespaceName string, cmd ...string) (string, string, error) {
	return WithOptions(Options{
		Command:       cmd,
		Namespace:     namespaceName,
		PodName:       podName,
		ContainerName: containerName,

		Stdin:         nil,
		CaptureStdout: true,
		CaptureStderr: true,
	})
}

// WithOptions executes a command in the specified container,
// returning stdout, stderr and error. `options` allowed for
// additional parameters to be passed.
func WithOptions(options Options) (string, string, error) {

	config, err := LoadConfig()

	if err != nil {
		panic("failed to load restclient config")
	}

	kubeClient := NewClientSet(config)

	const tty = false

	req := kubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(options.PodName).
		Namespace(options.Namespace).
		SubResource("exec").
		Param("container", options.ContainerName)
	req.VersionedParams(&v1.PodExecOptions{
		Container: options.ContainerName,
		Command:   options.Command,
		Stdin:     options.Stdin != nil,
		Stdout:    options.CaptureStdout,
		Stderr:    options.CaptureStderr,
		TTY:       tty,
	}, scheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	err = execute("POST", req.URL(), config, options.Stdin, &stdout, &stderr, tty)

	return stdout.String(), stderr.String(), err
}

func execute(method string, url *url.URL, config *rest.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
}

// LoadConfig returns the config of the K8S cluster
func LoadConfig() (*rest.Config, error) {
	var cfg *rest.Config
	var err error
	if os.Getenv(k8sutil.KubeConfigEnvVar) != "" {
		cfg, err = outOfClusterConfig()
	} else {
		cfg, err = inClusterConfig()
	}
	return cfg, err
}

// NewClientSet creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewClientSet(cfg *rest.Config) kubernetes.Interface {
	return kubernetes.NewForConfigOrDie(cfg)
}

// inClusterConfig returns the in-cluster config accessible inside a pod
func inClusterConfig() (*rest.Config, error) {
	// Work around https://github.com/kubernetes/kubernetes/issues/40973
	// See https://github.com/coreos/etcd-operator/issues/731#issuecomment-283804819
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		addrs, err := net.LookupHost("kubernetes.default.svc")
		if err != nil {
			return nil, err
		}
		err = os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
		if err != nil {
			return nil, err
		}
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		err := os.Setenv("KUBERNETES_SERVICE_PORT", "443")
		if err != nil {
			return nil, err
		}
	}
	return rest.InClusterConfig()
}

func outOfClusterConfig() (*rest.Config, error) {
	kubeconfig := os.Getenv(k8sutil.KubeConfigEnvVar)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	return config, err
}

func getPod(podName, namespaceName string) (*v1.Pod, error) {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespaceName,
		},
	}

	err := sdk.Get(pod)

	return pod, err
}
