package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/gin-gonic/gin"
)

// Each model provider defines their own individual request and response formats.
// For the format, ranges, and default values for the different models, refer to:
// https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters.html

type ClaudeRequest struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
}

type ClaudeResponse struct {
	Completion string `json:"completion"`
}

// Request structure for the API endpoint
type PromptRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}

// Response structure for the API endpoint
type PromptResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// Global variables for AWS client
var bedrockClient *bedrockruntime.Client
var awsRegion string

// Initialize AWS Bedrock client
func initializeAWSClient(region string) error {
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("couldn't load default configuration: %v", err)
	}

	bedrockClient = bedrockruntime.NewFromConfig(sdkConfig)
	awsRegion = region
	return nil
}

// Function to invoke Claude model
func invokeClaude(prompt string, temperature float64, maxTokens int) (*ClaudeResponse, error) {
	ctx := context.Background()
	modelId := "anthropic.claude-v2"

	// Anthropic Claude requires you to enclose the prompt as follows:
	prefix := "Human: "
	postfix := "\n\nAssistant:"
	wrappedPrompt := prefix + prompt + postfix

	// Set default values if not provided
	if temperature == 0 {
		temperature = 0.5
	}
	if maxTokens == 0 {
		maxTokens = 100
	}

	request := ClaudeRequest{
		Prompt:            wrappedPrompt,
		MaxTokensToSample: maxTokens,
		Temperature:       temperature,
		StopSequences:     []string{"\n\nHuman:"},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal the request: %v", err)
	}

	result, err := bedrockClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelId),
		ContentType: aws.String("application/json"),
		Body:        body,
	})

	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "no such host") {
			return nil, fmt.Errorf("the Bedrock service is not available in the selected region. Please double-check the service availability for your region")
		} else if strings.Contains(errMsg, "Could not resolve the foundation model") {
			return nil, fmt.Errorf("could not resolve the foundation model from model identifier: %v. Please verify that the requested model exists and is accessible within the specified region", modelId)
		} else {
			return nil, fmt.Errorf("couldn't invoke Anthropic Claude: %v", err)
		}
	}

	var response ClaudeResponse
	err = json.Unmarshal(result.Body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &response, nil
}

// Handler for the POST /prompt endpoint
func handlePrompt(c *gin.Context) {
	var req PromptRequest

	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, PromptResponse{
			Error: fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	// Validate prompt is not empty
	if strings.TrimSpace(req.Prompt) == "" {
		c.JSON(http.StatusBadRequest, PromptResponse{
			Error: "Prompt cannot be empty",
		})
		return
	}

	// Invoke Claude model
	response, err := invokeClaude(req.Prompt, 0.5, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PromptResponse{
			Error: err.Error(),
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, PromptResponse{
		Response: response.Completion,
	})
}

// Health check endpoint
func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"region":  awsRegion,
		"service": "bedrock-claude-api",
	})
}

func main() {
	region := flag.String("region", "us-east-1", "The AWS region")
	port := flag.String("port", "8080", "The port to run the server on")
	flag.Parse()

	fmt.Printf("Initializing AWS Bedrock client for region: %s\n", *region)

	// Initialize AWS client
	if err := initializeAWSClient(*region); err != nil {
		log.Fatalf("Failed to initialize AWS client: %v", err)
	}

	// Create Gin router
	r := gin.Default()

	// Add middleware for CORS if needed
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Define routes
	r.GET("/health", handleHealth)
	r.POST("/prompt", handlePrompt)

	// Start server
	fmt.Printf("Starting server on port %s\n", *port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  GET  /health - Health check\n")
	fmt.Printf("  POST /prompt - Send prompt to Claude\n")
	fmt.Printf("\nExample usage:\n")
	fmt.Printf("curl -X POST http://localhost:%s/prompt \\\n", *port)
	fmt.Printf("  -H \"Content-Type: application/json\" \\\n")
	fmt.Printf("  -d '{\"prompt\": \"What is the capital of France?\"}'\n")

	if err := r.Run(":" + *port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
