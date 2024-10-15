package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/chrlesur/json-ld-converter/internal/logger"
)

const AIYOUAPIURL = "https://ai.dragonflygroup.fr/api"

type AIYOUClient struct {
    Token       string
    AssistantID string
    Timeout     time.Duration
    HTTPClient  *http.Client
}

func NewAIYOUClient(assistantID string, timeout time.Duration) *AIYOUClient {
    return &AIYOUClient{
        AssistantID: assistantID,
        Timeout:     timeout,
        HTTPClient:  &http.Client{Timeout: timeout},
    }
}

func (c *AIYOUClient) Login(email, password string) error {
    loginData := map[string]string{
        "email":    email,
        "password": password,
    }
    jsonData, err := json.Marshal(loginData)
    if err != nil {
        return fmt.Errorf("error marshaling login data: %w", err)
    }

    resp, err := c.makeAPICall("/login", "POST", jsonData)
    if err != nil {
        return fmt.Errorf("login error: %w", err)
    }

    var loginResp struct {
        Token string `json:"token"`
    }
    if err := json.Unmarshal(resp, &loginResp); err != nil {
        return fmt.Errorf("error unmarshaling login response: %w", err)
    }

    c.Token = loginResp.Token
    return nil
}

func (c *AIYOUClient) Analyze(ctx context.Context, content string, analysisContext *AnalysisContext) (string, *AnalysisContext, error) {
    logger.Debug("Starting analysis with AI.YOU API")
    prompt := BuildPromptWithContext(content, analysisContext)

    logger.Debug(fmt.Sprintf("Prepared prompt for AI.YOU API:\n%s", prompt))

    threadID, err := c.createThread()
    if err != nil {
        return "", nil, fmt.Errorf("error creating thread: %w", err)
    }

    if err := c.addMessage(threadID, prompt); err != nil {
        return "", nil, fmt.Errorf("error adding message to thread: %w", err)
    }

    runID, err := c.createRun(threadID)
    if err != nil {
        return "", nil, fmt.Errorf("error creating run: %w", err)
    }

    completedRun, err := c.waitForCompletion(threadID, runID)
    if err != nil {
        return "", nil, fmt.Errorf("error waiting for run completion: %w", err)
    }

    response, ok := (*completedRun)["response"].(string)
    if !ok {
        return "", nil, fmt.Errorf("response could not be extracted from the run")
    }

    // Mettre à jour le contexte d'analyse
    updatedContext, err := UpdateAnalysisContext(response, analysisContext)
    if err != nil {
        logger.Error(fmt.Sprintf("Error updating analysis context: %v", err))
        return "", nil, fmt.Errorf("error updating analysis context: %w", err)
    }

    logger.Debug(fmt.Sprintf("AI.YOU API response:\n%s", response))
    logger.Info(fmt.Sprintf("Analysis completed successfully (response length: %d characters)", len(response)))
    return response, updatedContext, nil
}

func (c *AIYOUClient) createThread() (string, error) {
    resp, err := c.makeAPICall("/v1/threads", "POST", []byte("{}"))
    if err != nil {
        return "", err
    }

    var threadResp struct {
        ID string `json:"id"`
    }
    if err := json.Unmarshal(resp, &threadResp); err != nil {
        return "", fmt.Errorf("error unmarshaling thread response: %w", err)
    }

    return threadResp.ID, nil
}

func (c *AIYOUClient) addMessage(threadID, content string) error {
    messageData := map[string]string{
        "role":    "user",
        "content": content,
    }
    jsonData, err := json.Marshal(messageData)
    if err != nil {
        return fmt.Errorf("error marshaling message data: %w", err)
    }

    _, err = c.makeAPICall(fmt.Sprintf("/v1/threads/%s/messages", threadID), "POST", jsonData)
    return err
}

func (c *AIYOUClient) createRun(threadID string) (string, error) {
    runData := map[string]string{
        "assistantId": c.AssistantID,
    }
    jsonData, err := json.Marshal(runData)
    if err != nil {
        return "", fmt.Errorf("error marshaling run data: %w", err)
    }

    resp, err := c.makeAPICall(fmt.Sprintf("/v1/threads/%s/runs", threadID), "POST", jsonData)
    if err != nil {
        return "", err
    }

    var runResp struct {
        ID string `json:"id"`
    }
    if err := json.Unmarshal(resp, &runResp); err != nil {
        return "", fmt.Errorf("error unmarshaling run response: %w", err)
    }

    return runResp.ID, nil
}

func (c *AIYOUClient) waitForCompletion(threadID, runID string) (*map[string]interface{}, error) {
    maxRetries := 30
    for i := 0; i < maxRetries; i++ {
        run, err := c.retrieveRun(threadID, runID)
        if err != nil {
            return nil, err
        }

        status, ok := run["status"].(string)
        if !ok {
            return nil, fmt.Errorf("run status not found or invalid")
        }

        if status == "completed" {
            return &run, nil
        }

        if status == "failed" || status == "cancelled" {
            return nil, fmt.Errorf("run failed with status: %s", status)
        }

        time.Sleep(2 * time.Second)
    }

    return nil, fmt.Errorf("timeout waiting for run completion")
}

func (c *AIYOUClient) retrieveRun(threadID, runID string) (map[string]interface{}, error) {
    resp, err := c.makeAPICall(fmt.Sprintf("/v1/threads/%s/runs/%s", threadID, runID), "GET", nil)
    if err != nil {
        return nil, err
    }

    var runStatus map[string]interface{}
    if err := json.Unmarshal(resp, &runStatus); err != nil {
        return nil, fmt.Errorf("error unmarshaling run status: %w", err)
    }

    return runStatus, nil
}

func (c *AIYOUClient) makeAPICall(endpoint, method string, data []byte) ([]byte, error) {
    maxRetries := 5
    baseDelay := 20 * time.Second

    var resp *http.Response
    var err error

    for attempt := 0; attempt < maxRetries; attempt++ {
        url := AIYOUAPIURL + endpoint
        req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
        if err != nil {
            return nil, fmt.Errorf("error creating HTTP request: %w", err)
        }

        req.Header.Set("Content-Type", "application/json")
        if c.Token != "" {
            req.Header.Set("Authorization", "Bearer "+c.Token)
        }

        logger.Info(fmt.Sprintf("Sending request to AI.YOU API (Attempt %d of %d)", attempt+1, maxRetries))
        resp, err = c.HTTPClient.Do(req)
        if err == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
            defer resp.Body.Close()
            body, err := ioutil.ReadAll(resp.Body)
            if err == nil {
                return body, nil
            }
        }

        logger.Warning(fmt.Sprintf("Attempt %d failed: %v", attempt+1, err))
        if attempt < maxRetries-1 {
            delay := baseDelay + time.Duration(attempt)*20*time.Second
            logger.Info(fmt.Sprintf("Retrying in %v", delay))
            time.Sleep(delay)
        }
    }

    return nil, fmt.Errorf("API error after %d attempts: %v", maxRetries, err)
}