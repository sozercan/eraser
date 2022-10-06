//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"testing"

	eraserv1alpha1 "github.com/Azure/eraser/api/v1alpha1"
	"github.com/Azure/eraser/test/e2e/util"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"

	pkgUtil "github.com/Azure/eraser/pkg/utils"
)

func TestMain(m *testing.M) {
	utilruntime.Must(eraserv1alpha1.AddToScheme(scheme.Scheme))

	util.Testenv = env.NewWithConfig(envconf.New())
	// Create KinD Cluster
	util.Testenv.Setup(
		envfuncs.CreateKindClusterWithConfig(util.KindClusterName, util.NodeVersion, "../../kind-config.yaml"),
		envfuncs.CreateNamespace(util.TestNamespace),
		envfuncs.CreateNamespace(util.EraserNamespace),
		envfuncs.LoadDockerImageToCluster(util.KindClusterName, util.ManagerImage),
		envfuncs.LoadDockerImageToCluster(util.KindClusterName, util.EraserImage),
		envfuncs.LoadDockerImageToCluster(util.KindClusterName, util.CollectorImage),
		envfuncs.LoadDockerImageToCluster(util.KindClusterName, util.VulnerableImage),
		envfuncs.LoadDockerImageToCluster(util.KindClusterName, util.NonVulnerableImage),
		util.CreateExclusionList(util.EraserNamespace, pkgUtil.ExclusionList{
			Excluded: []string{"docker.io/library/alpine:*"},
		}),
		util.CreateExclusionList(util.EraserNamespace, pkgUtil.ExclusionList{
			Excluded: []string{util.NonVulnerableImage},
		}),
		util.DeployEraserHelm(util.EraserNamespace,
			"--set", util.HelmControllerManagerImage,
			"--set", util.HelmEraserImage,
			"--set", util.HelmCollectorImage,
			"--set", `scanner.image.repository=`,
			"--set", `controllerManager.additionalArgs={--job-cleanup-on-success-delay=1m}`),
	).Finish(
		envfuncs.DestroyKindCluster(util.KindClusterName),
	)
	os.Exit(util.Testenv.Run(m))
}
