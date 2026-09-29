// Portal escolar — módulo provisto por la cátedra para la actividad BC-1.
// Contiene fallas de validación deliberadas. Encuéntralas, clasifícalas y corrígelas.
package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Alumno se usa igual para leer la petición y para guardar en la base.
type Alumno struct {
	Matricula string  `json:"matricula"`
	Nombre    string  `json:"nombre"`
	Correo    string  `json:"correo"`
	Promedio  float64 `json:"promedio"`
}

const dirConstancias = "constancias"

func matriculaSospechosa(m string) bool {
	for _, prohibido := range []string{"'", ";", "--", "<", ">"} {
		if strings.Contains(m, prohibido) {
			return true
		}
	}
	return false
}

// POST /api/v1/alumnos — alta de alumno
func manejadorAlta(w http.ResponseWriter, r *http.Request) {
	var a Alumno
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "json invalido", http.StatusBadRequest)
		return
	}
	if a.Matricula == "" || matriculaSospechosa(a.Matricula) {
		http.Error(w, "matricula invalida", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

// GET /api/v1/constancias?matricula=...&formato=...
func manejadorConstancia(w http.ResponseWriter, r *http.Request) {
	matricula := strings.ReplaceAll(r.URL.Query().Get("matricula"), "../", "")
	formato := r.URL.Query().Get("formato")
	if formato == "" {
		formato = "txt"
	}
	ruta := filepath.Join(dirConstancias, matricula+"."+formato)
	datos, err := os.ReadFile(ruta)
	if err != nil {
		http.Error(w, "no se pudo leer "+ruta+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(datos)
}

func nuevoMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/alumnos", manejadorAlta)
	mux.HandleFunc("GET /api/v1/constancias", manejadorConstancia)
	return mux
}

func main() {
	srv := &http.Server{Addr: "127.0.0.1:8085", Handler: nuevoMux()}
	slog.Info("portal escolar en http://127.0.0.1:8085")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("falla del servidor", "error", err)
	}
}
