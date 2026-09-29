// Servidor de la S4: validación por esquema, límite de cuerpo y confinamiento de archivos.
package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/text/unicode/norm"
)

// PeticionRegistro es el DTO: solo los campos que el cliente puede enviar.
type PeticionRegistro struct {
	NombreUsuario string `json:"nombre_usuario" validate:"required,alphanum,min=3,max=30"`
	Correo        string `json:"correo" validate:"required,email,max=100"`
	Edad          int    `json:"edad" validate:"required,min=18,max=120"`
	Carrera       string `json:"carrera" validate:"required,oneof=software datos"`
}

const limiteCuerpo = 10 * 1024 // 10 KiB

var (
	validar        = validator.New()
	patronArchivo  = regexp.MustCompile(`^[a-z0-9_-]{1,40}\.txt$`)
	nombresVetados = map[string]bool{"admin": true, "root": true}
)

func responderJSON(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	json.NewEncoder(w).Encode(v)
}

func manejadorRegistro(w http.ResponseWriter, r *http.Request) {
	// 1. Control volumétrico
	r.Body = http.MaxBytesReader(w, r.Body, limiteCuerpo)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // 2. Esquema cerrado: nada de campos extra

	var p PeticionRegistro
	if err := dec.Decode(&p); err != nil {
		var demasiado *http.MaxBytesError
		if errors.As(err, &demasiado) {
			responderJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "cuerpo demasiado grande"})
			return
		}
		responderJSON(w, http.StatusBadRequest, map[string]string{"error": "json invalido o con campos no permitidos"})
		return
	}
	// Un solo objeto JSON: cualquier dato posterior se rechaza
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		responderJSON(w, http.StatusBadRequest, map[string]string{"error": "se esperaba un solo objeto json"})
		return
	}

	// 3. Canonicalizar ANTES de validar
	p.NombreUsuario = norm.NFKC.String(p.NombreUsuario)
	p.Correo = strings.ToLower(norm.NFKC.String(p.Correo))

	// 4. Validación sintáctica por lista de permitidos
	if err := validar.Struct(p); err != nil {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "datos fuera del esquema"})
		return
	}
	// 5. Validación semántica / de negocio
	if nombresVetados[strings.ToLower(p.NombreUsuario)] {
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "nombre de usuario reservado"})
		return
	}
	responderJSON(w, http.StatusCreated, map[string]any{"mensaje": "usuario registrado", "datos": p})
}

func nuevoManejadorExpedientes(raiz *os.Root) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nombre := r.URL.Query().Get("archivo")
		// Capa 1: lista de permitidos sobre la cadena
		if !patronArchivo.MatchString(nombre) {
			http.Error(w, "nombre de archivo no permitido", http.StatusBadRequest)
			return
		}
		// Capa 2: confinamiento en el sistema de archivos
		f, err := raiz.Open(nombre)
		if err != nil {
			http.Error(w, "recurso no encontrado", http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.Copy(w, f)
	}
}

func nuevoMux(raiz *os.Root) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/usuarios", manejadorRegistro)
	mux.HandleFunc("GET /api/v1/expedientes", nuevoManejadorExpedientes(raiz))
	return mux
}

func main() {
	raiz, err := os.OpenRoot("almacen")
	if err != nil {
		slog.Error("no se pudo abrir el almacen", "error", err)
		os.Exit(1)
	}
	defer raiz.Close()

	srv := &http.Server{Addr: "127.0.0.1:8084", Handler: nuevoMux(raiz), ReadHeaderTimeout: 5 * time.Second}
	slog.Info("servidor seguro en http://127.0.0.1:8084")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("falla del servidor", "error", err)
	}
}
