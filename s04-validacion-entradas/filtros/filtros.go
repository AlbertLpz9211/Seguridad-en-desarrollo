// Package filtros compara una lista de bloqueo contra una lista de permitidos
// para nombres de archivo. Solo evalúa cadenas: no abre archivos.
package filtros

import (
	"regexp"
	"strings"
)

// PasaBloqueo es el filtro ingenuo: quita "../" y acepta todo lo demás.
func PasaBloqueo(nombre string) (string, bool) {
	limpio := strings.ReplaceAll(nombre, "../", "")
	return limpio, true
}

// ContieneSalto indica si la cadena resultante aún sube de directorio.
func ContieneSalto(s string) bool {
	return strings.Contains(s, "..") || strings.HasPrefix(s, "/")
}

var patronExpediente = regexp.MustCompile(`^[a-z0-9_-]{1,40}\.txt$`)

// PasaPermitidos es la lista de permitidos: describe lo válido y rechaza lo demás.
func PasaPermitidos(nombre string) bool {
	return patronExpediente.MatchString(nombre)
}
