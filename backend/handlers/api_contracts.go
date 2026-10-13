package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"backend/db"
	"backend/middleware"
	"github.com/gin-gonic/gin"
)

func contractProjectForNode(nodeID string) (string, error) {
	var projectID string
	err := db.DB.QueryRow(`SELECT m.project_id FROM workflow_nodes n JOIN modules m ON m.id=n.module_id WHERE n.id=$1 AND n.deleted_at IS NULL`, nodeID).Scan(&projectID)
	return projectID, err
}

func contractProjectForRequest(requestID string) (string, error) {
	var projectID string
	err := db.DB.QueryRow(`SELECT m.project_id FROM api_requests r JOIN workflow_nodes n ON n.id=r.node_id JOIN modules m ON m.id=n.module_id WHERE r.id=$1`, requestID).Scan(&projectID)
	return projectID, err
}

func contractProjectForEnvironment(environmentID string) (string, error) {
	var projectID string
	err := db.DB.QueryRow(`SELECT project_id FROM environments WHERE id=$1`, environmentID).Scan(&projectID)
	return projectID, err
}

func contractNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "API contract resource not found"})
}
func contractBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message, "error_code": "validation_error"})
}

func authorizedContractProject(c *gin.Context, projectID string, capability middleware.ProjectCapability) bool {
	if projectID == "" {
		contractNotFound(c)
		return false
	}
	return middleware.AuthorizeProject(c, projectID, capability)
}

// Environment handlers return variable metadata only; encrypted values are never serialized.
func ListEnvironmentsHandler(c *gin.Context) {
	projectID := c.Param("id")
	if !authorizedContractProject(c, projectID, middleware.CapabilityView) {
		return
	}
	rows, err := db.DB.Query(`SELECT id,name,approved_base_url,is_default,runner_policy,created_at,updated_at FROM environments WHERE project_id=$1 ORDER BY is_default DESC,name`, projectID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list environments"})
		return
	}
	defer rows.Close()
	result := []gin.H{}
	for rows.Next() {
		var id, name, base, policy string
		var def bool
		var created, updated any
		if err := rows.Scan(&id, &name, &base, &def, &policy, &created, &updated); err != nil {
			c.JSON(500, gin.H{"error": "failed to parse environments"})
			return
		}
		result = append(result, gin.H{"id": id, "name": name, "approved_base_url": base, "is_default": def, "runner_policy": policy, "created_at": created, "updated_at": updated})
	}
	c.JSON(200, result)
}

func CreateEnvironmentHandler(c *gin.Context) {
	projectID := c.Param("id")
	if !authorizedContractProject(c, projectID, middleware.CapabilityManage) {
		return
	}
	var input struct {
		Name            string         `json:"name"`
		ApprovedBaseURL string         `json:"approved_base_url"`
		IsDefault       bool           `json:"is_default"`
		RunnerPolicy    map[string]any `json:"runner_policy"`
	}
	if c.ShouldBindJSON(&input) != nil {
		contractBadRequest(c, "invalid environment input")
		return
	}
	u, err := url.Parse(strings.TrimSpace(input.ApprovedBaseURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		contractBadRequest(c, "approved base URL must be an absolute HTTP/HTTPS URL without credentials")
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		contractBadRequest(c, "environment name is required")
		return
	}
	id := "env_" + GenerateUUID()
	_, err = db.DB.Exec(`INSERT INTO environments(id,project_id,name,approved_base_url,is_default,runner_policy) VALUES($1,$2,$3,$4,$5,COALESCE($6::jsonb,'{}'::jsonb))`, id, projectID, strings.TrimSpace(input.Name), u.String(), input.IsDefault, marshalJSON(input.RunnerPolicy))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create environment"})
		return
	}
	c.JSON(201, gin.H{"id": id})
}

func UpdateEnvironmentHandler(c *gin.Context) {
	projectID, err := contractProjectForEnvironment(c.Param("id"))
	if err != nil {
		contractNotFound(c)
		return
	}
	if !authorizedContractProject(c, projectID, middleware.CapabilityManage) {
		return
	}
	var input struct {
		Name            *string        `json:"name"`
		ApprovedBaseURL *string        `json:"approved_base_url"`
		IsDefault       *bool          `json:"is_default"`
		RunnerPolicy    map[string]any `json:"runner_policy"`
	}
	if c.ShouldBindJSON(&input) != nil {
		contractBadRequest(c, "invalid environment input")
		return
	}
	if input.ApprovedBaseURL != nil {
		u, e := url.Parse(*input.ApprovedBaseURL)
		if e != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
			contractBadRequest(c, "approved base URL must be an absolute HTTP/HTTPS URL without credentials")
			return
		}
	}
	_, err = db.DB.Exec(`UPDATE environments SET name=COALESCE(NULLIF($1,''),name),approved_base_url=COALESCE(NULLIF($2,''),approved_base_url),is_default=COALESCE($3,is_default),runner_policy=CASE WHEN $4::jsonb='null'::jsonb THEN runner_policy ELSE $4::jsonb END,updated_at=CURRENT_TIMESTAMP WHERE id=$5`, input.Name, input.ApprovedBaseURL, input.IsDefault, marshalJSON(input.RunnerPolicy), c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to update environment"})
		return
	}
	c.Status(204)
}

func DeleteEnvironmentHandler(c *gin.Context) {
	projectID, err := contractProjectForEnvironment(c.Param("id"))
	if err != nil {
		contractNotFound(c)
		return
	}
	if !authorizedContractProject(c, projectID, middleware.CapabilityManage) {
		return
	}
	if _, err = db.DB.Exec(`DELETE FROM environments WHERE id=$1`, c.Param("id")); err != nil {
		c.JSON(500, gin.H{"error": "failed to delete environment"})
		return
	}
	c.Status(204)
}

func runnerSecretKey() ([]byte, error) {
	key := os.Getenv("RUNNER_SECRET_KEY")
	if key == "" {
		return nil, fmt.Errorf("RUNNER_SECRET_KEY is required for secret variables")
	}
	sum := sha256.Sum256([]byte(key))
	return sum[:], nil
}
func encryptRunnerSecret(value string) ([]byte, error) {
	key, err := runnerSecretKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return []byte(base64.RawStdEncoding.EncodeToString(append(nonce, gcm.Seal(nil, nonce, []byte(value), nil)...))), nil
}

func ListEnvironmentVariablesHandler(c *gin.Context) {
	projectID, err := contractProjectForEnvironment(c.Param("id"))
	if err != nil {
		contractNotFound(c)
		return
	}
	if !authorizedContractProject(c, projectID, middleware.CapabilityView) {
		return
	}
	rows, err := db.DB.Query(`SELECT id,variable_key,is_secret,updated_by,updated_at FROM environment_variables WHERE environment_id=$1 ORDER BY variable_key`, c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list variables"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, key string
		var secret bool
		var updatedBy *string
		var updated any
		if rows.Scan(&id, &key, &secret, &updatedBy, &updated) != nil {
			c.JSON(500, gin.H{"error": "failed to parse variables"})
			return
		}
		out = append(out, gin.H{"id": id, "variable_key": key, "is_secret": secret, "updated_by": updatedBy, "updated_at": updated})
	}
	c.JSON(200, out)
}
func UpsertEnvironmentVariableHandler(c *gin.Context) {
	projectID, err := contractProjectForEnvironment(c.Param("id"))
	if err != nil {
		contractNotFound(c)
		return
	}
	if !authorizedContractProject(c, projectID, middleware.CapabilityManage) {
		return
	}
	var input struct {
		Key    string `json:"variable_key"`
		Value  string `json:"value"`
		Secret bool   `json:"is_secret"`
	}
	if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Key) == "" || input.Value == "" {
		contractBadRequest(c, "variable key and value are required")
		return
	}
	encrypted, err := encryptRunnerSecret(input.Value)
	if err != nil {
		c.JSON(500, gin.H{"error": "secret encryption is not configured"})
		return
	}
	userID, _ := middleware.GetUserID(c)
	id := "var_" + GenerateUUID()
	_, err = db.DB.Exec(`INSERT INTO environment_variables(id,environment_id,variable_key,encrypted_value,is_secret,updated_by) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(environment_id,variable_key) DO UPDATE SET encrypted_value=EXCLUDED.encrypted_value,is_secret=EXCLUDED.is_secret,updated_by=EXCLUDED.updated_by,updated_at=CURRENT_TIMESTAMP`, id, c.Param("id"), strings.TrimSpace(input.Key), encrypted, input.Secret, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to save variable"})
		return
	}
	c.JSON(201, gin.H{"variable_key": strings.TrimSpace(input.Key), "is_secret": input.Secret})
}
func DeleteEnvironmentVariableHandler(c *gin.Context) {
	envID := c.Param("id")
	if envID == "" {
		envID = c.Param("environmentId")
	}
	projectID, err := contractProjectForEnvironment(envID)
	if err != nil {
		contractNotFound(c)
		return
	}
	if !authorizedContractProject(c, projectID, middleware.CapabilityManage) {
		return
	}
	_, err = db.DB.Exec(`DELETE FROM environment_variables WHERE id=$1 AND environment_id=$2`, c.Param("variableId"), envID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to delete variable"})
		return
	}
	c.Status(204)
}

func marshalJSON(value any) string {
	if value == nil {
		return "null"
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "null"
	}
	return string(raw)
}
