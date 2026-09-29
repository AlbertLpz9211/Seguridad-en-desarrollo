package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRegistro(t *testing.T) {
	raiz, err := os.OpenRoot("../almacen")
	if err != nil {
		t.Fatal(err)
	}
	defer raiz.Close()
	mux := nuevoMux(raiz)

	casos := []struct {
		nombre, cuerpo string
		esperado       int
	}{
		{"valido", `{"nombre_usuario":"ana26","correo":"Ana@UNG.mx","edad":20,"carrera":"software"}`, 201},
		{"carrera fuera de lista", `{"nombre_usuario":"ana26","correo":"ana@ung.mx","edad":20,"carrera":"medicina"}`, 422},
		{"campo extra (asignacion masiva)", `{"nombre_usuario":"ana26","correo":"ana@ung.mx","edad":20,"carrera":"software","rol":"docente"}`, 400},
		{"edad como texto", `{"nombre_usuario":"ana26","correo":"ana@ung.mx","edad":"veinte","carrera":"software"}`, 400},
		{"menor de edad", `{"nombre_usuario":"ana26","correo":"ana@ung.mx","edad":15,"carrera":"software"}`, 422},
		{"nombre reservado en ancho completo", `{"nombre_usuario":"ａｄｍｉｎ","correo":"x@ung.mx","edad":20,"carrera":"datos"}`, 422},
		{"dos objetos pegados", `{"nombre_usuario":"ana26","correo":"ana@ung.mx","edad":20,"carrera":"software"}{}`, 400},
		{"cuerpo de 20 KiB", `{"nombre_usuario":"` + strings.Repeat("a", 20000) + `"}`, 413},
	}
	for _, c := range casos {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/usuarios", strings.NewReader(c.cuerpo))
		mux.ServeHTTP(rr, req)
		fmt.Printf("%-36s -> %d %s", c.nombre, rr.Code, rr.Body.String())
		if rr.Code != c.esperado {
			t.Errorf("%s: esperado %d, obtenido %d", c.nombre, c.esperado, rr.Code)
		}
	}
}

func TestExpedientes(t *testing.T) {
	raiz, err := os.OpenRoot("../almacen")
	if err != nil {
		t.Fatal(err)
	}
	defer raiz.Close()
	mux := nuevoMux(raiz)

	casos := []struct {
		archivo  string
		esperado int
	}{
		{"expediente_100.txt", http.StatusOK},
		{"expediente_999.txt", http.StatusNotFound},
		{"../secreto.txt", http.StatusBadRequest},
		{"Expediente_100.TXT", http.StatusBadRequest},
	}
	for _, c := range casos {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/expedientes?archivo="+c.archivo, nil)
		mux.ServeHTTP(rr, req)
		fmt.Printf("%-22q -> %d %s", c.archivo, rr.Code, rr.Body.String())
		if rr.Code != c.esperado {
			t.Errorf("%s: esperado %d, obtenido %d", c.archivo, c.esperado, rr.Code)
		}
	}
}
