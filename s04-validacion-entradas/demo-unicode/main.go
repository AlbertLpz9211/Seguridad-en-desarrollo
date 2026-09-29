// demo-unicode muestra por qué hay que canonicalizar antes de validar.
package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"golang.org/x/text/unicode/norm"
)

func main() {
	precompuesta := "jos" + string(rune(0x00E9))  // é como un solo punto de código
	descompuesta := "jose" + string(rune(0x0301)) // e + acento combinable
	cirilica := string(rune(0x0430)) + "dmin"     // "а" cirílica + "dmin"

	fmt.Printf("precompuesta %q bytes=% x\n", precompuesta, []byte(precompuesta))
	fmt.Printf("descompuesta %q bytes=% x\n", descompuesta, []byte(descompuesta))
	fmt.Println("iguales sin normalizar:", precompuesta == descompuesta)
	fmt.Println("iguales tras NFC:      ", norm.NFC.String(precompuesta) == norm.NFC.String(descompuesta))

	for _, s := range []string{"ａｄｍｉｎ", "ﬁle", "①②③"} {
		fmt.Printf("%-8q NFC=%-10q NFKC=%q\n", s, norm.NFC.String(s), norm.NFKC.String(s))
	}

	// NFKC convierte puntos y diagonales de ancho completo en "../"
	fmt.Printf("NFKC(%q) = %q\n", "．．／secreto.txt", norm.NFKC.String("．．／secreto.txt"))
	// NFKC NO convierte la "а" cirílica en "a" latina (homoglifo)
	fmt.Printf("NFKC(%q) == \"admin\": %v\n", cirilica, norm.NFKC.String(cirilica) == "admin")

	v := validator.New()
	for _, s := range []string{"jose26", "josé26", "ａｄｍｉｎ", norm.NFKC.String("ａｄｍｉｎ"), cirilica} {
		err := v.Var(s, "alphanum")
		fmt.Printf("alphanum(%q) valido=%v\n", s, err == nil)
	}
}
