package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Store struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateStoreRequest struct {
	Name string `json:"name"`
}

type AuthModel struct {
	SchemaVersion   string          `json:"schema_version"`
	TypeDefinitions json.RawMessage `json:"type_definitions"`
}

type WriteAuthModelResponse struct {
	AuthorizationModelID string `json:"authorization_model_id"`
}

type TupleKey struct {
	User     string `json:"user"`
	Relation string `json:"relation"`
	Object   string `json:"object"`
}

type WriteRequest struct {
	Writes []TupleKey `json:"writes"`
}

type CheckRequest struct {
	User     string `json:"user"`
	Relation string `json:"relation"`
	Object   string `json:"object"`
}

type CheckResponse struct {
	Allowed bool `json:"allowed"`
}

type ReadResponse struct {
	Tuples []struct {
		Key TupleKey `json:"key"`
	} `json:"tuples"`
}

type OpenFGAClient struct {
	baseURL              string
	httpClient          *http.Client
	storeID             string
	authorizationModelID string
}

func NewOpenFGAClient(baseURL string) *OpenFGAClient {
	return &OpenFGAClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *OpenFGAClient) doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		return json.Unmarshal(respBody, result)
	}

	return nil
}

func (c *OpenFGAClient) CreateStore(name string) (*Store, error) {
	req := CreateStoreRequest{Name: name}
	var store Store
	err := c.doRequest("POST", "/stores", req, &store)
	if err != nil {
		return nil, err
	}
	c.storeID = store.ID
	return &store, nil
}

func (c *OpenFGAClient) WriteAuthorizationModel(model *AuthModel) (*WriteAuthModelResponse, error) {
	path := fmt.Sprintf("/stores/%s/authorization-models", c.storeID)
	var resp WriteAuthModelResponse
	err := c.doRequest("POST", path, model, &resp)
	if err != nil {
		return nil, err
	}
	c.authorizationModelID = resp.AuthorizationModelID
	return &resp, nil
}

func (c *OpenFGAClient) Write(tuples []TupleKey) error {
	req := map[string]interface{}{
		"writes": []map[string]interface{}{
			{
				"tuple_key": map[string]string{
					"user":     tuples[0].User,
					"relation": tuples[0].Relation,
					"object":   tuples[0].Object,
				},
			},
		},
		"authorization_model_id": c.authorizationModelID,
	}
	path := fmt.Sprintf("/stores/%s/write", c.storeID)
	
	// Debug: print what we're sending
	reqJSON, _ := json.MarshalIndent(req, "", "  ")
	fmt.Printf("Sending to %s:\n%s\n", path, string(reqJSON))
	
	return c.doRequest("POST", path, req, nil)
}

func (c *OpenFGAClient) Check(user, relation, object string) (*CheckResponse, error) {
	req := CheckRequest{
		User:     user,
		Relation: relation,
		Object:   object,
	}
	path := fmt.Sprintf("/stores/%s/check", c.storeID)
	var resp CheckResponse
	err := c.doRequest("POST", path, req, &resp)
	return &resp, err
}

func (c *OpenFGAClient) Read() (*ReadResponse, error) {
	path := fmt.Sprintf("/stores/%s/read", c.storeID)
	var resp ReadResponse
	err := c.doRequest("POST", path, nil, &resp)
	return &resp, err
}

func main() {
	fmt.Println("🚀 Starting OpenFGA MongoDB Example")

	// Create OpenFGA client
	client := NewOpenFGAClient(getEnv("OPENFGA_API_URL", "http://localhost:8080"))

	// Step 1: Create a store
	fmt.Println("\n📁 Step 1: Creating OpenFGA store...")
	store, err := client.CreateStore("document-sharing-system")
	if err != nil {
		log.Fatalf("❌ Failed to create store: %v", err)
	}
	fmt.Printf("✅ Created store: %s (ID: %s)\n", store.Name, store.ID)

	// Step 2: Write authorization model
	fmt.Println("\n📝 Step 2: Writing authorization model...")
	typeDefinitions := json.RawMessage(`[
		{
			"type": "user"
		},
		{
			"type": "document",
			"relations": {
				"owner": {
					"union": {
						"child": [
							{
								"this": {}
							}
						]
					}
				}
			},
			"metadata": {
				"relations": {
					"owner": {
						"directly_related_user_types": [
							{
								"type": "user"
							}
						]
					}
				}
			}
		}
	]`)

	authModel := &AuthModel{
		SchemaVersion:   "1.1",
		TypeDefinitions: typeDefinitions,
	}

	writeModelResponse, err := client.WriteAuthorizationModel(authModel)
	if err != nil {
		log.Fatalf("❌ Failed to write authorization model: %v", err)
	}
	fmt.Printf("✅ Authorization model written (ID: %s)\n", writeModelResponse.AuthorizationModelID)

	// Step 3: Write relationship tuples
	fmt.Println("\n🔗 Step 3: Writing relationship tuples...")
	tuples := []TupleKey{
		{
			User:     "user:alice",
			Relation: "owner",
			Object:   "document:budget-2024",
		},
	}

	err = client.Write(tuples)
	if err != nil {
		log.Fatalf("❌ Failed to write tuples: %v", err)
	}
	fmt.Printf("✅ Written %d relationship tuples\n", len(tuples))

	// Step 4: Perform authorization checks
	fmt.Println("\n🔍 Step 4: Performing authorization checks...")

	// Test cases
	testCases := []struct {
		user     string
		relation string
		object   string
		expected bool
	}{
		{"user:alice", "owner", "document:budget-2024", true},   // Alice owns the document
		{"user:bob", "owner", "document:budget-2024", false},    // Bob is not owner
		{"user:charlie", "owner", "document:budget-2024", false}, // Charlie is not owner
		{"user:dave", "owner", "document:budget-2024", false},   // Dave has no access
	}

	allPassed := true
	for i, testCase := range testCases {
		checkResponse, err := client.Check(testCase.user, testCase.relation, testCase.object)
		if err != nil {
			log.Printf("❌ Check %d failed: %v", i+1, err)
			allPassed = false
			continue
		}

		passed := checkResponse.Allowed == testCase.expected
		if passed {
			fmt.Printf("✅ Check %d: %s can %s %s = %v\n", 
				i+1, testCase.user, testCase.relation, testCase.object, checkResponse.Allowed)
		} else {
			fmt.Printf("❌ Check %d: %s can %s %s = %v (expected %v)\n", 
				i+1, testCase.user, testCase.relation, testCase.object, checkResponse.Allowed, testCase.expected)
			allPassed = false
		}
	}

	// Step 5: Read stored tuples
	fmt.Println("\n📖 Step 5: Reading stored tuples...")
	readResponse, err := client.Read()
	if err != nil {
		log.Fatalf("❌ Failed to read tuples: %v", err)
	}
	fmt.Printf("✅ Found %d tuples in the store:\n", len(readResponse.Tuples))
	for _, tuple := range readResponse.Tuples {
		fmt.Printf("   • %s %s %s\n", tuple.Key.User, tuple.Key.Relation, tuple.Key.Object)
	}

	// Final result
	fmt.Println("\n🎉 Example completed!")
	if allPassed {
		fmt.Println("✅ All authorization checks passed!")
		fmt.Println("✅ MongoDB storage backend is working correctly")
		fmt.Printf("🌐 You can explore more at: http://localhost:3000/playground?storeId=%s\n", store.ID)
	} else {
		fmt.Println("❌ Some authorization checks failed")
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}