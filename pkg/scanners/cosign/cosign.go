package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	eraserv1alpha1 "github.com/Azure/eraser/api/v1alpha1"
	"github.com/Azure/eraser/pkg/logger"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sigstore/cosign/cmd/cosign/cli/fulcio"
	"github.com/sigstore/cosign/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/pkg/cosign"

	machinerytypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	_ "net/http/pprof"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	generalErr = 1

	apiPath         = "apis/eraser.sh/v1alpha1"
	resourceName    = "imagecollectors"
	subResourceName = "status"
)

var (
	collectorCRName = flag.String("collector-cr-name", "collector-cr", "name of the collector cr to read from and write to")
	enableProfile   = flag.Bool("enable-pprof", false, "enable pprof profiling")
	profilePort     = flag.Int("pprof-port", 6060, "port for pprof profiling. defaulted to 6060 if unspecified")

	log = logf.Log.WithName("scanner").WithValues("provider", "cosign")
)

type (
	patch struct {
		Status eraserv1alpha1.ImageCollectorStatus `json:"status"`
	}

	statusUpdate struct {
		apiPath          string
		ctx              context.Context
		clientset        *kubernetes.Clientset
		collectorCRName  string
		resourceName     string
		subResourceName  string
		vulnerableImages []eraserv1alpha1.Image
		failedImages     []eraserv1alpha1.Image
	}
)

func main() {
	flag.Parse()
	ctx := context.Background()

	if err := logger.Configure(); err != nil {
		fmt.Fprintln(os.Stderr, "Error setting up logger:", err)
		os.Exit(generalErr)
	}

	if *enableProfile {
		go func() {
			err := http.ListenAndServe(fmt.Sprintf("localhost:%d", *profilePort), nil)
			log.Error(err, "pprof server failed")
		}()
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "unable to get in-cluster config")
		os.Exit(generalErr)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "unable to get REST client")
		os.Exit(generalErr)
	}

	result := eraserv1alpha1.ImageCollector{}

	err = clientset.RESTClient().Get().
		AbsPath(apiPath).
		Resource(resourceName).
		Name(*collectorCRName).
		Do(context.Background()).
		Into(&result)
	if err != nil {
		log.Error(err, "RESTClient GET request failed", "apiPath", apiPath, "recourceName", resourceName, "collectorCRName", *collectorCRName)
		os.Exit(generalErr)
	}

	ro := options.RegistryOptions{}
	co, err := ro.ClientOpts(ctx)
	if err != nil {
		log.Error(err, "error while getting client opts")
	}

	vulnerableImages := make([]eraserv1alpha1.Image, 0, len(result.Spec.Images))
	failedImages := make([]eraserv1alpha1.Image, 0, len(result.Spec.Images))

	for k := range result.Spec.Images {
		img := result.Spec.Images[k]
		imageRef := img.Name

		ref, err := name.ParseReference(imageRef)
		if err != nil {
			log.Error(err, "error while parsing reference")
			continue
		}

		_, bundleVerified, err := cosign.VerifyImageSignatures(ctx, ref, &cosign.CheckOpts{
			RegistryClientOpts: co,
			RootCerts:          fulcio.GetRoots(),
		})
		if err != nil {
			log.Error(err, "error while verifying image signatures")
			failedImages = append(failedImages, img)
			continue
		}

		if bundleVerified {
			log.Info("signature verified", "image", imageRef)
		} else {
			log.Info("no valid signatures found", "image", imageRef)
			vulnerableImages = append(vulnerableImages, img)
		}
	}

	err = updateStatus(&statusUpdate{
		apiPath:          apiPath,
		ctx:              ctx,
		clientset:        clientset,
		collectorCRName:  *collectorCRName,
		resourceName:     resourceName,
		subResourceName:  subResourceName,
		vulnerableImages: vulnerableImages,
		failedImages:     failedImages,
	})
	if err != nil {
		log.Error(err, "error updating ImageCollectorStatus", "images", vulnerableImages)
		os.Exit(generalErr)
	}

	log.Info("scanning complete, exiting")
}

func updateStatus(opts *statusUpdate) error {
	collectorPatch := patch{
		Status: eraserv1alpha1.ImageCollectorStatus{
			Vulnerable: opts.vulnerableImages,
			Failed:     opts.failedImages,
		},
	}

	body, err := json.Marshal(&collectorPatch)
	if err != nil {
		return err
	}

	_, err = opts.clientset.RESTClient().Patch(machinerytypes.MergePatchType).
		AbsPath(opts.apiPath).
		Resource(opts.resourceName).
		SubResource(opts.subResourceName).
		Name(opts.collectorCRName).
		Body(body).DoRaw(opts.ctx)

	return err
}
