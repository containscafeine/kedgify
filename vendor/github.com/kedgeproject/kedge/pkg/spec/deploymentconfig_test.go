/*
Copyright 2017 The Kedge Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package spec

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	os_deploy_v1 "github.com/openshift/origin/pkg/deploy/apis/apps/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_v1 "k8s.io/kubernetes/pkg/api/v1"
)

func TestFixDeploymentConfig(t *testing.T) {
	tests := []struct {
		name           string
		input          *DeploymentConfigSpecMod
		expectedOutput *DeploymentConfigSpecMod
	}{
		{
			name:  "No replicas passed at input, expected 1",
			input: &DeploymentConfigSpecMod{},
			expectedOutput: &DeploymentConfigSpecMod{
				ControllerFields: ControllerFields{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{
							appLabelKey: "",
						},
					},
				},
				DeploymentConfigSpec: os_deploy_v1.DeploymentConfigSpec{
					Replicas: 1,
				},
				Replicas: getInt32Addr(1),
			},
		},
		{
			name: "replicas set to 0 by the end user, expected 0",
			input: &DeploymentConfigSpecMod{
				Replicas: getInt32Addr(0),
			},
			expectedOutput: &DeploymentConfigSpecMod{
				ControllerFields: ControllerFields{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{
							appLabelKey: "",
						},
					},
				},
				DeploymentConfigSpec: os_deploy_v1.DeploymentConfigSpec{
					Replicas: 0,
				},
				Replicas: getInt32Addr(0),
			},
		},
		{
			name: "replicas set to 2 by the end user, expected 2",
			input: &DeploymentConfigSpecMod{
				Replicas: getInt32Addr(2),
			},
			expectedOutput: &DeploymentConfigSpecMod{
				ControllerFields: ControllerFields{
					ObjectMeta: meta_v1.ObjectMeta{
						Labels: map[string]string{
							appLabelKey: "",
						},
					},
				},
				DeploymentConfigSpec: os_deploy_v1.DeploymentConfigSpec{
					Replicas: 2,
				},
				Replicas: getInt32Addr(2),
			},
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			test.input.fixDeploymentConfig()
			if !reflect.DeepEqual(test.input, test.expectedOutput) {
				t.Errorf("Expected output to be:\n%v\nBut got:\n%v\n",
					prettyPrintObjects(test.expectedOutput),
					prettyPrintObjects(test.input))
			}
		})
	}
}

func TestDeploymentConfigSpecMod_CreateOpenShiftController(t *testing.T) {
	tests := []struct {
		name                    string
		deploymentConfigSpecMod *DeploymentConfigSpecMod
		deployment              *os_deploy_v1.DeploymentConfig
		success                 bool
	}{
		{
			name: "Test that it correctly converts",
			deploymentConfigSpecMod: &DeploymentConfigSpecMod{
				ControllerFields: ControllerFields{
					ObjectMeta: meta_v1.ObjectMeta{
						Name: "testJob",
					},
					Controller: "deploymentconfig",
					PodSpecMod: PodSpecMod{
						PodSpec: api_v1.PodSpec{
							Containers: []api_v1.Container{
								{
									Name:  "testContainer",
									Image: "testImage",
								},
							},
						},
					},
				},
				DeploymentConfigSpec: os_deploy_v1.DeploymentConfigSpec{
					Replicas: 2,
				},
			},
			deployment: &os_deploy_v1.DeploymentConfig{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "testJob",
				},
				Spec: os_deploy_v1.DeploymentConfigSpec{
					Replicas: 2,
					Template: &api_v1.PodTemplateSpec{
						ObjectMeta: meta_v1.ObjectMeta{
							Name: "testJob",
						},
						Spec: api_v1.PodSpec{
							Containers: []api_v1.Container{
								{
									Name:  "testContainer",
									Image: "testImage",
								},
							},
						},
					},
				},
			},
			success: true,
		},
		{
			name: "Test that strategy is converted correctly",
			deploymentConfigSpecMod: &DeploymentConfigSpecMod{
				ControllerFields: ControllerFields{
					ObjectMeta: meta_v1.ObjectMeta{
						Name: "testJob",
					},
					Controller: "deploymentconfig",
					PodSpecMod: PodSpecMod{
						PodSpec: api_v1.PodSpec{
							Containers: []api_v1.Container{
								{
									Name:  "testContainer",
									Image: "testImage",
								},
							},
						},
					},
				},
				DeploymentConfigSpec: os_deploy_v1.DeploymentConfigSpec{
					Replicas: 3,
					Strategy: os_deploy_v1.DeploymentStrategy{
						Type: os_deploy_v1.DeploymentStrategyType("Rolling"),
					},
				},
			},
			deployment: &os_deploy_v1.DeploymentConfig{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "testJob",
				},
				Spec: os_deploy_v1.DeploymentConfigSpec{
					Replicas: 3,
					Strategy: os_deploy_v1.DeploymentStrategy{
						Type: os_deploy_v1.DeploymentStrategyType("Rolling"),
					},
					Template: &api_v1.PodTemplateSpec{
						ObjectMeta: meta_v1.ObjectMeta{
							Name: "testJob",
						},
						Spec: api_v1.PodSpec{
							Containers: []api_v1.Container{
								{
									Name:  "testContainer",
									Image: "testImage",
								},
							},
						},
					},
				},
			},
			success: true,
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			dc, err := test.deploymentConfigSpecMod.createOpenShiftController()

			switch test.success {
			case true:
				if err != nil {
					t.Errorf("Expected test to pass but got an error -\n%v", err)
				}
			case false:
				if err == nil {
					t.Errorf("For the input -\n%v\nexpected test to fail, but test passed", spew.Sprint(test.deploymentConfigSpecMod))
				}
			}

			if !reflect.DeepEqual(test.deployment, dc) {

				t.Errorf("Expected OpenShift DeploymentConfig to be -\n%v\nBut got -\n%v", prettyPrintObjects(test.deployment), prettyPrintObjects(dc))
			}
		})
	}
}
