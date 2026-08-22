# 🎵 Asylum

Терминальный плеер для SoundCloud с красивым ASCII-артом Gengar, системой избранного и автоматическим переключением треков.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-blue?style=for-the-badge)

## ✨ Возможности

- 🔍 Поиск треков и плейлистов на SoundCloud
- 👻 Красивый ASCII-арт Gengar вместо обложек
- ⭐ Система избранных треков (сохраняется между сессиями)
- ⏭️ Автоматическое переключение треков
- 🎛️ Управление громкостью и перемоткой
- 🔄 Поддержка SOCKS5/HTTP прокси (для регионов с блокировками)
- 💾 Локальное кэширование треков

## 📦 Установка

### Зависимости

Для работы требуются `yt-dlp` и `ffmpeg`:

**Arch Linux / CachyOS:**
```bash
sudo pacman -S yt-dlp ffmpeg
**Ubuntu / Debian:**
