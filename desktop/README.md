# CompostIOT Desktop — Cliente Nativo Go (Wails)

Cliente nativo multiplataforma (**Windows, Linux y macOS**) construido en **Go + Wails v2**, que reutiliza el frontend de `web/` y añade:

- **Soporte multi-servidor**: guarda múltiples perfiles (ej. `http://localhost:8080` para aula sin internet + instancia en la nube).
- **Llavero nativo del sistema operativo (`go-keyring`)**: el token Bearer JWT devuelto por `/api/v1/auth/login` se almacena de forma segura en **Windows Credential Manager**, **macOS Keychain** o **Linux Secret Service** (GNOME Keyring / KWallet) y se adjunta en cada petición REST.
- **Caché offline en SQLite**: sincroniza pilas y lecturas en la carpeta de configuración del usuario (`~/.config/CompostIOT/cache.db` en Linux, `%AppData%\CompostIOT\cache.db` en Windows, `~/Library/Application Support/CompostIOT/cache.db` en macOS). Si se pierde conexión, permite consultar sesiones, pilas, gráficas históricas e incluso exportar CSV desde el caché local.

## Pruebas automatizadas

```bash
cd desktop
go test -v ./internal/...
```

## Compilar aplicación nativa

### 1. Sincronizar el frontend compilado
```bash
cd desktop
bash sync-frontend.sh
```

### 2. Compilar binario nativo

- **Con Wails CLI** (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`):
  - **Linux (WebKitGTK 4.1)**: `wails build -tags webkit2_41`
  - **Windows**: `wails build -platform windows/amd64`
  - **macOS (Universal)**: `wails build -platform darwin/universal`
- **Directo con Go** (en Linux con `gtk3` y `webkit2gtk-4.1` instalados):
  ```bash
  go build -tags "desktop,production,webkit2_41" -o build/bin/CompostIOT .
  ```
