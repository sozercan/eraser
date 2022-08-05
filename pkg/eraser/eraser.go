package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/kubectl/pkg/scheme"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Azure/eraser/pkg/logger"

	util "github.com/Azure/eraser/pkg/utils"
	restclient "k8s.io/client-go/rest"
)

var (
	runtimePtr    = flag.String("runtime", "containerd", "container runtime")
	imageListPtr  = flag.String("imagelist", "", "name of ImageList")
	enableProfile = flag.Bool("enable-pprof", false, "enable pprof profiling")
	profilePort   = flag.Int("pprof-port", 6060, "port for pprof profiling. defaulted to 6060 if unspecified")
	emitRemovalEvents    = flag.Bool("emit-removal-events", true, "[alpha] emit events for removed images")

	// Timeout of connecting to server (default: 5m).
	timeout  = 5 * time.Minute
	log      = logf.Log.WithName("eraser")
	excluded map[string]struct{}
)

const (
	excludedPath = "/run/eraser.sh/excluded/excluded"
)

func main() {
	flag.Parse()
	if *enableProfile {
		go func() {
			err := http.ListenAndServe(fmt.Sprintf("localhost:%d", *profilePort), nil)
			log.Error(err, "pprof server failed")
		}()
	}

	if err := logger.Configure(); err != nil {
		fmt.Fprintln(os.Stderr, "Error setting up logger:", err)
		os.Exit(1)
	}

	socketPath, found := util.RuntimeSocketPathMap[*runtimePtr]
	if !found {
		log.Error(fmt.Errorf("unsupported runtime"), "runtime", *runtimePtr)
		os.Exit(1)
	}

	imageclient, conn, err := util.GetImageClient(context.Background(), socketPath)
	if err != nil {
		log.Error(err, "failed to get image client")
		os.Exit(1)
	}

	runtimeClient := pb.NewRuntimeServiceClient(conn)
	client := client{imageclient, runtimeClient}

	imagelist, err := util.ParseImageList(*imageListPtr)
	if err != nil {
		log.Error(err, "failed to parse image list file")
		os.Exit(1)
	}

	excluded, err = util.ParseExcluded(excludedPath)
	if err != nil {
		log.Error(err, "failed to parse exclusion list")
		os.Exit(1)
	}
	if len(excluded) == 0 {
		log.Info("excluded configmap was empty or does not exist")
	}

	var recorder record.EventRecorder
	if *emitRemovalEvents {
		kubeConfig, err := restclient.InClusterConfig()
		if err != nil {
			log.Error(err, "failed to get Kubernetes config")
			os.Exit(1)
		}
		kubeClient, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			log.Error(err, "failed to get Kubernetes client")
			os.Exit(1)
		}
		recorder = eventRecorder(kubeClient)
	}

	if err := removeImages(&client, imagelist, recorder); err != nil {
		log.Error(err, "failed to remove images")
		os.Exit(1)
	}
}

func eventRecorder(kubeClient *kubernetes.Clientset) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{
			Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme,
		v1.EventSource{Component: "eraser"})
	return recorder
}
