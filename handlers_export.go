package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// handleExportCSV handles GET /products/export
func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	products, err := s.store.ListProducts(category)
	if err != nil {
		http.Error(w, "failed to list products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=products_%s.csv", time.Now().Format("20060102_150405")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header row
	header := []string{"id", "name", "description", "price", "category", "in_stock", "quantity", "created_at", "updated_at"}
	if err := writer.Write(header); err != nil {
		log.Printf("ERROR: csv header write: %v", err)
		return
	}

	for _, p := range products {
		apiProduct := toAPIProduct(&p)
		record := []string{
			strconv.Itoa(apiProduct.ID),
			apiProduct.Name,
			apiProduct.Description,
			fmt.Sprintf("%.2f", apiProduct.Price),
			apiProduct.Category,
			strconv.FormatBool(apiProduct.InStock),
			strconv.Itoa(apiProduct.Quantity),
			apiProduct.CreatedAt.Format(time.RFC3339),
			apiProduct.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			log.Printf("ERROR: csv record write: %v", err)
			return
		}
	}
}

// handleImportCSV handles POST /products/import
func (s *Server) handleImportCSV(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/csv") && !strings.HasPrefix(contentType, "multipart/form-data") {
		http.Error(w, "expected CSV content", http.StatusBadRequest)
		return
	}

	var reader io.Reader

	if strings.HasPrefix(contentType, "multipart/form-data") {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "failed to read uploaded file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		reader = file
	} else {
		reader = r.Body
	}

	csvReader := csv.NewReader(reader)

	// Read and validate header
	header, err := csvReader.Read()
	if err != nil {
		http.Error(w, "failed to read CSV header", http.StatusBadRequest)
		return
	}

	expectedHeader := []string{"name", "description", "price", "category", "in_stock", "quantity"}
	headerMap := make(map[string]int)
	for i, col := range header {
		headerMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	for _, expected := range expectedHeader {
		if _, ok := headerMap[expected]; !ok {
			http.Error(w, fmt.Sprintf("missing required column: %s", expected), http.StatusBadRequest)
			return
		}
	}

	var imported, skipped int
	var errors []string

	for lineNum := 2; ; lineNum++ {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors = append(errors, fmt.Sprintf("line %d: %v", lineNum, err))
			skipped++
			continue
		}

		name := strings.TrimSpace(record[headerMap["name"]])
		description := strings.TrimSpace(record[headerMap["description"]])
		priceStr := strings.TrimSpace(record[headerMap["price"]])
		category := strings.TrimSpace(record[headerMap["category"]])
		inStockStr := strings.TrimSpace(record[headerMap["in_stock"]])
		quantityStr := strings.TrimSpace(record[headerMap["quantity"]])

		if name == "" {
			errors = append(errors, fmt.Sprintf("line %d: name is required", lineNum))
			skipped++
			continue
		}

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			errors = append(errors, fmt.Sprintf("line %d: invalid price %q", lineNum, priceStr))
			skipped++
			continue
		}
		priceCents := int(math.Round(price * 100))

		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			errors = append(errors, fmt.Sprintf("line %d: invalid quantity %q", lineNum, quantityStr))
			skipped++
			continue
		}

		inStock := strings.EqualFold(inStockStr, "true") || inStockStr == "1"

		_, err = s.store.CreateProduct(name, description, priceCents, category, inStock, quantity)
		if err != nil {
			errors = append(errors, fmt.Sprintf("line %d: %v", lineNum, err))
			skipped++
			continue
		}

		imported++
	}

	result := map[string]interface{}{
		"imported": imported,
		"skipped":  skipped,
	}
	if len(errors) > 0 {
		result["errors"] = errors
	}

	w.Header().Set("Content-Type", "application/json")
	if skipped > 0 {
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(result)
}

// handleExportJSON handles GET /products/export/json
func (s *Server) handleExportJSON(w http.ResponseWriter, r *http.Request) {
	products, err := s.store.ListProducts("")
	if err != nil {
		http.Error(w, "failed to list products", http.StatusInternalServerError)
		return
	}

	apiProducts := make([]Product, len(products))
	for i, p := range products {
		apiProducts[i] = toAPIProduct(&p)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=products_%s.json", time.Now().Format("20060102_150405")))

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(apiProducts)
}
