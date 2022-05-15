package k8s

import (
	keda "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateScaledObject(nameSpace, name, refName, scalerType string) *keda.ScaledObject {

	scaledObject := &keda.ScaledObject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nameSpace,
		},
		Spec: keda.ScaledObjectSpec{
			ScaleTargetRef: &keda.ScaleTarget{
				Name: refName,
			},
			Triggers: []keda.ScaleTriggers{
				{
					Type: scalerType,
				},
			},
		},
	}
	return scaledObject
}
