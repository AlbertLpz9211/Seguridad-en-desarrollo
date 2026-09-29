# S4 · Validación y sanitización de entradas: listas de permitidos frente a listas de bloqueo

**Sesión 4** · lunes 28 de septiembre de 2026 · Unidad 2 · Tema 2.1
**Evidencia:** Reporte de vulnerabilidades identificadas en código con las correcciones aplicadas (actividad **BC-1**).

> **Resumen de lo que entregas** (detalles en la sección 8):
>
> | Qué | Dónde | Fecha límite |
> | --- | --- | --- |
> | Reporte BC-1 + código corregido | Tu repositorio del proyecto integrador + enlace en la plataforma del curso | **martes 29 de septiembre de 2026, 23:59** (ver nota de fecha en 8.8) |
> | Validación aplicada a tu proyecto + etiqueta `v0.1-validacion` | Tu repositorio del proyecto integrador | **lunes 5 de octubre de 2026** (inicio de la S6) |

---

## Contenido

1. [Qué vas a aprender](#1-qué-vas-a-aprender)
2. [Preparación (5 minutos)](#2-preparación-5-minutos)
3. [Qué hay en esta carpeta](#3-qué-hay-en-esta-carpeta)
4. [Actividad 1 · La lista de bloqueo pierde](#4-actividad-1--la-lista-de-bloqueo-pierde)
5. [Actividad 2 · Validación por esquema en el servidor](#5-actividad-2--validación-por-esquema-en-el-servidor)
6. [Actividad 3 · Canonicalizar antes de validar](#6-actividad-3--canonicalizar-antes-de-validar)
7. [Actividad 4 · Confinar el acceso a archivos con `os.OpenRoot`](#7-actividad-4--confinar-el-acceso-a-archivos-con-osopenroot)
8. [Actividad BC-1 · Corrige el portal escolar (se califica)](#8-actividad-bc-1--corrige-el-portal-escolar-se-califica)
9. [Para tu proyecto integrador (entrega en la S6)](#9-para-tu-proyecto-integrador-entrega-en-la-s6)
10. [Problemas frecuentes](#10-problemas-frecuentes)
11. [Glosario y referencias](#11-glosario-y-referencias)

**Regla de oro del laboratorio:** todo se ejecuta en tu computadora con `go test` y `go run`. No
vas a levantar servidores ni a atacar nada. Probar estas técnicas contra un sistema ajeno sin
autorización escrita es delito (Código Penal Federal, arts. 211 bis 1 a 211 bis 7).

---

## 1. Qué vas a aprender

Al terminar esta sesión podrás:

1. **Explicar**, con un ejemplo, por qué una lista de bloqueo falla y una lista de permitidos no, y
   cuál es el costo de la lista de permitidos (rechazar entradas legítimas).
2. **Aplicar** validación por esquema en un manejador HTTP de Go con un DTO, `validator/v10`,
   `DisallowUnknownFields` y `http.MaxBytesReader`.
3. **Implementar** el orden correcto: **limitar → decodificar → canonicalizar → validar → usar
   confinado**, y justificarlo por escrito.
4. **Analizar** un programa vulnerable: localizar cada falla y clasificarla con su CWE y su
   categoría del OWASP Top 10:2025.
5. **Corregir** ese programa y **argumentar** por qué cada corrección cierra la falla.

**Términos que vas a encontrar** (están también en el [glosario](#11-glosario-y-referencias)):
**DTO** = struct que contiene solo los campos que el cliente puede enviar ·
**CWE** = catálogo internacional de tipos de debilidad de software, cada una con un número (por ejemplo, CWE-22) ·
**OWASP Top 10:2025** = las 10 categorías de riesgo más comunes en aplicaciones web, con códigos A01 a A10 ·
**NFC / NFKC** = formas estándar de normalizar texto Unicode (sección 6).

**Idea central:** un filtro que enumera lo malo apuesta contra la creatividad del atacante y
pierde. Un filtro que **describe lo válido** (tipo, longitud, formato, conjunto cerrado) y rechaza
todo lo demás no necesita adivinar nada. Y como el cliente está en manos del atacante, esa
descripción vive **en el servidor**.

---

## 2. Preparación (5 minutos)

Haz estos pasos **en orden**. Cada uno dice qué debes ver; no avances si ves otra cosa (revisa la
[sección 10](#10-problemas-frecuentes)).

**Paso 2.1 · Actualiza el repositorio.** Si todavía no lo tienes, sigue primero el
[README principal](../README.md).

```
cd ~/Seguridad-en-desarrollo
git pull
```

Debes ver `Already up to date.` (en español: `Ya está actualizado.`) o una lista de archivos actualizados.

**Paso 2.2 · Entra a la carpeta de la sesión.** A partir de aquí, **todos** los comandos se
ejecutan dentro de esta carpeta.

```
cd s04-validacion-entradas
```

**Paso 2.3 · Comprueba la versión de Go.**

```
go version
```

Debes ver `go1.26` o mayor, por ejemplo `go version go1.26.8 windows/amd64`.

**Paso 2.4 · Prueba que todo compila.**

```
go test ./filtros/ ./seguro/
```

Debes ver exactamente dos líneas que empiezan con `ok` (el número de segundos puede variar):

```
ok  	lab-s04/filtros	0.438s
ok  	lab-s04/seguro	0.363s
```

Si ves esas dos líneas, **ya estás listo**. No necesitas internet: las dependencias vienen en la
carpeta `vendor/`.

---

## 3. Qué hay en esta carpeta

| Carpeta o archivo | Qué es | Se usa en |
| --- | --- | --- |
| `filtros/` | Dos filtros de nombres de archivo: uno por lista de bloqueo y otro por lista de permitidos, más una prueba que los compara | Actividad 1 |
| `seguro/` | Un servidor HTTP escrito correctamente, con sus pruebas | Actividades 2 y 4 |
| `demo-unicode/` | Programa que muestra trucos de Unicode | Actividad 3 |
| `demo-openroot/` | Programa que muestra cómo `os.OpenRoot` confina el acceso a archivos | Actividad 4 |
| `almacen/` | Archivo de datos que usa `seguro/` | Actividad 4 |
| `ejercicio-bc1/` | **El portal escolar vulnerable que vas a corregir** y sus 10 pruebas | Actividad BC-1 |
| `plantilla-reporte-bc1.md` | Plantilla del reporte que entregas | Actividad BC-1 |
| `go.mod`, `go.sum`, `vendor/` | Configuración del módulo y dependencias. **No los modifiques.** | — |

---

## 4. Actividad 1 · La lista de bloqueo pierde

**Tiempo:** 20 minutos.

Abre `filtros/filtros.go`. Tiene dos funciones:

- `PasaBloqueo`: la **lista de bloqueo**. Quita la secuencia `../` del nombre y deja pasar el
  resultado.
- `PasaPermitidos`: la **lista de permitidos**. Acepta el nombre solo si coincide completo con el
  patrón `^[a-z0-9_-]{1,40}\.txt$` (letras minúsculas, números, `_` o `-`, de 1 a 40 caracteres,
  terminado en `.txt`).

**Paso 4.1 · Predice antes de ejecutar.** Copia esta tabla en tu cuaderno. Para cada entrada,
escribe **Sí** o **No** en las dos columnas **antes** de correr nada.

| Entrada | ¿Después de `PasaBloqueo` todavía sube de directorio? | ¿`PasaPermitidos` la acepta? |
| --- | --- | --- |
| `expediente_100.txt` | | |
| `../secreto.txt` | | |
| `....//secreto.txt` | | |
| `..././secreto.txt` | | |
| `/etc/hosts` | | |
| `Expediente_100.TXT` | | |
| `expediente_100.txt\x00.pdf` (`\x00` es el byte nulo) | | |

**Paso 4.2 · Ejecuta la comparación.**

```
go test -v ./filtros/
```

Debes ver:

```
=== RUN   TestComparaFiltros
entrada                        | bloqueo -> resultado   | permitidos
"expediente_100.txt"           | ok                     | true
"../secreto.txt"               | ok                     | false
"....//secreto.txt"            | SIGUE SALTANDO         | false
"..././secreto.txt"            | SIGUE SALTANDO         | false
"/etc/hosts"                   | SIGUE SALTANDO         | false
"Expediente_100.TXT"           | ok                     | false
"expediente_100.txt\x00.pdf"   | ok                     | false
--- PASS: TestComparaFiltros (0.00s)
PASS
ok  	lab-s04/filtros	0.276s
```

**Cómo leerla:**

- `SIGUE SALTANDO` significa que, después de pasar por la lista de bloqueo, el nombre **todavía**
  contiene `..` o empieza con `/`: el filtro no protegió.
- `....//secreto.txt`: `ReplaceAll` quita el `../` de en medio y lo que queda **se vuelve a juntar**
  como `../secreto.txt`. El filtro fabricó el ataque. Por eso **se rechaza, no se «limpia»**.
- `/etc/hosts`: la lista de bloqueo no contemplaba rutas absolutas.
- La columna `permitidos` rechazó 6 de 7 **sin mencionar** `..`, `/` ni `\x00` en ningún lado:
  simplemente no caben en el patrón.
- `Expediente_100.TXT` también se rechazó, aunque quizá era legítimo. **Ese es el costo de la lista
  de permitidos**: rechaza lo que no se previó. Aceptar mayúsculas es una decisión del negocio que se
  toma a propósito y se documenta.
- Sobre `\x00`: en C, el byte nulo corta la cadena. En Go no pasa eso (Go rechaza la ruta con
  `invalid argument`), pero sí puede pasar si la cadena llega a código escrito en C.

**Paso 4.3 · Responde en tu cuaderno** (1 línea cada una):

1. ¿Cuántas de tus predicciones acertaste?
2. Si mañana el sistema debe aceptar también archivos `.pdf`, ¿qué cambia en cada filtro?

---

## 5. Actividad 2 · Validación por esquema en el servidor

**Tiempo:** 20 minutos.

Abre `seguro/main.go` y localiza la función `manejadorRegistro`. Procesa el alta de un usuario en
este orden (cada número está marcado en el código con un comentario):

| # | Paso | Qué hace | Si falla responde |
| --- | --- | --- | --- |
| 1 | Limitar | `http.MaxBytesReader` corta el cuerpo en 10 KiB **antes** de leerlo | `413` |
| 2 | Decodificar estrictamente | El struct `PeticionRegistro` (el **DTO**) es el esquema; `DisallowUnknownFields` rechaza campos que no estén en él; después se exige que no venga nada más | `400` |
| 3 | Canonicalizar | Normaliza el texto con NFKC y pasa el correo a minúsculas | — |
| 4 | Validar la sintaxis | Las etiquetas `validate:"..."` del DTO: `required`, `alphanum`, `min`, `max`, `email`, `oneof` | `422` |
| 5 | Validar el negocio | Rechaza nombres reservados (`admin`, `root`) | `422` |

**DTO** (*Data Transfer Object*): un struct que contiene **solo** los campos que el cliente puede
enviar. No es el mismo struct que se guarda en la base de datos.

**Paso 5.1 · Ejecuta las pruebas del registro.**

```
go test -v -run TestRegistro ./seguro/
```

Debes ver:

```
=== RUN   TestRegistro
valido                               -> 201 {"datos":{"nombre_usuario":"ana26","correo":"ana@ung.mx","edad":20,"carrera":"software"},"mensaje":"usuario registrado"}
carrera fuera de lista               -> 422 {"error":"datos fuera del esquema"}
campo extra (asignacion masiva)      -> 400 {"error":"json invalido o con campos no permitidos"}
edad como texto                      -> 400 {"error":"json invalido o con campos no permitidos"}
menor de edad                        -> 422 {"error":"datos fuera del esquema"}
nombre reservado en ancho completo   -> 422 {"error":"nombre de usuario reservado"}
dos objetos pegados                  -> 400 {"error":"se esperaba un solo objeto json"}
cuerpo de 20 KiB                     -> 413 {"error":"cuerpo demasiado grande"}
--- PASS: TestRegistro (0.00s)
PASS
ok  	lab-s04/seguro	0.363s
```

Cada línea es una petición distinta. Para ver exactamente qué se envió en cada caso, abre
`seguro/main_test.go` y busca el nombre del caso (por ejemplo, `"campo extra (asignacion masiva)"`).

**Paso 5.2 · Responde en tu cuaderno:**

1. En el caso `valido` se envió el correo `Ana@UNG.mx`. ¿Por qué la respuesta dice `ana@ung.mx`?
   ¿Qué paso de la tabla lo hizo?
2. ¿Por qué unos casos responden `400` y otros `422`? (Pista: `400` = la petición no tiene la forma
   del esquema; `422` = tiene la forma, pero el contenido no cumple las reglas.)
3. El caso `campo extra` envía `"rol":"docente"`. Si se quitara la línea
   `dec.DisallowUnknownFields()`, ¿qué código respondería? ¿Por qué eso sería peligroso si el
   struct tuviera un campo `Rol`?

---

## 6. Actividad 3 · Canonicalizar antes de validar

**Tiempo:** 10 minutos.

**Canonicalizar** es llevar un dato a su forma única antes de revisarlo. Dos cadenas que **se ven
iguales** pueden ser **bytes distintos**. Si validas una forma y el sistema usa otra, tu validación
revisó algo que no es lo que se usa.

**Paso 6.1 · Ejecuta la demostración.**

```
go run ./demo-unicode
```

Debes ver:

```
precompuesta "josé" bytes=6a 6f 73 c3 a9
descompuesta "josé" bytes=6a 6f 73 65 cc 81
iguales sin normalizar: false
iguales tras NFC:       true
"ａｄｍｉｎ"  NFC="ａｄｍｉｎ"    NFKC="admin"
"ﬁle"    NFC="ﬁle"      NFKC="file"
"①②③"    NFC="①②③"      NFKC="123"
NFKC("．．／secreto.txt") = "../secreto.txt"
NFKC("аdmin") == "admin": false
alphanum("jose26") valido=true
alphanum("josé26") valido=false
alphanum("ａｄｍｉｎ") valido=false
alphanum("admin") valido=true
alphanum("аdmin") valido=false
```

> Si tu terminal muestra cuadros o signos `?` en lugar de algunos caracteres, es la fuente de la
> terminal, no un error del programa. En Windows usa **Windows Terminal**, que los muestra bien.

**Cómo leerla:**

| Línea | Qué demuestra |
| --- | --- |
| `precompuesta` / `descompuesta` | La `é` puede ser **un** carácter (`c3 a9`) o **dos**: `e` + acento suelto (`65 cc 81`). Se ven iguales, pero `==` dice que son distintas. |
| `iguales tras NFC: true` | **NFC** las une en una sola forma. |
| `ａｄｍｉｎ` → `admin` | **NFKC** además convierte letras de «ancho completo», ligaduras (`ﬁ`) y números en círculo a su forma básica. |
| `．．／secreto.txt` → `../secreto.txt` | Un filtro que busca `../` **antes** de NFKC no ve nada; si algo normaliza **después**, aparece el salto de directorio. |
| `аdmin` == `admin`: `false` | Esa `а` es **cirílica**. NFKC **no** la convierte en `a` latina. Lo que la detiene es la lista de permitidos: `alphanum` solo acepta letras ASCII. |
| `alphanum("josé26")`: `false` | `alphanum` rechaza acentos. Si tu sistema necesita nombres con acento, usa `alphanumunicode` junto con NFC, y documenta la decisión. |

**Regla:** **limitar → decodificar → canonicalizar → validar → usar confinado.** Nunca al revés.

---

## 7. Actividad 4 · Confinar el acceso a archivos con `os.OpenRoot`

**Tiempo:** 15 minutos.

La lista de permitidos revisa la **cadena**. `os.OpenRoot` (disponible desde Go 1.24) protege el
**sistema de archivos**: abre una carpeta como «raíz» y no deja que ninguna apertura salga de ella,
ni con `..`, ni con rutas absolutas, ni con enlaces simbólicos que apunten afuera. Son dos capas de
defensa distintas.

**Paso 7.1 · Ejecuta la demostración.** El programa crea, en una carpeta temporal, un almacén con
un archivo legítimo, un secreto **fuera** del almacén y dos enlaces simbólicos: uno que se queda
dentro (`enlace_interno`) y otro que apunta al secreto (`enlace`). Después intenta abrirlos **sin
ninguna validación de la cadena**.

```
go run ./demo-openroot
```

Debes ver (macOS y Linux):

```
"expediente_100.txt" -> PERMITIDO: "contenido legitimo del expediente 100\n"
"enlace_interno"     -> PERMITIDO: "contenido legitimo del expediente 100\n"
"../secreto.txt"     -> RECHAZADO: openat ../secreto.txt: path escapes from parent
"enlace"             -> RECHAZADO: openat enlace: path escapes from parent
"/etc/hosts"         -> RECHAZADO: openat /etc/hosts: path escapes from parent
```

> **En Windows** el texto después de `RECHAZADO:` es distinto; lo que importa es que las tres
> últimas digan **RECHAZADO**. Si no activaste el Modo de desarrollador (README principal, 1.4),
> verás un **AVISO** y no aparecerán las líneas `enlace_interno` ni `enlace`. Es normal.

**Cómo leerla:**

- `enlace` tiene un nombre inocente, que pasaría cualquier filtro de texto. Aun así se rechaza
  porque `os.OpenRoot` **sigue el enlace** y descubre que su destino está fuera de la raíz.
- `enlace_interno` también es un enlace, pero apunta dentro: se permite.
- `os.OpenRoot` **no** revisa qué archivo lees **dentro** de la raíz. Si alguien deja un archivo
  sensible dentro del almacén, se puede leer. Por eso se usan **las dos capas**: lista de permitidos
  **y** confinamiento.

**Paso 7.2 · Mira las dos capas juntas en el servidor.** En `seguro/main.go`, la función
`nuevoManejadorExpedientes` primero valida el nombre con un patrón (capa 1) y luego abre con
`os.Root` (capa 2).

```
go test -v -run TestExpedientes ./seguro/
```

Debes ver:

```
=== RUN   TestExpedientes
"expediente_100.txt"   -> 200 contenido legitimo del expediente 100
"expediente_999.txt"   -> 404 recurso no encontrado
"../secreto.txt"       -> 400 nombre de archivo no permitido
"Expediente_100.TXT"   -> 400 nombre de archivo no permitido
--- PASS: TestExpedientes (0.00s)
PASS
ok  	lab-s04/seguro	0.346s
```

`../secreto.txt` responde `400` porque **la capa 1 lo detuvo** y nunca llegó a `os.Root`. La
demostración del paso 7.1 es la que prueba la capa 2 por separado.

---

## 8. Actividad BC-1 · Corrige el portal escolar (se califica)

**Tiempo en clase:** 28 minutos. Lo que no termines, lo terminas en casa.
**Modalidad:** en parejas (si el grupo es impar, un equipo de tres).
**Entrega:** **una por pareja**. El mismo reporte y el mismo código van en el repositorio del proyecto integrador de **cada** integrante, y **cada** integrante sube su propio enlace en la plataforma. El reporte lleva los nombres de todos. Si trabajaste solo, lleva solo tu nombre.
**Fecha límite:** martes 29 de septiembre de 2026, 23:59 (ver nota de fecha en 8.8).

### 8.1 Qué es el portal escolar

`ejercicio-bc1/main.go` es un servicio con dos funciones:

- `POST /api/v1/alumnos`: da de alta a un estudiante con `matricula`, `nombre` y `correo`.
- `GET /api/v1/constancias?matricula=...&formato=...`: devuelve la constancia de estudios guardada
  en la carpeta `constancias/`.

Tiene **fallas de validación deliberadas**. Tu trabajo es encontrarlas, clasificarlas y corregirlas.

Datos de negocio que necesitas:

- **Formato de matrícula:** 4 dígitos, 1 letra mayúscula, 4 dígitos. Ejemplo: `2026A0001`.
- **Formato de constancia permitido:** solo `txt`.
- **El promedio** lo calcula la escuela; **nunca** lo envía el estudiante.
- **Nombres válidos:** de 2 a 80 caracteres; pueden tener espacios y acentos (`Ana Pérez` es válido).
- **Matrícula con dígitos de ancho completo (`２０２６A0001`): se rechaza, no se normaliza.** No
  apliques NFKC ni NFC a la matrícula (NFKC la convertiría en `2026A0001` y la aceptarías). Si
  decides aceptar la letra en minúscula, la única transformación permitida es `strings.ToUpper`.
  NFC o NFKC solo se aplican al **nombre**.

### 8.2 Roles en la pareja

| Rol | Qué hace |
| --- | --- |
| **Piloto** | Escribe el código. |
| **Navegante** | Lee la prueba que falla, explica qué exige y anota la falla en la tabla del reporte (ubicación, CWE, OWASP). |

**En clase, intercambien los roles cada 12 minutos. En casa, intercámbienlos cada vez que una prueba pase a `--- PASS`.** En un equipo de tres, la tercera persona es **auditora**:
revisa que cada CWE y categoría OWASP anotada sea correcta.

### 8.3 Ejecuta las pruebas

```
go test ./ejercicio-bc1/
```

Al empezar debes ver **8 pruebas que fallan**: son las líneas que empiezan con `--- FAIL`. (La terminal no las pinta de rojo; búscalas por ese texto.)

```
--- FAIL: TestAltaRechazaPromedioDelCliente (0.00s)
    main_test.go:32: el cliente no debe poder fijar su promedio (CWE-915)
--- FAIL: TestAltaRechazaMatriculaFueraDeFormato (0.00s)
    main_test.go:40: matricula "2026A0001 OR 1" no cumple el formato AAAALNNNN y fue aceptada
    main_test.go:40: matricula "abc" no cumple el formato AAAALNNNN y fue aceptada
    main_test.go:40: matricula "2026A00011111" no cumple el formato AAAALNNNN y fue aceptada
    main_test.go:40: matricula "２０２６A0001" no cumple el formato AAAALNNNN y fue aceptada
--- FAIL: TestAltaRechazaCorreoInvalido (0.00s)
    main_test.go:48: correo invalido aceptado
--- FAIL: TestAltaRechazaNombreInvalido (0.00s)
    main_test.go:56: nombre de 0 caracteres aceptado
    main_test.go:56: nombre de 200 caracteres aceptado
--- FAIL: TestAltaLimitaTamano (0.00s)
    main_test.go:64: cuerpo de 50 KB: esperado 413, obtenido 201
--- FAIL: TestConstanciaNoSaleDelDirectorio (0.00s)
    main_test.go:83: "matricula=....//main&formato=go" devolvio 200: la ruta no esta confinada o el formato no esta en lista
    main_test.go:83: "matricula=2026A0001&formato=txt/../../main.go" devolvio 200: la ruta no esta confinada o el formato no esta en lista
--- FAIL: TestConstanciaEnlaceNoEscapa (0.00s)
    main_test.go:101: un enlace simbolico saco la lectura fuera de constancias/ (codigo 200)
--- FAIL: TestErroresNoFiltranRutas (0.00s)
    main_test.go:109: el mensaje de error revela rutas internas: "no se pudo leer constancias/2026A9999.txt: open constancias/2026A9999.txt: no such file or directory\n"
FAIL
FAIL	lab-s04/ejercicio-bc1	0.328s
FAIL
```

> En Windows, la última línea de `TestErroresNoFiltranRutas` dice algo como
> `constancias\2026A9999.txt: The system cannot find the file specified.` Es equivalente.

Las otras 2 pruebas (`TestAltaValida` y `TestConstanciaValida`) **ya pasan** y deben **seguir
pasando** al terminar. Para ver las 10 con su estado usa `go test -v ./ejercicio-bc1/`.

> **En Windows sin Modo de desarrollador**, `TestConstanciaEnlaceNoEscapa` aparece como
> `--- SKIP` (omitida) en lugar de `--- FAIL`, porque no se pudo crear el enlace simbólico que
> necesita. Verás 7 líneas `--- FAIL` en lugar de 8. Eso **no** te exime de esa corrección: el docente la
> revisa a mano en tu código.

### 8.4 Qué exige cada prueba

Las pruebas describen el comportamiento **seguro**. No te dicen cómo programarlo.

| Prueba | Qué exige |
| --- | --- |
| `TestAltaValida` | Un alta correcta (con `"Ana Pérez"`) se acepta con `201`. **Ya pasa; no la rompas.** |
| `TestAltaRechazaPromedioDelCliente` | Si la petición trae `promedio`, **no** se acepta. |
| `TestAltaRechazaMatriculaFueraDeFormato` | Cualquier matrícula que no sea exactamente `AAAALNNNN` se rechaza, incluidos los dígitos de ancho completo (`２０２６`). |
| `TestAltaRechazaCorreoInvalido` | Un correo sin formato de correo se rechaza. |
| `TestAltaRechazaNombreInvalido` | Un nombre vacío o de más de 80 caracteres se rechaza. |
| `TestAltaLimitaTamano` | Un cuerpo de 50 KB se rechaza con `413`. |
| `TestConstanciaValida` | `matricula=2026A0001&formato=txt` devuelve la constancia con `200`. **Ya pasa; no la rompas.** |
| `TestConstanciaNoSaleDelDirectorio` | Ninguna combinación de `matricula` y `formato` puede leer archivos fuera de `constancias/`, y `formato=pdf` no se acepta. |
| `TestConstanciaEnlaceNoEscapa` | La prueba crea `constancias/2026A0002.txt` como **enlace simbólico** hacia `../main.go`. Su nombre tiene formato válido, así que el patrón no lo detiene: **no** debe poder leerse. |
| `TestErroresNoFiltranRutas` | Cuando una constancia no existe, la respuesta **no** debe mostrar rutas internas ni la matrícula consultada. |

### 8.5 Reglas

Consecuencias en la rúbrica (8.9): romper la **regla 1** deja en 0 el criterio «Calidad de la
corrección»; romper la **regla 2** lo baja a Insuficiente; romper la **regla 3** lo deja como máximo en
Aceptable; romper las reglas **4, 5 o 6** puede impedir que tu código compile o se revise.


1. **No modifiques `ejercicio-bc1/main_test.go`.** Solo modificas `ejercicio-bc1/main.go`.
2. **Corrige la causa, no el síntoma.** Prohibido hacer pasar una prueba con casos especiales, por
   ejemplo `if matricula == "abc" { ... }`.
3. **No uses listas de bloqueo.** Ni agregar más caracteres a `matriculaSospechosa`, ni
   `ReplaceAll`. Describe lo válido.
4. **No agregues dependencias nuevas.** Todo lo que necesitas ya está en `vendor/`:
   `github.com/go-playground/validator/v10`, `golang.org/x/text/unicode/norm` y la biblioteca
   estándar (`regexp`, `os`, `net/http`, `errors`, `io`, `log/slog`).
5. Ejecuta todo con `go test`. **No hace falta levantar el servidor.**
6. **Todo tu código va en `ejercicio-bc1/main.go`.** No crees archivos nuevos: solo se entrega ese archivo.

### 8.6 Cómo saber que terminaste

```
go test -v ./ejercicio-bc1/
```

Terminaste cuando las **10** pruebas dicen `--- PASS` (o 9 `PASS` y 1 `SKIP` en Windows sin Modo de
desarrollador) y la última línea es `ok  	lab-s04/ejercicio-bc1`.

**Que las pruebas pasen es necesario, pero no suficiente.** El docente revisa además:

- que el cuerpo se envuelva con `http.MaxBytesReader` **antes** de `Decode` (revisar `r.ContentLength` no cuenta, porque un cliente puede no enviar esa cabecera);
- que las correcciones sean por **lista de permitidos**;
- que la lectura de constancias use `os.OpenRoot`.

### 8.7 Pistas (úsalas solo si llevas más de 5 minutos atorado)

<details>
<summary>Pista 1 · Cómo leer una prueba que falla</summary>

Lee el nombre de la prueba y su mensaje como una historia: «¿qué debería pasar?» y «¿qué pasó?».
Luego abre `main_test.go`, busca esa función y mira qué petición envía.
</details>

<details>
<summary>Pista 2 · Matrícula</summary>

No busques caracteres malos. Escribe **la forma de una matrícula válida** como expresión regular
anclada con `^` y `$`. Recuerda que en Go `[0-9]` solo acepta dígitos ASCII.
</details>

<details>
<summary>Pista 3 · Promedio</summary>

¿Quién decide el promedio: quien se inscribe o el sistema? ¿Qué struct lees de la petición? ¿Qué
pasaría si fueran dos structs distintos? Revisa cómo lo hace `seguro/main.go`.
</details>

<details>
<summary>Pista 4 · Constancias</summary>

Hay **dos** parámetros que terminan formando una ruta. ¿Los validaste los dos? Y para el enlace
simbólico: ¿qué función abre archivos mirando el disco y no solo la cadena? (Actividad 4.)
</details>

<details>
<summary>Pista 5 · Mensajes de error</summary>

¿Quién necesita saber la ruta del archivo: el usuario o tú, en el registro (*log*)? Puedes usar
`slog.Warn(...)` para el detalle y responder al cliente un mensaje genérico.
</details>

### 8.8 Qué entregas y dónde

Haz esto **cuando las pruebas pasen** (8.6), no antes.

**En tu repositorio del proyecto integrador** van estos dos archivos:

| Archivo | Contenido |
| --- | --- |
| `docs/auditorias/reporte-bc1-s04.md` | Tu reporte, hecho a partir de `plantilla-reporte-bc1.md` (en esta carpeta). |
| `docs/auditorias/s04-portal-corregido.go` | Una copia de tu `ejercicio-bc1/main.go` corregido. |

**Paso 1 · Crea la carpeta** en tu proyecto (cambia `ruta-de-tu-proyecto` por la carpeta real de tu
proyecto integrador):

```
mkdir -p ~/ruta-de-tu-proyecto/docs/auditorias
```

En Windows PowerShell: `mkdir ~\ruta-de-tu-proyecto\docs\auditorias` (si ya existe, muestra un
error que puedes ignorar).

**Paso 2 · Copia la plantilla y tu código**, desde la carpeta `s04-validacion-entradas`:

```
cp plantilla-reporte-bc1.md ~/ruta-de-tu-proyecto/docs/auditorias/reporte-bc1-s04.md
cp ejercicio-bc1/main.go    ~/ruta-de-tu-proyecto/docs/auditorias/s04-portal-corregido.go
```

(En PowerShell `cp` también funciona.)

**Paso 3 · Evita que la copia rompa tu proyecto.** Abre
`docs/auditorias/s04-portal-corregido.go` y escribe como **primera línea** exactamente:

```go
//go:build ignore
```

seguida de una **línea en blanco**. Así Go no intenta compilar esa copia junto con tu proyecto.

**Paso 4 · Llena el reporte** en `docs/auditorias/reporte-bc1-s04.md`, siguiendo sus instrucciones.

**Paso 5 · Guarda y sube:**

```
cd ~/ruta-de-tu-proyecto
git add docs/auditorias
git commit -m "docs: reporte BC-1 de la sesion 4"
git push
```

Debes ver al final una línea parecida a `main -> main`. Abre tu repositorio en GitHub y comprueba
que aparecen los dos archivos en `docs/auditorias/`.

**Paso 6 · Entrega en la plataforma del curso.** En la tarea «BC-1 · S4» pega el **enlace al
commit**: en GitHub abre la pestaña *Commits*, entra al commit «docs: reporte BC-1 de la sesion 4» y
copia la dirección de la barra del navegador. Si tu repositorio es **privado**, antes agrega al
docente como colaborador (Settings → Collaborators → usuario `AlbertLpz9211`). **Si el docente no
puede abrir el enlace, la entrega cuenta como no entregada.**

**Fecha límite:** martes 29 de septiembre de 2026, 23:59. **La fecha oficial es la que el docente
publique en la plataforma al terminar la clase.** Si se aplaza, será al jueves 1 de octubre de 2026,
23:59.

### 8.9 Rúbrica (100 puntos; calificación = puntos ÷ 10; mínima aprobatoria 7.0)

| Criterio | Puntos | Sobresaliente | Aceptable | Insuficiente |
| --- | --- | --- | --- | --- |
| Identificación y clasificación | 30 | Al menos 6 de las 7 fallas con CWE y categoría OWASP 2025 correctas y severidad justificada (27–30) | 4 o 5 fallas, o clasificaciones con errores menores (21–26) | Menos de 4 fallas, o usa la lista OWASP 2021 (0–20) |
| Evidencia antes y después | 20 | Salida completa de `go test -v` antes y después, como texto (18–20) | Salidas incompletas (14–17) | Sin salidas o no reproducibles (0–13) |
| Calidad de la corrección | 35 | 10/10 pruebas (o 9 `PASS` + 1 `SKIP` de `TestConstanciaEnlaceNoEscapa` en Windows sin Modo de desarrollador; el docente la ejecuta en macOS/Linux) sin trucos; DTO, lista de permitidos, `os.OpenRoot`, errores genéricos, `MaxBytesReader` antes de `Decode` (31–35) | 8 o 9 `PASS`, o alguna corrección con lista de bloqueo (25–30) | Menos de 8 `PASS`, no compila o modifica `main_test.go` (0–24) |
| Justificación y decisión de negocio | 15 | Explica por qué cada cambio cierra la falla, justifica el orden del flujo y documenta una decisión de negocio (14–15) | Explicaciones superficiales (11–13) | Sin justificación (0–10) |

**Penalizaciones:** modificar `main_test.go` → el criterio «Calidad de la corrección» vale 0.
Probar las técnicas contra cualquier sistema que no sea el tuyo → el producto se anula y se reporta
a coordinación.

---

## 9. Para tu proyecto integrador (entrega en la S6)

Esto **no** forma parte de la BC-1; se califica con el proyecto (40% de la materia).
**Fecha:** lunes 5 de octubre de 2026, al inicio de la S6.

En tu proyecto, **al menos un endpoint** que reciba JSON debe tener:

1. un DTO con etiquetas de `validator/v10`;
2. `dec.DisallowUnknownFields()`;
3. `http.MaxBytesReader` **antes** de leer el cuerpo;
4. respuestas `400`/`413`/`422` con mensajes genéricos.

Cuando funcione, revisa con `git status` que solo aparezcan **tus** archivos (nada de ejecutables
ni carpetas del editor). Después crea y sube la etiqueta:

```
git add .
git commit -m "feat: validacion de entradas por esquema y limite de cuerpo"
git tag v0.1-validacion
git push origin main --tags
```

Si tu rama principal no se llama `main`, usa su nombre en el último comando. Comprueba en GitHub →
*Tags* que aparece `v0.1-validacion`.

Usa `seguro/main.go` de esta carpeta como ejemplo.

---

## 10. Problemas frecuentes

| Qué ves | Causa | Solución |
| --- | --- | --- |
| `go: cannot find main module` o `go.mod file not found` | No estás dentro de `s04-validacion-entradas` | `cd ~/Seguridad-en-desarrollo/s04-validacion-entradas` |
| `inconsistent vendoring` | Se modificó `go.mod` o `vendor/` | `git checkout -- go.mod go.sum vendor` |
| `no required module provides package ...` | Importaste un paquete que no está en `vendor/` | Usa solo los paquetes de la regla 4 (8.5) |
| `TestAltaValida` empezó a fallar | Tu validación del nombre es demasiado estricta (por ejemplo, pusiste `alphanum`) | `Ana Pérez` tiene espacio y acento, y es válido. Usa `min=2,max=80` |
| `panic: ... Undefined validation function 'matricula'` | Usaste la etiqueta `matricula` sin registrarla | Regístrala con `validar.RegisterValidation("matricula", ...)` antes de usarla |
| `TestConstanciaEnlaceNoEscapa` sigue en rojo aunque validaste el formato | El enlace tiene un nombre válido; el patrón no basta | Abre la constancia con `os.OpenRoot` (actividad 4) |
| `--- SKIP: TestConstanciaEnlaceNoEscapa` | Windows sin Modo de desarrollador | Actívalo (README principal, 1.4) o sigue: el docente lo revisa a mano |
| `git pull` falla por cambios locales | Modificaste `ejercicio-bc1/main.go` | `git stash`, `git pull`, `git stash pop` (README principal, sección 3) |
| Caracteres raros en la actividad 3 | Fuente de la terminal | Usa Windows Terminal o la terminal de VS Code |

---

## 11. Glosario y referencias

| Término | Significado |
| --- | --- |
| Lista de bloqueo (*blocklist*) | Filtro que enumera lo prohibido y deja pasar el resto |
| Lista de permitidos (*allowlist*) | Filtro que describe lo válido y rechaza el resto |
| Canonicalizar | Llevar un dato a su forma única antes de validarlo |
| DTO | Struct que contiene solo los campos que el cliente puede enviar |
| Asignación masiva (*mass assignment*, CWE-915) | El cliente fija campos que no debería controlar, enviándolos de más |
| Salto de directorio (*path traversal*, CWE-22) | Salir de la carpeta permitida con `..`, rutas absolutas o enlaces |
| Enlace simbólico (*symlink*) | Archivo que apunta a otro archivo o carpeta |

| CWE | Catálogo internacional de tipos de debilidad de software (Common Weakness Enumeration) |
| OWASP Top 10:2025 | Las 10 categorías de riesgo más comunes en aplicaciones web (A01 a A10) |

**Clasificación oficial que vas a usar en el reporte.** Clasifica cada falla **por su efecto**: si
la entrada termina formando una **ruta de archivo**, es CWE-22 / A01 aunque también sea un problema
de validación. Para CWE-770 escribe `A10` y justifícalo en una frase.

| CWE | Nombre | Categoría OWASP Top 10:2025 |
| --- | --- | --- |
| CWE-20 | Validación de entrada inadecuada | A05 Injection |
| CWE-22 | Salto de directorio | A01 Broken Access Control |
| CWE-915 | Asignación masiva | A08 Software or Data Integrity Failures |
| CWE-770 | Asignación de recursos sin límite | Relacionada con A10 Mishandling of Exceptional Conditions (y API4:2023) |
| CWE-209 | Mensaje de error con información sensible | A10 Mishandling of Exceptional Conditions |

Referencias:

- OWASP Input Validation Cheat Sheet — <https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html>
- OWASP Mass Assignment Cheat Sheet — <https://cheatsheetseries.owasp.org/cheatsheets/Mass_Assignment_Cheat_Sheet.html>
- OWASP Top 10:2025 — <https://owasp.org/Top10/2025/>
- Go, `os.Root` — <https://pkg.go.dev/os#Root> · «Traversal-resistant file APIs» — <https://go.dev/blog/osroot>
- Go, `http.MaxBytesReader` — <https://pkg.go.dev/net/http#MaxBytesReader>
- `validator/v10` — <https://pkg.go.dev/github.com/go-playground/validator/v10>
