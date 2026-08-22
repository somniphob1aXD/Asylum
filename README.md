# 🎵 Asylum

Терминальный плеер для SoundCloud с красивым ASCII-артом Gengar, системой избранного и автоматическим переключением треков.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-blue?style=for-the-badge)

## ✨ Возможности

- 🔍 Поиск треков и плейлистов на SoundCloud
- 👻 Красивый ASCII-арт Gengar вместо обложек
- ⭐ Система избранных треков (сохраняется между сессиями)
- ⏭️ Автоматическое переключение треков
- 🎛️ Управление громкостью и перемоткой
- 🔄 Поддержка SOCKS5/HTTP прокси (для регионов с блокировками)
- 💾 Локальное кэширование треков
- 🪟 Поддержка Windows (экспериментально)

## 📦 Установка

### Зависимости

Для работы требуются `yt-dlp` и `ffmpeg`:

**Arch Linux / CachyOS:**

    sudo pacman -S yt-dlp ffmpeg

**Ubuntu / Debian:**

    sudo apt install yt-dlp ffmpeg

**macOS:**

    brew install yt-dlp ffmpeg

**Windows (через Scoop):**

    scoop install yt-dlp ffmpeg

**Windows (через Chocolatey):**

    choco install yt-dlp ffmpeg

### Компиляция из исходников

**Linux / macOS:**

    # Клонируйте репозиторий
    git clone https://github.com/somniphob1aXD/asylum.git
    cd asylum

    # Установите Go зависимости
    go mod download

    # Скомпилируйте
    go build -o asylum main.go

    # Установите в систему (опционально)
    sudo install -m 755 asylum /usr/local/bin/asylum

**Windows (PowerShell):**

    # Клонируйте репозиторий
    git clone https://github.com/somniphob1aXD/asylum.git
    cd asylum

    # Установите Go зависимости
    go mod download

    # Скомпилируйте
    go build -o asylum.exe main.go

## 🚀 Использование

### Базовый запуск

**Linux / macOS:**

    # Если SoundCloud доступен в вашем регионе
    asylum

    # Если SoundCloud заблокирован (РФ, СНГ)
    export SC_PROXY="socks5://127.0.0.1:1080"
    asylum

**Windows (PowerShell):**

    # Если SoundCloud доступен в вашем регионе
    .\asylum.exe

    # Если SoundCloud заблокирован (РФ, СНГ)
    $env:SC_PROXY="socks5://127.0.0.1:1080"
    .\asylum.exe

### Настройка прокси

**Linux / macOS** — добавьте в `~/.bashrc` или `~/.zshrc`:

    export SC_PROXY="socks5://127.0.0.1:1080"

**Windows** — добавьте в профиль PowerShell (`$PROFILE`):

    $env:SC_PROXY="socks5://127.0.0.1:1080"

## ⚠️ Windows (экспериментально)

Windows-поддержка работает, но с ограничениями:

- ✅ Все функции плеера работают
- ✅ Прокси (SOCKS5/HTTP) поддерживается
- ⚠️ **Braille-арт Gengar отображается только в Windows Terminal** (не в cmd.exe)
- ⚠️ Требуется шрифт с поддержкой braille-символов: Cascadia Code, Fira Code или JetBrains Mono

### Требования для Windows

1. Установите [Windows Terminal](https://aka.ms/terminal)
2. Установите шрифт [Cascadia Code](https://github.com/microsoft/cascadia-code)
3. Установите зависимости через Scoop или Chocolatey (см. выше)
4. Скомпилируйте и запустите в Windows Terminal

## ️ Управление

| Клавиша | Действие |
|---------|----------|
| `/` | Поиск треков |
| `F` | Режим избранного |
| `f` | Добавить/убрать трек из избранного |
| `j` / `k` | Навигация вверх / вниз |
| `Enter` | Воспроизвести выбранный трек |
| `Space` / `p` | Пауза / продолжить |
| `h` / `l` | Перемотка назад / вперёд (±5 сек) |
| `n` / `N` | Следующий / предыдущий трек |
| `+` / `-` | Увеличить / уменьшить громкость |
| `Esc` | Выйти из режима избранного / закрыть поиск |
| `q` / `Ctrl+C` | Выход из программы |

## 📂 Структура проекта

    asylum/
    ├── main.go              # Основной код приложения
    ├── gengar.txt           # ASCII-арт Gengar (встраивается в бинарник)
    ├── go.mod               # Go модуль
    ├── go.sum               # Зависимости Go
    ├── README.md            # Документация
    └── .gitignore           # Игнорируемые файлы

## 💾 Хранение данных

Все данные сохраняются в папке `.sc-tui-cache/` в директории запуска:

- `favorites.json` — список избранных треков
- `track-*.mp3` — временно скачанные треки (удаляются после воспроизведения)

## 🎨 ASCII-арт

Хотите изменить Gengar на другого персонажа? Просто отредактируйте файл `gengar.txt` и пересоберите:

    nano gengar.txt  # Вставьте новый ASCII-арт
    go build -o asylum main.go
    sudo install -m 755 asylum /usr/local/bin/asylum

## ️ Разработка

    # Запуск в режиме разработки
    go run main.go

    # Сборка с отладочной информацией
    go build -gcflags="all=-N -l" -o asylum main.go

    # Запуск тестов (если будут)
    go test ./...

## 🤝 Вклад

Pull Request'ы приветствуются! Если вы нашли баг или хотите предложить новую функцию — создайте Issue.

## 📄 Лицензия

MIT License — см. файл [LICENSE](LICENSE) для деталей.

## 🙏 Благодарности

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI фреймворк
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) — загрузка аудио
- [beep](https://github.com/faiface/beep) — аудиоплеер для Go
- [go-runewidth](https://github.com/mattn/go-runewidth) — корректное отображение braille-символов

---

**Made by ancientreligionbtw #lowtier**
