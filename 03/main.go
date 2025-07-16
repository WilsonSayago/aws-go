package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Estructuras para Bedrock Claude
type ClaudeRequest struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
}

type ClaudeResponse struct {
	Completion string `json:"completion"`
}

// Estructura para el sistema RAG
type RAGSystem struct {
	s3Client      *s3.Client
	bedrockClient *bedrockruntime.Client
	bucketName    string
	chunks        []string
	modelID       string
}

// Configuración
type Config struct {
	Region     string
	BucketName string
	ModelID    string
}

func main() {
	// Flags de configuración
	region := flag.String("region", "us-east-1", "AWS region")
	bucket := flag.String("bucket", "", "S3 bucket name (required)")
	s3Key := flag.String("s3key", "", "S3 key/path to the document to process")
	query := flag.String("query", "", "Question to ask about the document")
	flag.Parse()

	if *bucket == "" {
		fmt.Println("Error: bucket name is required")
		flag.Usage()
		os.Exit(1)
	}

	config := &Config{
		Region:     *region,
		BucketName: *bucket,
		ModelID:    "anthropic.claude-v2",
	}

	// Inicializar sistema RAG
	rag, err := NewRAGSystem(config)
	if err != nil {
		log.Fatal("Error initializing RAG system:", err)
	}

	// Modo interactivo si no se proporcionan argumentos
	if *s3Key == "" && *query == "" {
		runInteractiveMode(rag)
		return
	}

	// Procesar archivo de S3 si se proporciona
	if *s3Key != "" {
		fmt.Printf("📄 Processing document from S3: s3://%s/%s\n", *bucket, *s3Key)
		err := rag.ProcessS3Document(*s3Key)
		if err != nil {
			log.Fatal("Error processing S3 document:", err)
		}
		fmt.Println("✅ Document processed successfully!")
	}

	// Responder consulta si se proporciona
	if *query != "" {
		fmt.Printf("🤖 Processing query: %s\n", *query)
		response, err := rag.Query(*query)
		if err != nil {
			log.Fatal("Error processing query:", err)
		}
		fmt.Printf("\n📝 Response:\n%s\n", response)
	}
}

// NewRAGSystem crea una nueva instancia del sistema RAG
func NewRAGSystem(cfg *Config) (*RAGSystem, error) {
	ctx := context.Background()

	// Cargar configuración de AWS
	sdkConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("couldn't load AWS config: %w", err)
	}

	return &RAGSystem{
		s3Client:      s3.NewFromConfig(sdkConfig),
		bedrockClient: bedrockruntime.NewFromConfig(sdkConfig),
		bucketName:    cfg.BucketName,
		chunks:        make([]string, 0),
		modelID:       cfg.ModelID,
	}, nil
}

// ProcessS3Document lee y procesa un documento existente de S3
func (r *RAGSystem) ProcessS3Document(s3Key string) error {
	ctx := context.Background()

	// Descargar documento de S3
	result, err := r.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("error downloading from S3: %w", err)
	}
	defer result.Body.Close()

	// Leer contenido
	content, err := io.ReadAll(result.Body)
	if err != nil {
		return fmt.Errorf("error reading content: %w", err)
	}

	fmt.Printf("📥 Document downloaded from S3: %d bytes\n", len(content))

	// Dividir en chunks
	r.chunks = r.splitIntoChunks(string(content), 1000)
	fmt.Printf("📊 Document split into %d chunks\n", len(r.chunks))

	return nil
}

// splitIntoChunks divide el texto en fragmentos de tamaño específico
func (r *RAGSystem) splitIntoChunks(text string, chunkSize int) []string {
	lines := strings.Split(text, "\n")
	var chunks []string
	var currentChunk strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Si agregar esta línea excede el tamaño Y ya tenemos contenido
		if currentChunk.Len()+len(line) > chunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			currentChunk.Reset()
		}

		currentChunk.WriteString(line + "\n")
	}

	// Agregar último chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	return chunks
}

// findRelevantChunks busca chunks relevantes basándose en palabras clave
func (r *RAGSystem) findRelevantChunks(question string, maxChunks int) []string {
	questionWords := strings.Fields(strings.ToLower(question))

	type chunkScore struct {
		chunk string
		score int
	}

	var scores []chunkScore

	for _, chunk := range r.chunks {
		chunkLower := strings.ToLower(chunk)
		score := 0

		for _, word := range questionWords {
			// Permitir palabras de 2 caracteres si son relevantes (como "Go")
			if len(word) > 2 || (len(word) == 2 && (word == "go" || word == "js" || word == "c+")) {
				score += strings.Count(chunkLower, word)
			}
		}

		scores = append(scores, chunkScore{chunk: chunk, score: score})
	}

	// Ordenar por score (simple bubble sort para este ejemplo)
	for i := 0; i < len(scores)-1; i++ {
		for j := 0; j < len(scores)-i-1; j++ {
			if scores[j].score < scores[j+1].score {
				scores[j], scores[j+1] = scores[j+1], scores[j]
			}
		}
	}

	// Seleccionar los mejores chunks
	var relevantChunks []string
	for i := 0; i < len(scores) && i < maxChunks; i++ {
		if scores[i].score > 0 {
			relevantChunks = append(relevantChunks, scores[i].chunk)
		}
	}

	// Si no hay chunks relevantes, usar los primeros chunks
	if len(relevantChunks) == 0 && len(r.chunks) > 0 {
		limit := maxChunks
		if len(r.chunks) < limit {
			limit = len(r.chunks)
		}
		relevantChunks = r.chunks[:limit]
	}

	return relevantChunks
}

// Query procesa una consulta usando RAG
func (r *RAGSystem) Query(question string) (string, error) {
	if len(r.chunks) == 0 {
		return "", fmt.Errorf("no document has been processed yet")
	}

	// Buscar chunks relevantes
	relevantChunks := r.findRelevantChunks(question, 3)

	// Log de chunks seleccionados
	fmt.Printf("\n📋 Chunks seleccionados para la pregunta '%s':\n", question)
	fmt.Printf("📊 Total de chunks relevantes: %d\n", len(relevantChunks))
	for i, chunk := range relevantChunks {
		fmt.Printf("\n--- Chunk %d ---\n", i+1)
		fmt.Printf("%s\n", chunk)
		fmt.Printf("--- Fin Chunk %d ---\n", i+1)
	}
	fmt.Println()

	// Construir contexto
	documentContext := strings.Join(relevantChunks, "\n\n")

	// Crear prompt para Claude
	prompt := fmt.Sprintf(`Human: Basándote en el siguiente contexto, responde la pregunta de manera precisa y detallada. \n\n CONTEXTO: %s \n\n PREGUNTA: %s \n\n RESPUESTA:`, documentContext, question)

	// Responde únicamente basándote en la información proporcionada en el contexto. Si la información no está disponible en el contexto, indica que no puedes responder basándote en el documento proporcionado.
	prompt += "\n\n Responde únicamente basándote en la información proporcionada en el contexto. Si la información no está disponible en el contexto, indica que no puedes responder basándote en el documento proporcionado."

	// Crear request para Claude
	claudeRequest := ClaudeRequest{
		Prompt:            prompt + "\n\nAssistant:",
		MaxTokensToSample: 500,
		Temperature:       0.7,
		StopSequences:     []string{"\n\nHuman:"},
	}

	// Serializar request
	body, err := json.Marshal(claudeRequest)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %w", err)
	}

	// Invocar modelo Claude
	ctx := context.Background()
	response, err := r.bedrockClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(r.modelID),
		ContentType: aws.String("application/json"),
		Body:        body,
	})
	if err != nil {
		return "", fmt.Errorf("error invoking Claude: %w", err)
	}

	// Deserializar respuesta
	var claudeResponse ClaudeResponse
	err = json.Unmarshal(response.Body, &claudeResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response: %w", err)
	}

	return claudeResponse.Completion, nil
}

// runInteractiveMode ejecuta el modo interactivo
func runInteractiveMode(rag *RAGSystem) {
	fmt.Println("🚀 Modo interactivo RAG - AWS Bedrock + S3")
	fmt.Println("Comandos disponibles:")
	fmt.Println("  /load <s3-key>  - Cargar documento desde S3")
	fmt.Println("  /status         - Ver estado del sistema")
	fmt.Println("  /quit           - Salir")
	fmt.Println("O simplemente escribe tu pregunta sobre el documento cargado.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("RAG> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Procesar comandos
		if strings.HasPrefix(input, "/") {
			parts := strings.Fields(input)
			command := parts[0]

			switch command {
			case "/quit":
				fmt.Println("👋 ¡Hasta luego!")
				return
			case "/status":
				if len(rag.chunks) == 0 {
					fmt.Println("❌ No hay documento cargado")
				} else {
					fmt.Printf("✅ Documento cargado: %d chunks\n", len(rag.chunks))
				}
			case "/load":
				if len(parts) < 2 {
					fmt.Println("❌ Uso: /load <s3-key>")
					continue
				}
				s3Key := parts[1]
				fmt.Printf("📄 Cargando documento: %s\n", s3Key)
				err := rag.ProcessS3Document(s3Key)
				if err != nil {
					fmt.Printf("❌ Error cargando documento: %v\n", err)
				} else {
					fmt.Println("✅ Documento cargado exitosamente!")
				}
			default:
				fmt.Printf("❌ Comando desconocido: %s\n", command)
			}
			continue
		}

		// Procesar pregunta
		if len(rag.chunks) == 0 {
			fmt.Println("❌ Primero debes cargar un documento con /load <s3-key>")
			continue
		}

		fmt.Println("🤖 Procesando pregunta...")
		response, err := rag.Query(input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Printf("\n📝 Respuesta:\n%s\n\n", response)
		}
	}
}
