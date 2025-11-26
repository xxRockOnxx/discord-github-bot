package rest

type ProjectsV2ItemsResponse []ProjectsV2Item

type ProjectsV2Item struct {
	ID         int      `json:"id"`
	NodeID     string   `json:"node_id"`
	ProjectURL string   `json:"project_url"`
	ContentType string   `json:"content_type"`
	Content    *ItemContent `json:"content"`
	Creator    SimpleUser `json:"creator"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
	ArchivedAt *string  `json:"archived_at"`
	ItemURL    *string  `json:"item_url"`
	Fields     []map[string]interface{} `json:"fields"`
}

type ItemContent struct {
	Title    string `json:"title"`
	Number   int    `json:"number"`
	Typename string `json:"__typename"` // GraphQL typename, often used to differentiate content types
}

type SimpleUser struct {
	Name      *string `json:"name"`
	Email     *string `json:"email"`
	Login     string  `json:"login"`
	ID        int     `json:"id"`
	NodeID    string  `json:"node_id"`
	AvatarURL string  `json:"avatar_url"`
	HTMLURL   string  `json:"html_url"`
}

// ProjectsV2Response represents the response from listing projects
type ProjectsV2Response struct {
	Projects []ProjectV2 `json:"projects"`
}

// ProjectV2 represents a GitHub Project v2
type ProjectV2 struct {
	ID          int    `json:"id"`
	NodeID      string `json:"node_id"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ShortDescription string `json:"short_description"`
	Public      bool   `json:"public"`
	Closed      bool   `json:"closed"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	HTMLURL     string `json:"html_url"`
}

// AddItemRequest represents the request body for adding an item to a project
type AddItemRequest struct {
	ContentID   string `json:"contentId"`
	ContentType string `json:"contentType,omitempty"`
}

// AddItemResponse represents the response from adding an item to a project
type AddItemResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	ContentID string `json:"contentId"`
}
