package filtros

import (
	"fmt"
	"testing"
)

func TestComparaFiltros(t *testing.T) {
	entradas := []string{
		"expediente_100.txt",
		"../secreto.txt",
		"....//secreto.txt",
		"..././secreto.txt",
		"/etc/hosts",
		"Expediente_100.TXT",
		"expediente_100.txt\x00.pdf",
	}
	fmt.Printf("%-30s | %-22s | %-9s\n", "entrada", "bloqueo -> resultado", "permitidos")
	for _, e := range entradas {
		limpio, _ := PasaBloqueo(e)
		veredicto := "ok"
		if ContieneSalto(limpio) {
			veredicto = "SIGUE SALTANDO"
		}
		fmt.Printf("%-30q | %-22s | %-9v\n", e, veredicto, PasaPermitidos(e))
		if PasaPermitidos(e) && ContieneSalto(e) {
			t.Errorf("la lista de permitidos dejó pasar %q", e)
		}
	}
}
