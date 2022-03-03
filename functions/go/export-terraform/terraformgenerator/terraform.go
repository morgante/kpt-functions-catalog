// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terraformgenerator

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/pkg/errors"
)

func formatTerraformConfig(files map[string]string) (map[string]string, error) {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.0.6")),
	}

	ctx := context.Background()

	execPath, err := installer.Install(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "Error installing Terraform")
	}

	formattedFiles := make(map[string]string)
	for name, content := range files {
		fmt.Printf("format %s: %s", name, content)
		formatted, err := tfexec.FormatString(ctx, execPath, content)
		if err != nil {
			return nil, errors.Wrapf(err, "Error formatting %s", name)
		}
		formattedFiles[name] = formatted
	}
	return formattedFiles, nil
}

func (rs *terraformResources) getHCL() (map[string]string, error) {
	tmpl, err := template.New("").ParseFS(templates, "templates/*")
	if err != nil {
		return nil, err
	}

	groupedResources := rs.getGrouped()

	data := make(map[string]string)
	resourceFiles := []string{"folders.tf", "iam.tf", "projects.tf"}
	for _, file := range resourceFiles {
		err := addFile(tmpl, file, groupedResources, data)
		if err != nil {
			return nil, err
		}
	}

	// only add other files if resource files exist
	metaFiles := []string{"README.md", "versions.tf", "variables.tf"}
	if len(data) > 0 {
		for _, file := range metaFiles {
			err := addFile(tmpl, file, rs, data)
			if err != nil {
				return nil, err
			}
		}
	}

	return data, nil
}

func addFile(tmpl *template.Template, name string, inputData interface{}, data map[string]string) error {
	builder := strings.Builder{}
	wr := &(builder)

	err := tmpl.ExecuteTemplate(wr, name, inputData)
	if err != nil {
		return err
	}

	content := strings.TrimSpace(builder.String())

	if len(content) < 1 {
		return nil
	}

	content = content + "\n"
	data[name] = content

	return nil
}
