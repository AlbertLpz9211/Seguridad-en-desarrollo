# Seguridad en Desarrollo de Software — Laboratorios

Universidad Nueva Galicia · Ingeniería en Desarrollo de Software Full Stack
Clave DSING007 · Cuarto cuatrimestre · Septiembre–diciembre 2026 · Grupo IDSE4A
Docente: Mtro. Alberto López García

Este repositorio contiene el código de los laboratorios de la materia. Cada sesión con laboratorio
tiene su propia carpeta y su propio manual (`README.md`) con los pasos exactos.

| Sesión | Fecha | Tema | Carpeta |
| --- | --- | --- | --- |
| S4 | lunes 28 de septiembre de 2026 | 2.1 Validación y sanitización de entradas | [`s04-validacion-entradas/`](s04-validacion-entradas/README.md) |

---

## Antes de empezar: reglas de uso (léelas completas)

> **Todo el código de este repositorio se ejecuta únicamente en tu computadora.** Las pruebas usan
> `go test` y no abren conexiones a ningún otro equipo.
>
> Probar cualquier técnica de este curso contra un sistema que no es tuyo, sin autorización escrita
> de su dueño, puede constituir el delito de acceso ilícito a sistemas y equipos de informática
> (Código Penal Federal, artículos 211 bis 1 a 211 bis 7). Esto incluye los sistemas de la
> universidad y los de tus compañeros. **Sin autorización escrita, no se prueba.**

---

## 1. Qué necesitas instalar (una sola vez)

Necesitas tres cosas: **Go 1.26 o más reciente**, **Git** y un **editor de código**.

### 1.1 Comprueba si ya los tienes

Abre una terminal:

- **Windows:** menú Inicio → escribe `PowerShell` → abre «Windows PowerShell».
- **macOS:** `Cmd + Espacio` → escribe `Terminal` → Enter.
- **Linux:** `Ctrl + Alt + T`.

Escribe estos dos comandos, uno por uno, y presiona Enter después de cada uno:

```
go version
git --version
```

**Lo que debes ver** (el final de cada línea cambia según tu sistema; en Windows Git dice algo como `2.50.1.windows.1`):

```
go version go1.26.8 darwin/arm64
git version 2.50.1
```

Para Go importa que el número sea **1.26 o mayor**. Para Git sirve cualquier versión.

- Si `go version` muestra `go1.26` o mayor, **Go está listo**; salta a 1.3.
- Si muestra `go1.25` o menor, o dice `command not found` / `no se reconoce como nombre de un
  cmdlet`, instala Go (1.2).
- Si `git --version` falla, instala Git (1.2).

### 1.2 Instalar Go y Git

| Sistema | Go | Git |
| --- | --- | --- |
| **Windows 10/11** | En PowerShell: `winget install GoLang.Go` — o descarga el instalador `.msi` de <https://go.dev/dl/> y dale «Siguiente» hasta terminar. | `winget install Git.Git` — o el instalador de <https://git-scm.com/download/win> |
| **macOS** | Si tienes Homebrew: `brew install go`. Si `brew` no existe, descarga el instalador `.pkg` de <https://go.dev/dl/> y ábrelo | `xcode-select --install` (o `brew install git`) |
| **Linux (Ubuntu/Debian)** | Descarga `go1.26.x.linux-amd64.tar.gz` de <https://go.dev/dl/> a tu carpeta Descargas y ejecuta: `sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf ~/Descargas/go1.26.*.linux-amd64.tar.gz && echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile`. Luego cierra sesión y vuelve a entrar. (No uses el paquete `golang` de `apt`: suele ser viejo.) | `sudo apt install git` |

**Después de instalar, cierra la terminal y abre una nueva.** Vuelve a correr `go version` para
confirmar. Si sigue sin reconocerse, reinicia la computadora.

### 1.3 Editor

Recomendado: **Visual Studio Code** (<https://code.visualstudio.com/>) con la extensión oficial
**Go** (publicada por «Go Team at Google»). Al abrir un archivo `.go` por primera vez, VS Code te
ofrecerá instalar herramientas adicionales: acepta («Install All»).

### 1.4 Solo en Windows: enlaces simbólicos

Un laboratorio (S4) usa enlaces simbólicos. Windows solo permite crearlos si activas el **Modo de
desarrollador**: Configuración → Privacidad y seguridad → Para programadores → **Modo de
desarrollador: Activado**. Si no lo activas, todo funciona igual, salvo una prueba que se marcará
como `SKIP` (omitida) y dos líneas de una demostración que no aparecerán. El manual de la sesión
te dice cuáles.

---

## 2. Descargar el repositorio (una sola vez)

Clónalo en tu **carpeta personal** (así funcionan tal cual todos los comandos de los manuales):

```
cd ~
git clone https://github.com/AlbertLpz9211/Seguridad-en-desarrollo.git
cd Seguridad-en-desarrollo
```

> En Windows PowerShell `cd ~` también funciona y te lleva a `C:\Users\<tu-usuario>`.

**Lo que debes ver** al final:

```
Cloning into 'Seguridad-en-desarrollo'...
...
Resolving deltas: 100% ...
```

Comprueba que estás en el lugar correcto:

```
git status
```

Debe decir `On branch main` y `nothing to commit, working tree clean`.

---

## 3. Actualizar el repositorio (antes de cada clase)

El docente agrega los laboratorios nuevos antes de cada sesión. Para recibirlos:

```
cd ~/Seguridad-en-desarrollo
git pull
```

**Si `git pull` se niega** con un mensaje como `Your local changes to the following files would be
overwritten by merge`, es porque modificaste archivos de un laboratorio (por ejemplo, al resolver
un ejercicio). Guarda tus cambios aparte y vuelve a intentar:

```
git stash
git pull
git stash pop
```

`git stash` guarda tus cambios, `git pull` descarga lo nuevo y `git stash pop` vuelve a poner tus
cambios encima. Si aparece la palabra `CONFLICT`, avisa al docente antes de borrar nada.

> **Tu trabajo no se sube a este repositorio.** Tú no tienes permiso de escritura aquí. Lo que
> entregas va en **tu propio repositorio del proyecto integrador** y en la plataforma, como indica
> el manual de cada sesión.

---

## 4. Cómo están organizadas las carpetas

```
Seguridad-en-desarrollo/
├── README.md                     ← este archivo
└── s04-validacion-entradas/      ← un laboratorio por sesión
    ├── README.md                 ← manual de la sesión: LÉELO COMPLETO antes de empezar
    ├── go.mod, go.sum            ← definición del módulo y sus dependencias
    ├── vendor/                   ← dependencias ya descargadas (funciona sin internet)
    └── ...                       ← programas y pruebas de la sesión
```

Cada carpeta de sesión es un **módulo de Go independiente**. Todos los comandos de una sesión se
ejecutan **dentro de su carpeta** (por ejemplo, después de `cd s04-validacion-entradas`).

---

## 5. Problemas frecuentes

| Mensaje que ves | Qué significa | Qué hacer |
| --- | --- | --- |
| `go: command not found` / `go : El término 'go' no se reconoce` | Go no está instalado o la terminal no lo encuentra | Instala Go (1.2), **cierra y abre la terminal** |
| `go: go.mod requires go >= 1.26 (running go 1.25...)` | Tu Go es más viejo que el que pide el laboratorio | Actualiza Go a 1.26 o más reciente (1.2) |
| `go: downloading go1.26...` y luego un error de red | Tu Go intenta descargar una versión nueva y no hay internet | Instala Go 1.26 directamente (1.2) |
| `directory prefix . does not contain main module` o `go.mod file not found` | Estás en la carpeta equivocada | Entra a la carpeta de la sesión: `cd s04-validacion-entradas` |
| `fatal: not a git repository` | No estás dentro del repositorio | `cd ~/Seguridad-en-desarrollo` |
| `inconsistent vendoring` | Modificaste `go.mod` o `vendor/` | Descarta ese cambio: `git checkout -- go.mod go.sum vendor` |
