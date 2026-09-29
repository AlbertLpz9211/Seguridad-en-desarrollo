package main

// Estas pruebas describen el comportamiento SEGURO esperado.
// Contra el código original fallan; tu corrección debe hacerlas pasar todas.

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func peticion(t *testing.T, metodo, url, cuerpo string) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	nuevoMux().ServeHTTP(rr, httptest.NewRequest(metodo, url, strings.NewReader(cuerpo)))
	return rr
}

func TestAltaValida(t *testing.T) {
	rr := peticion(t, "POST", "/api/v1/alumnos", `{"matricula":"2026A0001","nombre":"Ana Pérez","correo":"ana@ung.mx"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("alta valida: esperado 201, obtenido %d", rr.Code)
	}
}

func TestAltaRechazaPromedioDelCliente(t *testing.T) {
	rr := peticion(t, "POST", "/api/v1/alumnos", `{"matricula":"2026A0001","nombre":"Ana","correo":"ana@ung.mx","promedio":10}`)
	if rr.Code == http.StatusCreated {
		t.Fatalf("el cliente no debe poder fijar su promedio (CWE-915)")
	}
}

func TestAltaRechazaMatriculaFueraDeFormato(t *testing.T) {
	for _, m := range []string{"2026A0001 OR 1", "abc", "2026A00011111", "２０２６A0001"} {
		rr := peticion(t, "POST", "/api/v1/alumnos", `{"matricula":"`+m+`","nombre":"Ana","correo":"ana@ung.mx"}`)
		if rr.Code == http.StatusCreated {
			t.Errorf("matricula %q no cumple el formato AAAALNNNN y fue aceptada", m)
		}
	}
}

func TestAltaRechazaCorreoInvalido(t *testing.T) {
	rr := peticion(t, "POST", "/api/v1/alumnos", `{"matricula":"2026A0001","nombre":"Ana","correo":"no-es-correo"}`)
	if rr.Code == http.StatusCreated {
		t.Fatalf("correo invalido aceptado")
	}
}

func TestAltaRechazaNombreInvalido(t *testing.T) {
	for _, n := range []string{"", strings.Repeat("x", 200)} {
		rr := peticion(t, "POST", "/api/v1/alumnos", `{"matricula":"2026A0001","nombre":"`+n+`","correo":"ana@ung.mx"}`)
		if rr.Code == http.StatusCreated {
			t.Errorf("nombre de %d caracteres aceptado", len(n))
		}
	}
}

func TestAltaLimitaTamano(t *testing.T) {
	rr := peticion(t, "POST", "/api/v1/alumnos", `{"matricula":"2026A0001","nombre":"`+strings.Repeat("a", 50000)+`","correo":"ana@ung.mx"}`)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("cuerpo de 50 KB: esperado 413, obtenido %d", rr.Code)
	}
}

func TestConstanciaValida(t *testing.T) {
	rr := peticion(t, "GET", "/api/v1/constancias?matricula=2026A0001&formato=txt", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("constancia valida: esperado 200, obtenido %d", rr.Code)
	}
}

func TestConstanciaNoSaleDelDirectorio(t *testing.T) {
	for _, q := range []string{
		"matricula=....//main&formato=go",
		"matricula=2026A0001&formato=txt/../../main.go",
		"matricula=2026A0001&formato=pdf",
	} {
		rr := peticion(t, "GET", "/api/v1/constancias?"+q, "")
		if rr.Code == http.StatusOK {
			t.Errorf("%q devolvio 200: la ruta no esta confinada o el formato no esta en lista", q)
		}
	}
}

// La prueba crea constancias/2026A0002.txt como enlace simbolico que apunta
// fuera del directorio (a main.go). Su nombre cumple el formato de matricula:
// solo un confinamiento real (os.Root) lo detiene.
func TestConstanciaEnlaceNoEscapa(t *testing.T) {
	enlace := filepath.Join("constancias", "2026A0002.txt")
	os.Remove(enlace)
	if err := os.Symlink(filepath.Join("..", "main.go"), enlace); err != nil {
		t.Skipf("este sistema no permite crear enlaces simbolicos (%v); en Windows activa el Modo de desarrollador o usa WSL", err)
	}
	t.Cleanup(func() { os.Remove(enlace) })

	rr := peticion(t, "GET", "/api/v1/constancias?matricula=2026A0002&formato=txt", "")
	if rr.Code == http.StatusOK || strings.Contains(rr.Body.String(), "package main") {
		t.Fatalf("un enlace simbolico saco la lectura fuera de constancias/ (codigo %d)", rr.Code)
	}
}

func TestErroresNoFiltranRutas(t *testing.T) {
	rr := peticion(t, "GET", "/api/v1/constancias?matricula=2026A9999&formato=txt", "")
	cuerpo := rr.Body.String()
	if strings.Contains(cuerpo, "constancias") || strings.Contains(cuerpo, "2026A9999") {
		t.Fatalf("el mensaje de error revela rutas internas: %q", cuerpo)
	}
}
