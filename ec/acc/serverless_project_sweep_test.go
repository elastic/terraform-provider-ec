// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package acc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/elastic/terraform-provider-ec/ec/internal/gen/serverless"
)

// serverlessProjectStaleAfter is the minimum age before the sweeper deletes a
// project. It must exceed the ACC test timeout (120m) to avoid deleting
// resources from a still-running build when another build's pre-exit runs.
const serverlessProjectStaleAfter = 3 * time.Hour

func init() {
	resource.AddTestSweepers("ec_serverless_projects", &resource.Sweeper{
		Name: "ec_serverless_projects",
		F:    testSweepServerlessProjects,
	})
}

type serverlessProjectSweeper struct {
	projectType string
	delete      func(context.Context, *serverless.ClientWithResponses, string) error
}

type serverlessProject struct {
	id        string
	name      string
	createdAt time.Time
}

func testSweepServerlessProjects(_ string) error {
	client, err := newServerlessAPI()
	if err != nil {
		return err
	}

	withResponses, ok := client.(*serverless.ClientWithResponses)
	if !ok {
		return fmt.Errorf("unexpected serverless API client type %T", client)
	}

	sweepers := []serverlessProjectSweeper{
		{projectType: "elasticsearch", delete: deleteElasticsearchProject},
		{projectType: "observability", delete: deleteObservabilityProject},
		{projectType: "security", delete: deleteSecurityProject},
	}

	var (
		merr = multierror.NewPrefixed("failed sweeping serverless projects")
		mu   sync.Mutex
	)

	for _, sweeper := range sweepers {
		projects, err := listServerlessProjects(context.Background(), sweeper.projectType)
		if err != nil {
			mu.Lock()
			merr = merr.Append(err)
			mu.Unlock()
			continue
		}

		var wg sync.WaitGroup
		for _, project := range projects {
			if !strings.HasPrefix(project.name, prefix) {
				continue
			}
			if !staleServerlessProject(project.createdAt) {
				continue
			}

			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				log.Printf("[DEBUG] Deleting serverless project %s", id)
				if err := sweeper.delete(context.Background(), withResponses, id); err != nil {
					mu.Lock()
					merr = merr.Append(err)
					mu.Unlock()
				}
			}(project.id)
		}
		wg.Wait()
	}

	return merr.ErrorOrNil()
}

func listServerlessProjects(ctx context.Context, projectType string) ([]serverlessProject, error) {
	cfg, err := newAPIConfig()
	if err != nil {
		return nil, err
	}

	baseURL := strings.TrimRight(cfg.Host, "/") + "/api/v1/serverless/projects/" + projectType
	projects := make([]serverlessProject, 0)
	nextPage := ""

	for {
		reqURL := baseURL
		if nextPage != "" {
			reqURL += "?next_page=" + url.QueryEscape(nextPage)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, err
		}
		cfg.AuthWriter.AuthRequest(req)

		resp, err := cfg.Client.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected response listing %s projects: %s %s", projectType, resp.Status, body)
		}

		var page struct {
			Items []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Metadata struct {
					CreatedAt time.Time `json:"created_at"`
				} `json:"metadata"`
			} `json:"items"`
			NextPage *string `json:"next_page"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}

		for _, item := range page.Items {
			projects = append(projects, serverlessProject{
				id:        item.ID,
				name:      item.Name,
				createdAt: item.Metadata.CreatedAt,
			})
		}

		if page.NextPage == nil || *page.NextPage == "" {
			break
		}
		nextPage = *page.NextPage
	}

	return projects, nil
}

// staleServerlessProject uses created_at because ProjectMetadata does not
// expose last_modified (unlike deployment sweepers).
func staleServerlessProject(createdAt time.Time) bool {
	return createdAt.Before(time.Now().Add(-serverlessProjectStaleAfter))
}

func deleteElasticsearchProject(ctx context.Context, client *serverless.ClientWithResponses, id string) error {
	res, err := client.DeleteElasticsearchProjectWithResponse(ctx, id, nil)
	return deleteServerlessProjectResponse(res, err)
}

func deleteObservabilityProject(ctx context.Context, client *serverless.ClientWithResponses, id string) error {
	res, err := client.DeleteObservabilityProjectWithResponse(ctx, id, nil)
	return deleteServerlessProjectResponse(res, err)
}

func deleteSecurityProject(ctx context.Context, client *serverless.ClientWithResponses, id string) error {
	res, err := client.DeleteSecurityProjectWithResponse(ctx, id, nil)
	return deleteServerlessProjectResponse(res, err)
}

type serverlessDeleteResponse interface {
	StatusCode() int
}

func deleteServerlessProjectResponse(res serverlessDeleteResponse, err error) error {
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK && res.StatusCode() != http.StatusNotFound {
		return fmt.Errorf("unexpected status deleting serverless project: %d", res.StatusCode())
	}
	return nil
}

func assertServerlessProjectDeleted(
	get func(context.Context, string) (bool, error),
	deleteFn func(context.Context, string) error,
	resourceType string,
	id string,
) error {
	ctx := context.Background()

	exists, err := get(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	if err := deleteFn(ctx, id); err != nil {
		return fmt.Errorf("failed to delete %s [%s]: %w", resourceType, id, err)
	}

	exists, err = get(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s [%s] still exists", resourceType, id)
	}

	return nil
}
