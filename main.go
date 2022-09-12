// Copyright 2020 Google LLC
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// package main
//
// This KPT function is a helper function to make updating the Climate Insights infrastructure deployment
// easier. It's not requred and everything can be changed manually. This helper function just sets up some
// files from the vaules changed in the interface.yaml file.
package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

// main function is the entrypoint for the kpt function sdk. It will call the exported Run function
func main() {
	if err := fn.AsMain(fn.ResourceListProcessorFunc(Run)); err != nil {
		os.Exit(1)
	}
}

// The Run function is the core function that will the kpt functions resource list. This object consists of 2 main objects, the function config yaml file
// and an slice of all the other yaml files in the a kubeObject format
//
// The function will pull the relevant data from the function config file then will reach into the other files to set the paramaters. This is purley a 
// simplified interface method for those that do not wish to modify configuration manually. This will assume a bunch of defaults stay the same.
func Run(rl *fn.ResourceList) (bool, error) {
	ko := rl.FunctionConfig
	dataPtr := ko.GetMap("data")

	subNetwork := dataPtr.GetString("subnetwork-range")
	namespace := dataPtr.GetString("namespace")
	
	for _, item := range rl.Items {
		// Automatically add the Private GKE masterAuthorizedNetwork configs CIDR block to match that of the subnet that the bastion hist
		// will be created in.
		if len(subNetwork) > 0  && item.IsGVK("container.cnrm.cloud.google.com", "v1beta1", "ContainerCluster") {
			authNetwork := make([]map[string]string, 1)

			authNet := make(map[string]string)
			authNet["cidrBlock"] = subNetwork
			authNet["displayName"] = "private-net"
			authNetwork[0] = authNet

			err := item.SetNestedField(authNetwork, "spec", "masterAuthorizedNetworksConfig", "cidrBlocks")

			if err != nil {
				rl.Results = append(rl.Results, fn.ErrorResult(err))
			} else {
				rl.Results = append(rl.Results, fn.GeneralResult("Added auth-network block for private GKE cluster to match subnetwork-range", fn.Info))
			}
		}

		// Update the apply-time mutation on the SQL IAM accounts so thaty they  use the private SQL instances service account. Since the
		// sqlInstance and namespace are supplied as setter values this needs to be updated.
		if len(namespace) > 0 && item.IsGVK("iam.cnrm.cloud.google.com", "v1beta1", "IAMPolicyMember") {
			name := item.GetName()

			switch name {
			case "k8s-developer-access":
				item.SetNestedString(namespace, "spec", "memberFrom", "serviceAccountRef", "namespace")
				rl.Results = append(rl.Results, fn.GeneralResult("Updated k8-devloper-role referenced namesace", fn.Info))
			}
		}

		if len(namespace) > 0 {
			// Set the namespaces at the resource level so that the KRMs are appied in the correct KCC namespace
			err := item.SetNestedString(namespace, "metadata", "namespace")

			if err != nil {
				rl.Results = append(rl.Results, fn.ErrorResult(err))
			} else {
				rl.Results = append(rl.Results, fn.GeneralResult("Updated "+item.GetName()+" namespace", fn.Info))
			}

			// Updated the namespace in the projectserviceset config file so that when it creates the KRM resources for each service
			// the namespace is correct.
			if item.IsGVK("blueprints.cloud.google.com", "v1alpha1", "ProjectServiceSet") {
				if len(item.GetMap("metadata").GetString("namespace")) > 0 {
					err := item.SetNestedString(namespace, "metadata", "namespace")

					if err != nil {
						rl.Results = append(rl.Results, fn.ErrorResult(err))
					} else {
						rl.Results = append(rl.Results, fn.GeneralResult("Updated project-services namespace", fn.Info))
					}
				}
			}
		}
	}

	return true, nil
}
