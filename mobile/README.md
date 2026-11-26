# CompostIOT Mobile (Android / iOS — Teléfono y Tablet)

Aplicación móvil híbrida construida con **Capacitor 7** sobre el mismo frontend responsive de `web/`, adaptada a:
- **Teléfonos** (navegación inferior fija, selector deslizable de pilas, áreas seguras para notch/barra de gestos).
- **Tablets** (aprovechamiento de ancho en 2 y 3 columnas, rotación libre vertical/horizontal).
- **Entornos locales y nube**: soporta múltiples servidores (`http://192.168.x.x:8080` en red local SENA/finca y servidores HTTPS en la nube) y caché offline para consultar lecturas en campo sin señal.

## Sincronizar cambios del frontend web

Cada vez que actualices la interfaz en `web/`, ejecuta:

```bash
cd mobile
pnpm run sync:android
```

## Generar APK de Android

1. Abre el proyecto nativo en Android Studio:
   ```bash
   cd mobile
   pnpm run open:android
   ```
   O compila desde terminal si tienes Android SDK / JDK 21 instalado:
   ```bash
   cd mobile/android
   ./gradlew assembleDebug
   ```
   El APK quedará en `mobile/android/app/build/outputs/apk/debug/app-debug.apk`.

## Compilar para iOS (en macOS con Xcode)

```bash
cd mobile
pnpm exec cap add ios
pnpm run sync:ios
pnpm run open:ios
```

## Instalación rápida sin compilar APK (PWA)

También puedes abrir la aplicación web desde Chrome (Android) o Safari (iOS/iPadOS) y pulsar **"Agregar a la pantalla de inicio"**: gracias a `web/public/manifest.webmanifest`, se instala en pantalla completa como aplicación móvil.
