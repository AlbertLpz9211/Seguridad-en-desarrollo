# Reporte BC-1 · Sesión 4 · Portal escolar

**Integrantes:** Nombre Apellido (matrícula) · Nombre Apellido (matrícula)
**Fecha:** AAAA-MM-DD
**Versión de Go:** (pega aquí la salida de `go version`)

> Instrucciones: reemplaza los textos de ejemplo (nombres, fechas y lo que está entre paréntesis en
> las secciones 1 a 6). Borra estas líneas de instrucciones antes de entregar. Las salidas de terminal van como **texto** dentro de bloques de código, no
> como capturas de pantalla.

---

## 1. Tabla de hallazgos

El portal tiene **7 fallas**. Una prueba puede revelar dos fallas, y una falla puede hacer fallar
varias pruebas: cuenta **fallas en el código**, no pruebas. Agrega las filas que necesites.

Severidad, con una frase que la justifique:
**Alta** = permite leer datos ajenos o alterar datos del negocio (por ejemplo, el promedio).
**Media** = deja pasar datos inválidos o permite tumbar el servicio.
**Baja** = revela información que ayuda a un atacante.

| # | Ubicación (archivo:línea) | Descripción de la falla | CWE | Categoría OWASP Top 10:2025 | Severidad y por qué |
| --- | --- | --- | --- | --- | --- |
| 1 | `main.go:__` | | CWE-___ | A__ | |
| 2 | | | | | |
| 3 | | | | | |
| 4 | | | | | |
| 5 | | | | | |
| 6 | | | | | |
| 7 | | | | | |

## 2. Evidencia antes de corregir

Salida completa de `go test -v ./ejercicio-bc1/` con el código **original**:

```
(pega aquí la salida)
```

## 3. Correcciones

Repite este bloque para cada falla de la tabla.

### Falla 1 · (título corto)

**Código antes:**

```go
(fragmento original)
```

**Código después:**

```go
(fragmento corregido)
```

**Por qué esta corrección cierra la falla** (no basta con «porque la prueba pasa»):

(2 a 4 líneas)

## 4. Orden del flujo de validación

Explica en qué orden procesa tu manejador `manejadorAlta` la petición (limitar → decodificar →
canonicalizar → validar → usar) y qué pasaría si se invirtiera un paso. (3 a 6 líneas)

## 5. Decisión de negocio documentada

Elige **una** y justifícala (3 a 5 líneas):

- ¿Se aceptan matrículas con la letra en minúscula (`2026a0001`)? Si sí, ¿dónde se canonicaliza?
  Si no, ¿qué mensaje ve el usuario?
- ¿Qué alfabeto aceptas en el nombre y por qué?

## 6. Evidencia después de corregir

Salida completa de `go test -v ./ejercicio-bc1/` con tu código **corregido**:

```
(pega aquí la salida)
```
