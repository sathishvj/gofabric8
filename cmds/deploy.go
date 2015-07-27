/**
 * Copyright (C) 2015 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cmds

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"

	cmdutil "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd/util"
	"github.com/fabric8io/gofabric8/client"
	"github.com/fabric8io/gofabric8/util"
	"github.com/spf13/cobra"
)

const (
	consoleMetadataUrl           = "https://repo1.maven.org/maven2/io/fabric8/apps/base/maven-metadata.xml"
	baseConsoleUrl               = "https://repo1.maven.org/maven2/io/fabric8/apps/base/%s/base-%s-kubernetes.json"
	consoleKubernetesMetadataUrl = "https://repo1.maven.org/maven2/io/fabric8/apps/console-kubernetes/maven-metadata.xml"
	baseConsoleKubernetesUrl     = "https://repo1.maven.org/maven2/io/fabric8/apps/console-kubernetes/%s/console-kubernetes-%s-kubernetes.json"
)

func NewCmdDeploy(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy fabric8 to your Kubernetes or OpenShift environment",
		Long:  `deploy fabric8 to your Kubernetes or OpenShift environment`,
		Run: func(cmd *cobra.Command, args []string) {
			c, cfg := client.NewClient(f)
			ns, _, _ := f.DefaultNamespace()
			util.Info("Deploying fabric8 to your ")
			util.Success(string(util.TypeOfMaster(c)))
			util.Info(" installation at ")
			util.Success(cfg.Host)
			util.Info(" in namespace ")
			util.Successf("%s\n\n", ns)

			util.Info("Continue? [Y/n] ")
			cont := util.AskForConfirmation(true)
			if !cont {
				util.Fatal("Cancelled...\n")
			}

			v := cmd.Flags().Lookup("version").Value.String()

			v = f8Version(v, util.TypeOfMaster(c))

			util.Warnf("\nStarting deployment of %s...\n\n", v)
		},
	}

	return cmd
}

func f8Version(v string, typeOfMaster util.MasterType) string {
	metadataUrl := consoleMetadataUrl
	if typeOfMaster == util.Kubernetes {
		metadataUrl = consoleKubernetesMetadataUrl
	}

	resp, err := http.Get(metadataUrl)
	if err != nil {
		util.Fatalf("Cannot get fabric8 version to deploy: %v", err)
	}
	defer resp.Body.Close()
	// read xml http response
	xmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.Fatalf("Cannot get fabric8 version to deploy: %v", err)
	}

	type Metadata struct {
		Release  string   `xml:"versioning>release"`
		Versions []string `xml:"versioning>versions>version"`
	}

	var m Metadata
	err = xml.Unmarshal(xmlData, &m)
	if err != nil {
		util.Fatalf("Cannot get fabric8 version to deploy: %v", err)
	}

	if v == "latest" {
		return m.Release
	}

	for _, version := range m.Versions {
		if v == version {
			return version
		}
	}

	util.Errorf("\nUnknown version: %s\n", v)
	util.Fatalf("Valid versions: %v\n", append(m.Versions, "latest"))
	return ""
}
