package main

import (
	"net/http"
	"strings"
)

type Server struct {
	store  *Store
	router *http.ServeMux
}

func NewServer(store *Store) *Server {
	s := &Server{store: store}
	s.routes()
	return s
}

func (s *Server) routes() {
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Page routes (HTML)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.handlePageProductList(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/new", s.handlePageNewProduct)

	// API routes for /products
	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleListProducts(w, r)
		case http.MethodPost:
			s.handleCreateProduct(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/products/")

		// Handle /products/:id/purchase
		if strings.HasSuffix(path, "/purchase") {
			if r.Method == http.MethodPost {
				s.handlePurchaseProduct(w, r)
				return
			}
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check Accept header to decide: HTML page or JSON API
		accept := r.Header.Get("Accept")
		wantsHTML := strings.Contains(accept, "text/html")

		switch r.Method {
		case http.MethodGet:
			if wantsHTML {
				s.handlePageProductDetail(w, r)
			} else {
				s.handleGetProduct(w, r)
			}
		case http.MethodPut:
			s.handleUpdateProduct(w, r)
		case http.MethodDelete:
			s.handleDeleteProduct(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	s.router = mux
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
