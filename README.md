# 📖 txtreader

**txtreader** es un lector de archivos de texto en la terminal, construido en **Go** utilizando la librería [Bubbletea](https://github.com/charmbracelet/bubbletea) y [Lipgloss](https://github.com/charmbracelet/lipgloss).  
Su objetivo es proporcionar una experiencia interactiva para leer, tomar notas, gestionar vocabulario y mantener el progreso de lectura guardado automáticamente.

---

## 🚀 Funcionalidades

### Navegación de Texto
- Muestra el archivo de texto en pantalla con resaltado de la línea actual.
- Posicionamiento centrado en torno a la línea que se está leyendo.
- **Destacado de palabras individuales** dentro de la línea para facilitar estudio y vocabulario.
- Atajos de teclado:
  - `j` → Mover hacia abajo (siguiente línea).
  - `k` → Mover hacia arriba (línea anterior).
  - `→ / ←` → Mover palabra seleccionada dentro de la línea actual.
  - `g` → Ir a un número de línea específico (abre un cuadro de diálogo).
  - `q` o `Ctrl+C` → Salir del programa.

### Vocabulario
- Permite agregar palabras seleccionadas al **vocabulario personal** presionando `w`.
- Copiar palabra seleccionada al portapapeles con `c`.
- Navegar entre palabras guardadas:
  - `j` / `k` → Moverse por la lista de vocabulario.
- Eliminar palabra seleccionada con `d`.

### Notas
- Creación de **notas rápidas y multilinea** con `n`.
- Guardar nota actual con `Ctrl+S`.
- Cancelar edición con `Ctrl+C`.
- Navegar entre notas guardadas con `j` / `k`.
- Eliminar una nota seleccionada con `d`.
  - **Con confirmación opcional** si la variable de entorno `CONFIRM_NOTES_DELETE=true` está activa.
  - Si no está activa, la nota se elimina inmediatamente sin preguntar.

### Enlaces Rápidos
- Con la tecla `o` se abre un cuadro de selección de enlaces a:
  - **GoodReads**.
  - **Real Academia Española (RAE)**.
- Navegación con `k/j` y confirmación con `Enter`.

### Guardado de Progreso
- El progreso de lectura se guarda manualmente con la tecla `s`:
  - Línea actual.
  - Lista de vocabulario.
  - Notas guardadas.
- Se almacena en `~/.ltbr/progress.json` (en el directorio del usuario).
- El archivo JSON persiste múltiples archivos de lectura (claveada por hash MD5 de la ruta).

---

## 📂 Estructura del Proyecto