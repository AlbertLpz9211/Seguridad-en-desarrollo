// demo-openroot muestra que os.Root confina las aperturas al directorio raíz,
// incluso sin ninguna validación previa de la cadena.
//
// El programa arma su propio escenario en una carpeta temporal:
//
//	<temporal>/secreto.txt                      (fuera de la raíz)
//	<temporal>/almacen/expediente_100.txt       (dentro de la raíz)
//	<temporal>/almacen/enlace_interno  -> expediente_100.txt
//	<temporal>/almacen/enlace          -> ../secreto.txt
//
// Así funciona igual en macOS, Linux y Windows, sin depender de cómo Git
// guarde los enlaces simbólicos.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	base, err := os.MkdirTemp("", "demo-openroot-")
	if err != nil {
		fmt.Println("no se pudo crear la carpeta temporal:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(base)

	almacen := filepath.Join(base, "almacen")
	if err := os.Mkdir(almacen, 0o755); err != nil {
		fmt.Println("no se pudo crear almacen:", err)
		os.Exit(1)
	}
	escribir(filepath.Join(base, "secreto.txt"), "SECRETO_DEL_SISTEMA_NO_LEER\n")
	escribir(filepath.Join(almacen, "expediente_100.txt"), "contenido legitimo del expediente 100\n")

	hayEnlaces := true
	if err := os.Symlink("expediente_100.txt", filepath.Join(almacen, "enlace_interno")); err != nil {
		hayEnlaces = false
	}
	if err := os.Symlink(filepath.Join("..", "secreto.txt"), filepath.Join(almacen, "enlace")); err != nil {
		hayEnlaces = false
	}
	if !hayEnlaces {
		fmt.Println("AVISO: este sistema no permitió crear enlaces simbólicos.")
		fmt.Println("En Windows activa el Modo de desarrollador o usa WSL; las líneas de enlaces se omiten.")
	}

	raiz, err := os.OpenRoot(almacen)
	if err != nil {
		fmt.Println("no se pudo abrir la raiz:", err)
		os.Exit(1)
	}
	defer raiz.Close()

	nombres := []string{"expediente_100.txt", "enlace_interno", "../secreto.txt", "enlace", "/etc/hosts"}
	for _, nombre := range nombres {
		if !hayEnlaces && (nombre == "enlace" || nombre == "enlace_interno") {
			continue
		}
		f, err := raiz.Open(nombre)
		if err != nil {
			fmt.Printf("%-20q -> RECHAZADO: %v\n", nombre, err)
			continue
		}
		datos, _ := io.ReadAll(f)
		f.Close()
		fmt.Printf("%-20q -> PERMITIDO: %q\n", nombre, datos)
	}
}

func escribir(ruta, contenido string) {
	if err := os.WriteFile(ruta, []byte(contenido), 0o644); err != nil {
		fmt.Println("no se pudo escribir", ruta, ":", err)
		os.Exit(1)
	}
}
