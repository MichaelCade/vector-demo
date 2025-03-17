package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/tmc/langchaingo/vectorstores"
)

// --- Types ---
type ChatRequest struct {
	Query string `json:"query"`
}

type ChatResponse struct {
	Answer string `json:"answer"`
}

var store vectorstores.VectorStore

// --- Main ---
func main() {
	ctx := context.Background()

	// 1. Create embedder
	embedder, err := createEmbedder()
	if err != nil {
		panic(fmt.Sprintf("Failed to create embedder: %v", err))
	}

	// 2. Connect to PGVector
	store, err = connectToPGVector(ctx, embedder)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to PGVector: %v", err))
	}

	// 3. Load markdown files
	docs, err := loadMarkdownFiles("./markdowns")
	if err != nil {
		panic(fmt.Sprintf("Failed to load markdown files: %v", err))
	}

	// 4. Chunk and embed documents
	startEmbedding := time.Now()
	var allChunks []string
	for _, doc := range docs {
		chunks := chunkDocument(doc, 200)
		allChunks = append(allChunks, chunks...)
	}
	fmt.Printf("⏱️ Time to chunk: %s\n", time.Since(startEmbedding))
	printSystemStats()
	printGPUStats()

	// 5. Store embeddings
	startStore := time.Now()
	err = storeDocuments(ctx, store, allChunks)
	if err != nil {
		panic(fmt.Sprintf("Failed to store documents: %v", err))
	}
	fmt.Printf("⏱️ Time to store embeddings: %s\n", time.Since(startStore))
	printSystemStats()
	printGPUStats()

	fmt.Println("✅ Successfully stored markdown embeddings!")

	// 6. Serve chat API
	http.HandleFunc("/chat", chatHandler)
	fmt.Println("🚀 Chat API running at http://localhost:8080/chat")
	http.ListenAndServe(":8080", nil)
}

// --- Chat Handler ---
func chatHandler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK) // Preflight OK
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode query
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Vector search
	chunks, err := queryVectorStore(context.Background(), store, req.Query)
	if err != nil {
		http.Error(w, "Vector search failed", http.StatusInternalServerError)
		return
	}

	fmt.Println("🔍 Retrieved chunks from vector store (API call):")
for i, chunk := range chunks {
    fmt.Printf("[%d] %s\n", i+1, chunk)
}

	// Combine context and query
	contextText := combineChunks(chunks)

	// LLM Call
	answer, err := callOllamaLLM(contextText, req.Query)
	if err != nil {
		http.Error(w, "LLM call failed", http.StatusInternalServerError)
		return
	}

	// Send response
	resp := ChatResponse{Answer: answer}
	json.NewEncoder(w).Encode(resp)
}

// --- Ollama (Mistral) Call ---
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func callOllamaLLM(contextText, userQuery string) (string, error) {
	url := "http://localhost:11434/api/generate"

	prompt := fmt.Sprintf("Use this context to answer the question:\n\nContext:\n%s\n\nQuestion: %s", contextText, userQuery)

	reqBody := OllamaRequest{
		Model:  "mistral",
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("Failed to marshal Ollama request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Failed to call Ollama: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read Ollama response: %v", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("Failed to parse Ollama response: %v", err)
	}

	return ollamaResp.Response, nil
}

// --- Helper: Combine Chunks ---
func combineChunks(chunks []string) string {
	return joinChunks(chunks, "\n\n")
}

func joinChunks(chunks []string, sep string) string {
	var buffer bytes.Buffer
	for i, chunk := range chunks {
		buffer.WriteString(chunk)
		if i < len(chunks)-1 {
			buffer.WriteString(sep)
		}
	}
	return buffer.String()
}

// --- System Stats ---
func printSystemStats() {
	cpuPercent, _ := cpu.Percent(0, false)
	vmStat, _ := mem.VirtualMemory()
	fmt.Printf("💻 CPU Usage: %.2f%%\n", cpuPercent[0])
	fmt.Printf("💾 Memory Usage: %.2f%% (Used: %v MB, Total: %v MB)\n",
		vmStat.UsedPercent,
		vmStat.Used/1024/1024,
		vmStat.Total/1024/1024,
	)
}

func printGPUStats() {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
	if err := cmd.Run(); err != nil {
		fmt.Println("⚠️ No NVIDIA GPU detected or 'nvidia-smi' not available.")
		return
	}

	out, err := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu,memory.used,memory.total", "--format=csv,noheader,nounits").Output()
	if err != nil {
		fmt.Println("❌ Error fetching GPU stats:", err)
		return
	}
	fmt.Println("🖥️ GPU Stats (Util %, Mem Used MB, Mem Total MB):")
	fmt.Println(string(out))
}
