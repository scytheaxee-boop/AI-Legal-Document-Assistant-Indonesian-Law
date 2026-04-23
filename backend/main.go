package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/genai"
)

const (
	DB_URL     = "postgresql://postgres.gdupesltihkwiygmfdfe:Kit@brad_1320@aws-1-ap-south-1.pooler.supabase.com:5432/postgres"
	GEMINI_KEY = "AIzaSyABFBdKbwUfNCFF9rNcilzzo7CiV998onI"
)

type ChatRequest struct {
	Pertanyaan string `json:"pertanyaan"`
}

func main() {
	db, err := sql.Open("postgres", DB_URL)
	if err != nil {
		log.Fatal("Gagal koneksi awal ke database:", err)
	}
	defer db.Close()

	r := gin.Default()

	r.POST("/api/tanya", func(c *gin.Context) {
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Println("Error Binding JSON:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format request salah"})
			return
		}

		ctx := context.Background()
		client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: GEMINI_KEY})
		if err != nil {
			fmt.Println("Error Init AI:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal inisialisasi AI"})
			return
		}

		// 1. Mengubah Pertanyaan menjadi Vektor (Model Embedding Generasi Baru)
		resEmbed, err := client.Models.EmbedContent(ctx, "gemini-embedding-001", genai.Text(req.Pertanyaan), nil)
		if err != nil || len(resEmbed.Embeddings) == 0 {
			fmt.Println("Error Membuat Vektor (Embedding):", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat vektor pertanyaan"})
			return
		}
		queryVector := resEmbed.Embeddings[0].Values

		// --- BUG FIX: Memotong vektor pertanyaan menjadi 768 dimensi ---
		if len(queryVector) > 768 {
			queryVector = queryVector[:768]
		}

		// 2. Memaksa Golang memisahkan angka dengan KOMA (Bukan Spasi) untuk Supabase
		vectorBytes, _ := json.Marshal(queryVector)
		vectorStr := string(vectorBytes)

		// 3. Mencari Dokumen yang Cocok di Database (Menggunakan ::vector)
		rows, err := db.Query(`
			SELECT pasal, isi_teks 
			FROM dokumen_hukum 
			ORDER BY embedding <=> $1::vector 
			LIMIT 2`, vectorStr)

		if err != nil {
			fmt.Println("Error Database Query:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencari di database"})
			return
		}
		defer rows.Close()

		var contextHukum string
		for rows.Next() {
			var pasal, isi_teks string
			rows.Scan(&pasal, &isi_teks)
			contextHukum += fmt.Sprintf("- %s: %s\n", pasal, isi_teks)
		}

		// 4. Menyusun Prompt untuk AI
		promptFinal := fmt.Sprintf(`Anda adalah asisten hukum Indonesia yang cerdas. 
Jawab pertanyaan user HANYA berdasarkan konteks hukum berikut. Jika jawabannya tidak ada di konteks, bilang Anda tidak tahu.

Konteks Hukum:
%s

Pertanyaan User: %s`, contextHukum, req.Pertanyaan)

		// 5. Menghasilkan Jawaban dengan Model Stabil
		respGen, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(promptFinal), nil)
		if err != nil {
			fmt.Println("Error Generate Jawaban (Gemini):", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate jawaban"})
			return
		}

		jawabanAkhir := respGen.Text()

		// 6. Mengirim Balasan ke Frontend
		c.JSON(http.StatusOK, gin.H{
			"pertanyaan": req.Pertanyaan,
			"jawaban":    jawabanAkhir,
			"referensi":  contextHukum,
		})
	})

	fmt.Println("Server backend berjalan di http://localhost:8080")
	r.Run(":8080")
}
