package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type GitHubRESTClient struct {
	BaseURL string
	HTTPClient *http.Client
}

func NewGitHubRESTClient() *GitHubRESTClient {
	return &GitHubRESTClient{
		BaseURL: "https://api.github.com",
		HTTPClient: &http.Client{},
	}
}

func (c *GitHubRESTClient) DoRequest(method, path, token string, body []byte) ([]byte, error) {
	var reqBody *strings.Reader
	if body != nil {
		reqBody = strings.NewReader(string(body))
	}

	var req *http.Request
	var err error
	if reqBody != nil {
		req, err = http.NewRequest(method, fmt.Sprintf("%s%s", c.BaseURL, path), reqBody)
	} else {
		req, err = http.NewRequest(method, fmt.Sprintf("%s%s", c.BaseURL, path), nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer " + token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return respBody, nil
}

func (c *GitHubRESTClient) ListProjectItems(org string, projectNumber int, token string) (*ProjectsV2ItemsResponse, error) {
	path := fmt.Sprintf("/orgs/%s/projectsV2/%d/items", org, projectNumber)
	body, err := c.DoRequest(http.MethodGet, path, token, nil)
	if err != nil {
		return nil, err
	}

	var projectItemsResponse ProjectsV2ItemsResponse
	err = json.Unmarshal(body, &projectItemsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal project items response: %w", err)
	}

	return &projectItemsResponse, nil
}

func (c *GitHubRESTClient) ListProjects(org string, token string) (*ProjectsV2Response, error) {
	path := fmt.Sprintf("/orgs/%s/projectsV2", org)
	body, err := c.DoRequest(http.MethodGet, path, token, nil)
	if err != nil {
		return nil, err
	}

	var projectsResponse ProjectsV2Response
	err = json.Unmarshal(body, &projectsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal projects response: %w", err)
	}

	return &projectsResponse, nil
}

func (c *GitHubRESTClient) AddIssueToProject(org string, projectNumber int, issueNodeID string, token string) (*AddItemResponse, error) {
	path := fmt.Sprintf("/orgs/%s/projectsV2/%d/items", org, projectNumber)

	reqBody := AddItemRequest{
		ContentID: issueNodeID,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	respBody, err := c.DoRequest(http.MethodPost, path, token, bodyBytes)
	if err != nil {
		return nil, err
	}

	var addItemResponse AddItemResponse
	err = json.Unmarshal(respBody, &addItemResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal add item response: %w", err)
	}

	return &addItemResponse, nil
}
