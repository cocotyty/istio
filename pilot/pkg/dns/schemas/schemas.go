package schemas

import (
	istioioapimetav1alpha1 "istio.io/api/meta/v1alpha1"
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config/schema/collection"
	"istio.io/istio/pkg/config/schema/resource"
	"istio.io/istio/pkg/config/validation"
	"reflect"
)

var EgressSidecarSchema = (&collection.Builder{Name: "k8s/networking.istio.io/v1alpha3/sidecars",
	VariableName: "K8SNetworkingInternalIstioIoV1Alpha3Sidecars",
	Disabled:     false,
	Resource: resource.Builder{
		Group:   "networking.internal.istio.io",
		Kind:    "Sidecar",
		Plural:  "sidecars",
		Version: "v1alpha3",
		Proto:   "istio.networking.v1alpha3.Sidecar", StatusProto: "istio.meta.v1alpha1.IstioStatus",
		ReflectType: reflect.TypeOf(&networking.Sidecar{}).Elem(), StatusType: reflect.TypeOf(&istioioapimetav1alpha1.IstioStatus{}).Elem(),
		ProtoPackage: "istio.io/api/networking/v1alpha3", StatusPackage: "istio.io/api/meta/v1alpha1",
		ClusterScoped: false,
		ValidateProto: validation.ValidateSidecar,
	}.MustBuild(),
}).MustBuild()

var EgressSidecarGVK = EgressSidecarSchema.Resource().GroupVersionKind()
var EgressScopeSchemas = collection.NewSchemasBuilder().MustAdd(EgressSidecarSchema).Build()
